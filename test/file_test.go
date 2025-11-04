package test

import (
	"testing"
)

func Test(t *testing.T) {
	SetupTestFs(t)

	MakeDir(t, "dir")
	MakeFile(t, "dir/test", "foo", 0644)

	RemoveDir(t, "dir")
}
