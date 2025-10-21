package test

import (
	"testing"
)

func TestFile(t *testing.T) {
	teardown := SetupTestFs()
	defer teardown()

	MakeDir(t, "dir")
	MakeFile(t, "dir/test", "foo", 0644)
}
