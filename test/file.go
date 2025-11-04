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
func SetupTestFs(t *testing.T) {
	oldFs := file.GetFs()

	testFs := afero.NewMemMapFs()
	file.SetFs(testFs)

	t.Cleanup(func() {
		file.SetFs(oldFs)
	})
}

// MakeDir creates the given directory
func MakeDir(t *testing.T, dir string) {
	err := file.GetFs().MkdirAll(dir, 0755)

	if err != nil {
		t.Fatalf("cannot make directory '%s': %v", dir, err)
	}
}

func RemoveDir(t *testing.T, dir string) {
	err := file.GetFs().RemoveAll(dir)

	if err != nil {
		t.Fatalf("cannot remove directory '%s': %v", dir, err)
	}
}

// MakeFile creates a file with the given data.
func MakeFile(t *testing.T, path string, data string, perm os.FileMode) os.FileInfo {
	err := afero.WriteFile(file.GetFs(), path, []byte(data), perm)

	if err != nil {
		t.Fatalf("cannot make file '%s': %v", path, err)
	}

	info, err := file.GetFs().Stat(path)

	if err != nil {
		t.Fatalf("cannot stat file '%s': %v", path, err)
	}

	return info
}
