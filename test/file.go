//go:build !release
// +build !release

package test

import (
	"os"
	"testing"

	"github.com/spf13/afero"

	"github.com/hpresnall/yabrc/file"
)

// SetupTestFs creates a filesystem for testing.
// The returned function must be called / deferred in tests.
func SetupTestFs() func() {
	oldFs := file.GetFs()

	testFs := afero.NewMemMapFs()
	file.SetFs(testFs)

	return func() {
		file.SetFs(oldFs)
	}
}

// MakeDir creates the given directory
func MakeDir(t *testing.T, dir string) {
	err := file.GetFs().MkdirAll("testRoot", 0755)

	if err != nil {
		t.Fatal("cannot make directory", dir, err)
	}
}

// MakeFile creates a file with the given data.
func MakeFile(t *testing.T, path string, data string, perm os.FileMode) os.FileInfo {
	err := afero.WriteFile(file.GetFs(), path, []byte(data), perm)

	if err != nil {
		t.Fatal("cannot make file", path, err)
	}

	info, err := file.GetFs().Stat(path)

	if err != nil {
		t.Fatal("cannot stat file", path, err)
	}

	return info
}
