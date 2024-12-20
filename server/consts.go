package server

const VERSION = 136
const assetMaxSize = 50 * 1024 * 1024 // limit asset size to 50mb
const nodeTypesVersion = "22.9.0"

// asset extensions
var assetExts = map[string]bool{
	"node":       true,
	"wasm":       true,
	"less":       true,
	"sass":       true,
	"scss":       true,
	"stylus":     true,
	"styl":       true,
	"json":       true,
	"jsonc":      true,
	"csv":        true,
	"xml":        true,
	"plist":      true,
	"tmLanguage": true,
	"tmTheme":    true,
	"yml":        true,
	"yaml":       true,
	"txt":        true,
	"glsl":       true,
	"frag":       true,
	"vert":       true,
	"md":         true,
	"mdx":        true,
	"markdown":   true,
	"html":       true,
	"htm":        true,
	"svg":        true,
	"png":        true,
	"jpg":        true,
	"jpeg":       true,
	"webp":       true,
	"gif":        true,
	"ico":        true,
	"eot":        true,
	"ttf":        true,
	"otf":        true,
	"woff":       true,
	"woff2":      true,
	"m4a":        true,
	"mp3":        true,
	"m3a":        true,
	"ogg":        true,
	"oga":        true,
	"wav":        true,
	"weba":       true,
	"gz":         true,
	"tgz":        true,
}

// node built-in modules
var nodeBuiltinModules = map[string]bool{
	"assert":              true,
	"assert/strict":       true,
	"async_hooks":         true,
	"buffer":              true,
	"child_process":       true,
	"cluster":             true,
	"console":             true,
	"constants":           true,
	"crypto":              true,
	"dgram":               true,
	"diagnostics_channel": true,
	"dns":                 true,
	"dns/promises":        true,
	"domain":              true,
	"events":              true,
	"fs":                  true,
	"fs/promises":         true,
	"http":                true,
	"http2":               true,
	"https":               true,
	"inspector":           true,
	"inspector/promises":  true,
	"module":              true,
	"net":                 true,
	"os":                  true,
	"path":                true,
	"path/posix":          true,
	"path/win32":          true,
	"perf_hooks":          true,
	"process":             true,
	"punycode":            true,
	"querystring":         true,
	"readline":            true,
	"readline/promises":   true,
	"repl":                true,
	"stream":              true,
	"stream/consumers":    true,
	"stream/promises":     true,
	"stream/web":          true,
	"string_decoder":      true,
	"sys":                 true,
	"timers":              true,
	"timers/promises":     true,
	"tls":                 true,
	"trace_events":        true,
	"tty":                 true,
	"url":                 true,
	"util":                true,
	"util/types":          true,
	"v8":                  true,
	"vm":                  true,
	"wasi":                true,
	"worker_threads":      true,
	"zlib":                true,
}

// css packages
var cssPackages = map[string]string{
	"@unocss/reset":    "tailwind.css",
	"inter-ui":         "inter.css",
	"normalize.css":    "normalize.css",
	"modern-normalize": "modern-normalize.css",
	"reset-css":        "reset.css",
}
