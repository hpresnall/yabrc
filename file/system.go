// Package file provides a wrapper around the file system to allow for testing.
package file

import "github.com/spf13/afero"

var fileSystem afero.Fs

func init() {
	fileSystem = afero.NewOsFs()
}

// GetFs gets the current filesystem used by the Index. Meant for testing.
func GetFs() afero.Fs {
	// Fs is an interface => ok to return obj and not pointer
	return fileSystem
}

// SetFs sets the current filesystem used by the Index. Meant for testing. Returns the previous filesystem.
// The behavior is undefined if this function is called between building an Index and adding Entries to it.
func SetFs(fs afero.Fs) afero.Fs {
	if fs == nil {
		panic("cannot use nil fs")
	}

	oldFs := fileSystem
	fileSystem = fs

	return oldFs
}
