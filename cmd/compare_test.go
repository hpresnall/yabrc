package cmd

import (
	"testing"

	"github.com/hpresnall/yabrc/config"
	"github.com/hpresnall/yabrc/test"
)

func TestCompareSelf(t *testing.T) {
	setup(t)

	if err := runCompare(nil, args); err != nil {
		t.Error("should not error on compare", err)
	}
}

func TestCompareSame(t *testing.T) {
	setup(t)

	idx.Store("_same")

	ext2 = "_same"

	if err := runCompare(nil, args); err != nil {
		t.Error("should not error on compare", err)
	}
}

func TestCompareDifferent(t *testing.T) {
	setup(t)

	// update index with a new file
	// use zzz to ensure sorted paths in compare find missing files last
	path := cfg.Root() + "/zzz"
	f := test.MakeFile(t, path, "zzz", 0644)
	idx.Add(path, f)

	// update file so hash is different
	path = cfg.Root() + "/test2/sub1/" + "test2_sub1_2"
	f = test.MakeFile(t, path, "data2_1_x", 0644) // different hash
	idx.Add(path, f)

	ext2 = "_different"

	idx.Store(ext2)

	err := runCompare(nil, args)

	if err == nil {
		t.Fatal("should error on compare when different", err)
	}
	if err.Error() != "" {
		t.Error("should error with empty Error when different")
	}
}

func TestCompareTwoConfigs(t *testing.T) {
	setup(t)

	idx.Store("_same")

	ext2 = "_same"

	if err := runCompare(nil, []string{config.TestFile, config.TestFile}); err != nil {
		t.Error("should not error on compare", err)
	}
}

func TestCompareBadConfig(t *testing.T) {
	setup(t)

	if err := runCompare(nil, []string{"invalid", config.TestFile}); err == nil {
		t.Error("should error on invalid config", err)
	}

	if err := runCompare(nil, []string{config.TestFile, "invalid"}); err == nil {
		t.Error("should error on invalid config", err)
	}
}
