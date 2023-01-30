package server

import (
	"encoding/base64"
	"errors"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/ije/esbuild-internal/js_ast"
	"github.com/ije/esbuild-internal/js_parser"
	"github.com/ije/esbuild-internal/logger"
	"github.com/ije/gox/utils"
	"github.com/ije/gox/valid"
)

var (
	regexpFullVersion      = regexp.MustCompile(`^\d+\.\d+\.\d+[\w\.\+\-]*$`)
	regexpFullVersionPath  = regexp.MustCompile(`([^/])@\d+\.\d+\.\d+[\w\.\+\-]*(/|$)`)
	regexpBuildVersionPath = regexp.MustCompile(`^/v\d+(/|$)`)
	regexpLocPath          = regexp.MustCompile(`(\.js):\d+:\d+$`)
	regexpJSIdent          = regexp.MustCompile(`^[a-zA-Z_$][\w$]*$`)
	regexpAliasExport      = regexp.MustCompile(`^export\s*\*\s*from\s*['"](\.+/.+?)['"];?$`)
	npmNaming              = valid.Validator{valid.FromTo{'a', 'z'}, valid.FromTo{'0', '9'}, valid.Eq('_'), valid.Eq('.'), valid.Eq('-')}
)

type stringSet struct {
	lock sync.RWMutex
	m    map[string]struct{}
}

func newStringSet() *stringSet {
	return &stringSet{m: map[string]struct{}{}}
}

func (s *stringSet) Size() int {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return len(s.m)
}

func (s *stringSet) Has(key string) bool {
	s.lock.RLock()
	defer s.lock.RUnlock()

	_, ok := s.m[key]
	return ok
}

func (s *stringSet) Add(key string) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.m[key] = struct{}{}
}

func (s *stringSet) Remove(key string) {
	s.lock.Lock()
	defer s.lock.Unlock()

	delete(s.m, key)
}

func (s *stringSet) Reset() {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.m = map[string]struct{}{}
}

func (s *stringSet) Values() []string {
	s.lock.RLock()
	defer s.lock.RUnlock()

	a := make([]string, len(s.m))
	i := 0
	for key := range s.m {
		a[i] = key
		i++
	}
	return a
}

func splitPkgPath(pathname string) (pkgName string, submodule string) {
	a := strings.Split(strings.Trim(pathname, "/"), "/")
	pkgName = a[0]
	submodule = strings.Join(a[1:], "/")
	if strings.HasPrefix(pkgName, "@") && len(a) > 1 {
		pkgName = a[0] + "/" + a[1]
		submodule = strings.Join(a[2:], "/")
	}
	return
}

// ref https://github.com/npm/validate-npm-package-name
func validateNpmName(name string) bool {
	scope := ""
	nameWithoutScope := name
	if strings.HasPrefix(name, "@") {
		scope, nameWithoutScope = utils.SplitByFirstByte(name, '/')
		scope = scope[1:]
	}
	if (scope != "" && !npmNaming.Is(scope)) || (nameWithoutScope == "" || !npmNaming.Is(nameWithoutScope)) || len(name) > 214 {
		return false
	}
	return true
}

// isRemoteImport returns true if the import path is a remote URL.
func isRemoteImport(importPath string) bool {
	return strings.HasPrefix(importPath, "https://") || strings.HasPrefix(importPath, "http://")
}

// isLocalImport returns true if the import path is a local path.
func isLocalImport(importPath string) bool {
	return strings.HasPrefix(importPath, "file://") || strings.HasPrefix(importPath, "/") || strings.HasPrefix(importPath, "./") || strings.HasPrefix(importPath, "../") || importPath == "." || importPath == ".."
}

// include returns true if the given string is included in the given array.
func includes(a []string, s string) bool {
	for _, v := range a {
		if v == s {
			return true
		}
	}
	return false
}

func startsWith(s string, prefixs ...string) bool {
	for _, prefix := range prefixs {
		if strings.HasPrefix(s, prefix) {
			return true
		}
	}
	return false
}

func endsWith(s string, suffixs ...string) bool {
	for _, suffix := range suffixs {
		if strings.HasSuffix(s, suffix) {
			return true
		}
	}
	return false
}

type devFS struct {
	cwd string
}

func (fs devFS) ReadFile(name string) ([]byte, error) {
	return ioutil.ReadFile(path.Join(fs.cwd, name))
}

func dirExists(filepath string) bool {
	fi, err := os.Lstat(filepath)
	return err == nil && fi.IsDir()
}

func fileExists(filepath string) bool {
	fi, err := os.Lstat(filepath)
	return err == nil && !fi.IsDir()
}

func ensureDir(dir string) (err error) {
	_, err = os.Stat(dir)
	if err != nil && os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0755)
	}
	return
}

func clearDir(dir string) (err error) {
	os.RemoveAll(dir)
	err = os.MkdirAll(dir, 0755)
	return
}

func btoaUrl(s string) string {
	return strings.TrimRight(base64.URLEncoding.EncodeToString([]byte(s)), "=")
}

func atobUrl(s string) (string, error) {
	if l := len(s) % 4; l > 0 {
		s += strings.Repeat("=", 4-l)
	}
	data, err := base64.URLEncoding.DecodeString(s)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func kill(pidFile string) (err error) {
	if pidFile == "" {
		return
	}
	data, err := ioutil.ReadFile(pidFile)
	if err != nil {
		if os.IsNotExist(err) {
			err = nil
		}
		return
	}
	pid, err := strconv.Atoi(string(data))
	if err != nil {
		return
	}
	process, err := os.FindProcess(pid)
	if err != nil {
		return
	}
	return process.Kill()
}

func validateJS(filename string) (isESM bool, hasDefaultExport bool, err error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return
	}
	log := logger.NewDeferLog(logger.DeferLogNoVerboseOrDebug, nil)
	ast, pass := js_parser.Parse(log, logger.Source{
		Index:          0,
		KeyPath:        logger.Path{Text: "<stdin>"},
		PrettyPath:     "<stdin>",
		Contents:       string(data),
		IdentifierName: "stdin",
	}, js_parser.Options{})
	if !pass {
		err = errors.New("invalid syntax, require javascript/typescript")
		return
	}
	isESM = ast.ExportsKind == js_ast.ExportsESM
	_, hasDefaultExport = ast.NamedExports["default"]
	return
}
