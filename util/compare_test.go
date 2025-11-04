package util

import (
	"testing"

	"github.com/hpresnall/yabrc/config"
	"github.com/hpresnall/yabrc/index"
	"github.com/hpresnall/yabrc/test"
	log "github.com/spf13/jwalterweatherman"
)

func TestCompareEmpty(t *testing.T) {
	if !Compare(&index.Index{}, &index.Index{}, false) {
		t.Error("empty indexes should be equal")
	}
}

func TestCompareNil(t *testing.T) {
	if !Compare(nil, nil, false) {
		t.Error("two nil indexes should be equal")
	}

	if Compare(&index.Index{}, nil, false) {
		t.Error("index compared to nil should be false")
	}
}

func TestCompareDifferentRoots(t *testing.T) {
	config1, err := config.FromString(t, "root: testRoot1\nbaseName: testBaseName")

	if err != nil {
		t.Fatal("should be able to load config")
	}

	config2, err := config.FromString(t, "root: testRoot2\nbaseName: testBaseName")

	if err != nil {
		t.Fatal("should be able to load config")
	}

	idx1, _ := index.New(&config1)
	idx2, _ := index.New(&config2)

	if !Compare(idx1, idx2, false) {
		t.Error("indexes with different roots should be equal")
	}
}

func TestCompareSame(t *testing.T) {
	idx := IndexForTest(t)

	if !Compare(idx, idx, false) {
		t.Error("index should equal itself")
	}
}

func TestCompareEqual(t *testing.T) {
	idx1 := IndexForTest(t)
	idx2 := IndexForTest(t)

	if !Compare(idx1, idx2, false) {
		t.Error("indexes should be equal")
	}
}

func TestCompare(t *testing.T) {
	// increase code coverage
	log.SetLogThreshold(log.LevelTrace)

	idx1 := IndexForTest(t)
	root := idx1.Config().Root()

	// file paths must match BuildTestIndex()
	// changes in updated index
	test.MakeFile(t, root+"/test1/"+"test1_1", "1", 0644)                   // smaller
	test.MakeFile(t, root+"/test2/"+"test2_1", "data2_1 updated", 0644)     // larger
	test.MakeFile(t, root+"/test2/sub1/"+"test2_sub1_2", "data2_1_x", 0644) // different hash

	// missing in updated index
	test.RemoveDir(t, root+"/test3")

	// new file
	test.MakeFile(t, root+"/"+"test4/"+"test4_1", "data4_1", 0644)

	idx2, err := BuildIndex(idx1.Config(), idx1)

	if err != nil {
		t.Fatal("should be able to reload index")
	}

	// track changes; ensure everything is removed
	comparisons := make(map[string]struct{})
	comparisons["test1/"+"test1_1"] = struct{}{}
	comparisons["test2/"+"test2_1"] = struct{}{}
	comparisons["test4/"+"test4_1"] = struct{}{}

	oldMissing := OnMissing
	oldHash := OnHashChange

	OnMissing = func(missing index.Entry, other *index.Index) {
		delete(comparisons, missing.Path())
		oldMissing(missing, other)
	}

	OnHashChange = func(e1 index.Entry, e2 index.Entry) {
		delete(comparisons, e1.Path())
		oldHash(e1, e2)
	}

	defer func() {
		OnMissing = oldMissing
		OnHashChange = oldHash
	}()

	if Compare(idx1, idx2, false) {
		t.Error("indexes should not be equal")
	}

	if len(comparisons) != 0 {
		t.Error("not all comparison cases were detected", comparisons)
	}

	// run again with ignoreMissing; Compare should ignore test5 since it is only in idx2
	comparisons["test4/"+"test4_1"] = struct{}{}

	if Compare(idx1, idx2, true) {
		t.Error("indexes should not be equal when ignoreMissing is true")
	}
	if len(comparisons) != 1 {
		t.Error("ignoreMissing did not ignore missing file")
	}
}
