package storage

import (
	"bytes"
	"io"
	"os"
	"path"
	"testing"

	"github.com/ije/gox/crypto/rand"
)

func TestMigrationStorage(t *testing.T) {
	root1 := path.Join(os.TempDir(), "storage_test_"+rand.Hex.String(8))
	back, err := NewFSStorage(&StorageOptions{Type: "fs", Endpoint: root1})
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(root1)

	root2 := path.Join(os.TempDir(), "storage_test_"+rand.Hex.String(8))
	front, err := NewFSStorage(&StorageOptions{Type: "fs", Endpoint: root2})
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(root2)

	migrationStorage := NewMigrationStorage(front, back)

	err = back.Put("test.txt", bytes.NewBufferString("Hello World!"))
	if err != nil {
		t.Fatal(err)
	}

	_, err = front.Stat("test.txt")
	if err != ErrNotFound {
		t.Fatal("Expected error, but got nil")
	}

	_, _, err = front.Get("test.txt")
	if err != ErrNotFound {
		t.Fatal("Expected error, but got nil")
	}

	fi, err := migrationStorage.Stat("test.txt")
	if err != nil {
		t.Fatal(err)
	}

	if fi.Size() != 12 {
		t.Fatalf("invalid file size(%d), shoud be 12", fi.Size())
	}

	f, fi, err := migrationStorage.Get("test.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	if fi.Size() != 12 {
		t.Fatalf("invalid file size(%d), shoud be 12", fi.Size())
	}

	data, err := io.ReadAll(f)
	if err != nil {
		t.Fatal(err)
	}

	if string(data) != "Hello World!" {
		t.Fatalf("invalid file content('%s'), shoud be 'Hello World!'", string(data))
	}

	fi, err = front.Stat("test.txt")
	if err != nil {
		t.Fatal(err)
	}

	if fi.Size() != 12 {
		t.Fatalf("invalid file size(%d), shoud be 12", fi.Size())
	}

	f, fi, err = front.Get("test.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	if fi.Size() != 12 {
		t.Fatalf("invalid file size(%d), shoud be 12", fi.Size())
	}

	data, err = io.ReadAll(f)
	if err != nil {
		t.Fatal(err)
	}

	if string(data) != "Hello World!" {
		t.Fatalf("invalid file content('%s'), shoud be 'Hello World!'", string(data))
	}
}
