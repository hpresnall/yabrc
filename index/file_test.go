package index

import (
	"reflect"
	"testing"
	"time"

	"github.com/spf13/afero"

	"github.com/hpresnall/yabrc/config"
	"github.com/hpresnall/yabrc/file"
)

func TestLoadMissingIndex(t *testing.T) {
	config, teardown := config.ForTest(t)
	defer teardown()

	_, err := Load(config, "_test")

	if err == nil {
		t.Error("should not be able to load index from a missing file")
	}
}

func TestStoreBadDir(t *testing.T) {
	config, teardown := config.ForTest(t)
	defer teardown()

	idx, err := New(config.Root())

	if err != nil {
		t.Fatal("should be able to create index")
	}

	// use read only fs to simulate unwriatable dir
	// teardown will set back the original
	file.SetFs(afero.NewReadOnlyFs(file.GetFs()))

	err = Store(idx, config, "_test")

	if err == nil {
		t.Error("should not be able to store index in missing dir")
	}
}

func TestLoadAndStoreIndex(t *testing.T) {
	config, teardown := config.ForTest(t)
	defer teardown()

	// FIXME
	idx, err := New("test")

	if err != nil {
		t.Fatal("should be able to create index", err)
	}

	e := Entry{path: "testdir/test", lastMod: time.Now(), size: 1, hash: "n4bQgYhMfWWaL+qgxVrQFaO/TxsrC4Is0V1sFbDwCgg"}
	err = idx.AddEntry(e)

	if err != nil {
		t.Fatal("should be able to add entry", err)
	}

	err = Store(idx, config, "_test")

	if err != nil {
		t.Fatal("should be able to store index", err)
	}

	idx2, err := Load(config, "_test")

	if err != nil {
		t.Fatal("should be able to load index", err)
	}

	if idx.root != idx2.root {
		t.Error("root not the same")
	}

	if idx.rootLen != idx2.rootLen {
		t.Error("rootLen not the same")
	}

	if idx.timestamp != idx2.timestamp {
		t.Error("timestamp not the same")
	}

	if reflect.DeepEqual(idx, idx2) {
		t.Error("data not the same")
	}
}
