package util

import (
	"testing"

	"github.com/hpresnall/yabrc/index"
)

func TestCompareEmpty(t *testing.T) {
	if !Compare(&index.Index{}, &index.Index{}) {
		t.Error("empty indexes should be equal")
	}
}

func TestCompareNil(t *testing.T) {
	if !Compare(nil, nil) {
		t.Error("two nil indexes should be equal")
	}

	if Compare(&index.Index{}, nil) {
		t.Error("index compared to nil should be false")
	}
}

func TestCompareDifferentRoots(t *testing.T) {
	idx1, _ := index.New("root1")
	idx2, _ := index.New("root2")

	if !Compare(idx1, idx2) {
		t.Error("indexes with different roots should be equal")
	}
}

func TestCompareSame(t *testing.T) {
	config, teardown := LoadTestConfig(t)
	defer teardown()

	idx := BuildTestIndex(t, config)

	if !Compare(idx, idx) {
		t.Error("index should equal itself")
	}
}

func TestCompareEqual(t *testing.T) {
	config, teardown := LoadTestConfig(t)
	defer teardown()

	idx1 := BuildTestIndex(t, config)
	idx2 := BuildTestIndex(t, config)

	if !Compare(idx1, idx2) {
		t.Error("indexes should be equal")
	}
}

func TestCompare(t *testing.T) {
	config, teardown := LoadTestConfig(t)
	defer teardown()

	idx1 := BuildTestIndex(t, config)
	idx2 := BuildTestIndex(t, config)

	// file paths must match BuildTestIndex()
	info1 := MakeFile(t, "testRoot/"+"test1/"+"test1_1", "1", 0644)               // smaller
	info2 := MakeFile(t, "testRoot/"+"test2/"+"test2_1", "data2_1 updated", 0644) // larger
	info2_1 := MakeFile(t, "testRoot/test2/sub1"+"test2_2", "data2_x", 0644)      // different hash
	info4 := MakeFile(t, "testRoot/"+"test4/"+"test4_1", "data4_1", 0644)         // missing
	info5 := MakeFile(t, "testRoot/"+"test5/"+"test5_1", "data5_1", 0644)         // new

	idx1.Add("testRoot/"+"test1/"+"test1_1", info1)
	idx1.Add("testRoot/"+"test2/"+"test2_1", info2)
	idx1.Add("testRoot/"+"test2/sub1"+"test2_2", info2_1)
	idx1.Add("testRoot/"+"test4/"+"test4_1", info4)

	idx2.Add("testRoot/"+"test5/"+"test5_1", info5)

	// track changes; ensure everything is removed
	comparisons := make(map[string]struct{})
	comparisons["test1/"+"test1_1"] = struct{}{}
	comparisons["test2/"+"test2_1"] = struct{}{}
	comparisons["test4/"+"test4_1"] = struct{}{}
	comparisons["test5/"+"test5_1"] = struct{}{}

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

	if Compare(idx1, idx2) {
		t.Error("indexes should not be equal")
	}

	if len(comparisons) != 0 {
		t.Error("not all cases were detected", comparisons)
	}
}
