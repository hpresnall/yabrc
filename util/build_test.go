package util

import (
	"testing"
	"time"

	log "github.com/spf13/jwalterweatherman"

	"github.com/hpresnall/yabrc/config"
	"github.com/hpresnall/yabrc/file"
	"github.com/hpresnall/yabrc/test"
)

func TestBuildIndexBadConfig(t *testing.T) {
	_, err := BuildIndex(&config.Config{}, nil)

	if err == nil {
		t.Error("should not be able to create an index with a nil config")
	}
}

func TestBuildEmptyIndex(t *testing.T) {
	config := config.ForTest(t)

	// root directory exists; sub directory with no files
	test.MakeDir(t, config.Root()+"/test1")

	idx, err := BuildIndex(&config, nil)

	if err != nil {
		t.Error("should be able to build an Index", err)
	}

	if idx.Size() != 0 {
		t.Error("Index should be empty")
	}
}

func TestBuildIndexMissingRoot(t *testing.T) {
	config := config.ForTest(t)

	test.RemoveDir(t, config.Root())

	idx, err := BuildIndex(&config, nil)

	if err == nil {
		t.Error("should not be able to build an Index", err)
	}

	if idx.Size() != 0 {
		t.Error("Index should be empty")
	}
}

func TestBuildIndex(t *testing.T) {
	config := config.ForTest(t)

	test.MakeFile(t, config.Root()+"/test2/sub1/"+"test2_sub1_2", "test", 0644)

	// increase code coverage
	log.SetLogThreshold(log.LevelTrace)

	idx, err := BuildIndex(&config, nil)

	if err != nil {
		t.Fatal("should not be able to build an Index", err)
	}

	// change times so file is re-hashed
	updated := time.Now().Add(time.Second * 5)
	file.GetFs().Chtimes(config.Root()+"/test2/sub1"+"test2_2", updated, updated)

	// test fast path branch with existing index; changing a single file
	test.MakeFile(t, config.Root()+"/test2/sub1/"+"test2_sub1_2", "data2_1_2 updated", 0644)

	newIdx, err := BuildIndex(idx.Config(), idx)

	if err != nil {
		t.Fatal("should be able to build an Index", err)
	}

	if idx.Size() != newIdx.Size() {
		t.Fatal("Index size should be the same", idx.Size(), newIdx.Size())
	}

	entry1, exists1 := idx.Get("test2/sub1/" + "test2_sub1_2")
	entry2, exists2 := newIdx.Get("test2/sub1/" + "test2_sub1_2")

	if !exists1 {
		t.Fatal("entry should exist in original Index")
	}
	if !exists2 {
		t.Fatal("entry should exist in new Index")
	}

	if entry1.LastMod() == entry2.LastMod() {
		t.Error("entries should have different lastMod times")
	}

	if entry1.Hash() == entry2.Hash() {
		t.Error("entries should not have the same hashes")
	}
}
