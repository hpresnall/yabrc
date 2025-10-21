package util

import (
	"testing"
	"time"

	"github.com/hpresnall/yabrc/config"
	"github.com/hpresnall/yabrc/file"
	"github.com/hpresnall/yabrc/test"

	log "github.com/spf13/jwalterweatherman"
)

func TestBuildIndexBadConfig(t *testing.T) {
	_, err := BuildIndex(config.Config{}, nil)

	if err == nil {
		t.Error("should not be able to create an index with a nil config")
	}
}

func TestBuildEmptyIndex(t *testing.T) {
	config, teardown := config.ForTest(t)
	defer teardown()

	// root directory exists; sub directory with no files
	test.MakeDir(t, "testRoot/test1")

	idx, err := BuildIndex(config, nil)

	if err != nil {
		t.Error("should be able to build an Index", err)
	}

	if idx.Size() != 0 {
		t.Error("Index should be empty")
	}
}

func TestBuildIndexMissingRoot(t *testing.T) {
	config, teardown := config.ForTest(t)
	defer teardown()

	idx, err := BuildIndex(config, nil)

	if err == nil {
		t.Error("should not be able to build an Index", err)
	}

	if idx.Size() != 0 {
		t.Error("Index should be empty")
	}
}

func TestBuildIndex(t *testing.T) {
	c, teardown := config.ForTest(t)
	defer teardown()

	idx := IndexForTest(t, c)

	// change times so one file is updated
	updated := time.Now().Add(time.Second * 5)
	file.GetFs().Chtimes("testRoot/test2/sub1"+"test2_2", updated, updated)

	// test fast path branch with existing index; changing a single file
	log.SetLogThreshold(log.LevelTrace)
	test.MakeFile(t, "testRoot/test2/sub1/"+"test2_sub1_2", "data2_1_2 updated", 0644)
	newIdx, err := BuildIndex(c, idx)

	if err != nil {
		t.Error("should be able to build an Index", err)
	}

	if newIdx.Size() != idx.Size() {
		t.Error("Index size should be the same", idx.Size(), newIdx.Size())
	}

	if Compare(idx, newIdx, false) {
		t.Error("fast path Index should different from previous")
	}
}
