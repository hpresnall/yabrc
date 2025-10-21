package cmd

import (
	"testing"

	"github.com/hpresnall/yabrc/config"
	"github.com/hpresnall/yabrc/index"
	"github.com/hpresnall/yabrc/test"
)

func TestCompareSelf(t *testing.T) {
	defer setup(t)()

	if err := runCompare(nil, args); err != nil {
		t.Error("should not error on compare", err)
	}
}

func TestCompareSame(t *testing.T) {
	defer setup(t)()

	index.Store(idx, cfg, "_same")

	ext2 = "_same"

	if err := runCompare(nil, args); err != nil {
		t.Error("should not error on compare", err)
	}
}

func TestCompareDifferent(t *testing.T) {
	defer setup(t)()

	// update index with a new file
	path := cfg.Root() + "/another"
	f := test.MakeFile(t, path, "another", 0644)
	idx.Add(path, f)

	ext2 = "_different"

	index.Store(idx, cfg, ext2)

	err := runCompare(nil, args)

	if err == nil {
		t.Fatal("should error on compare when different", err)
	}
	if err.Error() != "" {
		t.Error("should error with empty Error when different")
	}
}

func TestCompareTwoConfigs(t *testing.T) {
	defer setup(t)()

	index.Store(idx, cfg, "_same")

	ext2 = "_same"

	if err := runCompare(nil, []string{config.TestFile, config.TestFile}); err != nil {
		t.Error("should not error on compare", err)
	}
}

func TestCompareBadConfig(t *testing.T) {
	defer setup(t)()

	if err := runCompare(nil, []string{"invalid", config.TestFile}); err == nil {
		t.Error("should error on invalid config", err)
	}

	if err := runCompare(nil, []string{config.TestFile, "invalid"}); err == nil {
		t.Error("should error on invalid config", err)
	}
}
