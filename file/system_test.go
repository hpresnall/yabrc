package file

import (
	"testing"

	"github.com/spf13/afero"
)

func TestSetIndexFs(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("should not be able to set the index fs to nil")
		}
	}()

	SetFs(nil)
}

func TestGetIndexFs(t *testing.T) {
	fs := afero.NewReadOnlyFs(GetFs())
	SetFs(fs)

	if GetFs() != fs {
		t.Fatal("get and set do not match")
	}
}
