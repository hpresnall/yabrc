//+build !release

package index

import (
	"github.com/spf13/afero"
)

// create a filesystem for testing
// the returned function must be called / deferred in tests
func setupTestFs() (afero.Fs, func()) {
	oldFs := GetIndexFs()

	testFs := afero.NewMemMapFs()
	SetIndexFs(testFs)

	return testFs, func() {
		SetIndexFs(oldFs)
	}
}

func setupReadOnlyTestFs() (afero.Fs, func()) {
	testFs, teardown := setupTestFs()
	rofs := afero.NewReadOnlyFs(testFs)
	SetIndexFs(rofs)
	// original teardown will reset index fs to original value
	return rofs, teardown
}
