package index

import (
	"os"
	"testing"
	"time"

	"github.com/spf13/afero"

	"github.com/hpresnall/yabrc/file"
	"github.com/hpresnall/yabrc/test"
)

func TestEntryFromFile(t *testing.T) {
	_, info, teardown := setupEntryFs(t)
	defer teardown()

	e, err := buildEntry("test", info)

	if err != nil {
		t.Fatal("cannot build entry", err)
	}

	if !e.IsValid() {
		t.Fatal("entry is not valid", e)
	}

	if e.LastMod() != info.ModTime() {
		t.Error("last mod should be", info.ModTime(), "not", e.LastMod())
	}

	if e.Size() != 4 {
		t.Error("size should be 4 not", e.Size())
	}

	if e.Hash() != "n4bQgYhMfWWaL+qgxVrQFaO/TxsrC4Is0V1sFbDwCgg" {
		t.Errorf("incorrect hash '%s' is not '%s'", e.Hash(), "n4bQgYhMfWWaL+qgxVrQFaO/TxsrC4Is0V1sFbDwCgg")
	}

	// for coverage
	e.AsCsv()
}

func TestEntryFromMissingFile(t *testing.T) {
	testFs, info, teardown := setupEntryFs(t)
	defer teardown()

	// missing file but valid info
	testFs.Remove("test")

	_, err := buildEntry("test", info)

	if err == nil {
		t.Error("should fail to build entry from missing file", err)
	}
}

func TestEntryFromEmptyPath(t *testing.T) {
	_, info, teardown := setupEntryFs(t)
	defer teardown()

	// path does not match info
	_, err := buildEntry("", info)

	if err == nil {
		t.Error("should fail to build entry with empty path", err)
	}
}

func TestEntryFromWrongPath(t *testing.T) {
	_, info, teardown := setupEntryFs(t)
	defer teardown()

	// path does not match info
	_, err := buildEntry("another", info)

	if err == nil {
		t.Error("should fail to build entry when path and info are not the same", err)
	}
}

func TestEntryFromNilInfo(t *testing.T) {
	_, err := buildEntry("test", nil)

	if err == nil {
		t.Error("should fail to build entry with nil info", err)
	}
}

func TestValidEntry(t *testing.T) {
	e := Entry{}

	if e.IsValid() {
		t.Fatal("zero value Entry should not be valid")
	}

	e.path = "valid"

	if e.IsValid() {
		t.Fatal("Entry with just a path should not be valid")
	}

	e.lastMod = time.Now()

	if e.IsValid() {
		t.Error("Entry without a size should not be valid")
	}

	e.size = 1

	if e.IsValid() {
		t.Error("Entry without a hash should not be valid")
	}

	e.hash = "hash"

	if e.IsValid() {
		t.Error("Entry with a short hash should not be valid")
	}

	e.hash = "n4bQgYhMfWWaL+qgxVrQFaO/TxsrC4Is0V1sFbDwCgg"

	if !e.IsValid() {
		t.Error("Entry should be valid")
	}
}

func setupEntryFs(t *testing.T) (afero.Fs, os.FileInfo, func()) {
	teardown := test.SetupTestFs()
	testFs := file.GetFs()

	err := afero.WriteFile(testFs, "test", []byte("test"), 0644)

	if err != nil {
		teardown()
		t.Fatal("cannot make file", "test", err)
	}

	info, err := testFs.Stat("test")

	if err != nil {
		teardown()
		t.Fatal("cannot load FileInfo", err)
	}

	return testFs, info, teardown
}
