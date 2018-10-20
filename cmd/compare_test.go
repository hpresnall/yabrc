package cmd

import (
	"testing"

	"github.com/hpresnall/yabrc/util"
)

func TestCompareSelf(t *testing.T) {
	defer setup(t)()

	if err := runCompare(nil, args); err != nil {
		t.Error("should not error on compare", err)
	}
}

func TestCompareSame(t *testing.T) {
	defer setup(t)()

	util.StoreIndex(idx, config, "_same")

	ext2 = "_same"

	if err := runCompare(nil, args); err != nil {
		t.Error("should not error on compare", err)
	}
}

func TestCompareDifferent(t *testing.T) {
	defer setup(t)()

	// update index with a new file
	path := config.Root() + "/another"
	f := util.MakeFile(t, path, "another", 0644)
	idx.Add(path, f)

	ext2 = "_different"

	util.StoreIndex(idx, config, ext2)

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

	util.StoreIndex(idx, config, "_same")

	ext2 = "_same"

	if err := runCompare(nil, []string{util.ConfigFile, util.ConfigFile}); err != nil {
		t.Error("should not error on compare", err)
	}
}
