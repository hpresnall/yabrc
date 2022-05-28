package util

import (
	"testing"

	"github.com/hpresnall/yabrc/index"
	"github.com/spf13/afero"
)

func TestLoadMissingConfig(t *testing.T) {
	_, err := LoadConfig("missing")

	if err == nil {
		t.Error("should not be able to load config from a missing file")
	}
}

func TestLoadTestConfig(t *testing.T) {
	_, teardown := LoadTestConfig(t)
	defer teardown()
}

func TestLoadMissingIndex(t *testing.T) {
	config, teardown := LoadTestConfig(t)
	defer teardown()

	_, err := LoadIndex(config, "_test")

	if err == nil {
		t.Error("should not be able to load index from a missing file")
	}
}

func TestStoreBadDir(t *testing.T) {
	config, teardown := LoadTestConfig(t)
	defer teardown()

	idx, err := index.New(config.Root())

	if err != nil {
		t.Error("should be able to create index")
	}

	// use read only fs to simulate unwriatable dir
	index.SetIndexFs(afero.NewReadOnlyFs(index.GetIndexFs()))

	err = StoreIndex(idx, config, "_test")

	if err == nil {
		t.Error("should not be able to store index in missing dir")
	}
}

func TestLoadAndStoreIndex(t *testing.T) {
	config, teardown := LoadTestConfig(t)
	defer teardown()

	idx := BuildTestIndex(t, config)

	err := StoreIndex(idx, config, "_test")

	if err != nil {
		t.Error("should be able to store index", err)
	}

	idx2, err := LoadIndex(config, "_test")

	if err != nil {
		t.Error("should be able to load index", err)
	}

	if !Compare(idx, idx2, false) {
		t.Error("indexes should be equal")
	}
}
