package index

import (
	"testing"
	"time"

	"github.com/spf13/afero"

	"github.com/hpresnall/yabrc/file"
)

func TestNewNilConfig(t *testing.T) {
	_, err := New(nil)

	if err == nil {
		t.Fatal("should not be able to create Index with a nil Config")
	}
}

func TestNew(t *testing.T) {
	idx := ForTest(t)

	// if !strings.HasSuffix(idx.Root(), "/") {
	// 	t.Error("root should end in '/'")
	// }

	if idx.rootLen != (len(idx.Config().Root()) + 1) {
		t.Error("rootLen should be ", len("test/"), "not", idx.rootLen)
	}

	if idx.Timestamp().IsZero() {
		t.Error("timestamp should not be the zero value")
	}

	if idx.Size() != 0 {
		t.Error("data should be empty")
	}
}

func TestAdd(t *testing.T) {
	idx := ForTest(t)
	testFs := file.GetFs()
	root := idx.Config().Root()

	err := testFs.MkdirAll(root, 0755)

	if err != nil {
		t.Fatal("cannot create test dir", err)
	}

	afero.WriteFile(testFs, root+"/test", []byte("test"), 0644)

	if err != nil {
		t.Fatal("cannot make file", root+"/test", err)
	}

	info, err := testFs.Stat(root + "/test")

	if err != nil {
		t.Fatal("cannot load FileInfo", err)
	}

	err = idx.Add(root+"/test", info)

	if err != nil {
		t.Error("cannot add Entry", err)
	}

	if idx.Size() == 0 {
		t.Error("data should not be empty")
	}

	e, exists := idx.Get("test")

	if !exists {
		t.Error("did not add Entry to Index")
	}

	if e.Path() != "test" {
		t.Errorf("entry has path '%s', not '%s'", e.Path(), "test")
	}

	// test 0 byte file
	afero.WriteFile(testFs, root+"/testzero", []byte(""), 0644)

	if err != nil {
		t.Fatal("cannot make file", root+"/testzero", err)
	}

	info, err = testFs.Stat(root + "/testzero")

	if err != nil {
		t.Fatal("cannot load FileInfo", err)
	}

	err = idx.Add(root+"/testzero", info)

	if err != nil {
		t.Error("cannot add Entry", err)
	}

	if idx.Size() != 1 {
		t.Error("data should have 1 Entry")
	}
}

func TestAddBadPath(t *testing.T) {
	idx := ForTest(t)

	// assume buildEntry errors with bad path / info
	err := idx.Add("", nil)

	if err == nil {
		t.Error("should not be able to add a path that does not start with Index root")
	}

	err = idx.Add(idx.Config().Root()+"/test", nil)

	if err == nil {
		t.Error("should not be able to add with a nil FileInfo")
	}
}

func TestAddEntry(t *testing.T) {
	idx := ForTest(t)

	// valid entry
	e := Entry{path: idx.Config().Root() + "/test", lastMod: time.Now(), size: 1, hash: "n4bQgYhMfWWaL+qgxVrQFaO/TxsrC4Is0V1sFbDwCgg"}
	err := idx.AddEntry(e)

	if err != nil {
		t.Error("should be able to add a valid entry", err)
	}

	// ensure more than one entry is successful
	e = Entry{path: idx.Config().Root() + "/test2", lastMod: time.Now(), size: 1, hash: "n4bQgYhMfWWaL+qgxVrQFaO/TxsrC4Is0V1sFbDwCgg"}
	err = idx.AddEntry(e)

	if err != nil {
		t.Error("should be able to add a valid entry", err)
	}

	// should overwrite existing entry
	err = idx.AddEntry(e)

	if err != nil {
		t.Error("should be able to add a valid entry", err)
	}

	if _, exists := idx.Get("test"); !exists {
		t.Error("Index should have an entry for 'test'")
	}

	// test path when it does not start with Index.Root()
	e.path = "test"
	delete(idx.data, "test")

	err = idx.AddEntry(e)

	if err != nil {
		t.Error("should be able to add a valid entry", err)
	}

	if _, exists := idx.Get("test"); !exists {
		t.Error("Index should have an entry for 'test'")
	}

	if idx.Size() != 2 {
		t.Error("data should contain 2 entries, not", idx.Size())
	}

	// coverage for String()
	s := idx.String()

	if s == "" {
		t.Error("should return string")
	}

	sJSON := idx.StringWithEntries()

	if s == "" {
		t.Error("should return string")
	}

	if len(s) > len(sJSON) {
		t.Error("JSON string should be larger than default String()")
	}

	// invalid entry
	e = Entry{path: idx.Config().Root() + "/test"}
	err = idx.AddEntry(e)

	if err == nil {
		t.Error("should not be able to add invalid entry")
	}
}

func TestGetNonExistentEntry(t *testing.T) {
	idx := ForTest(t)

	_, exists := idx.Get("nonexistent")

	if exists {
		t.Error("should not have nonexistent entry")
	}
}

func TestForEach(t *testing.T) {
	idx := ForTest(t)

	// valid entry
	e := Entry{path: idx.Config().Root() + "/test", lastMod: time.Now(), size: 1, hash: "n4bQgYhMfWWaL+qgxVrQFaO/TxsrC4Is0V1sFbDwCgg"}
	err := idx.AddEntry(e)

	if err != nil {
		t.Error("should be able to add a valid entry", err)
	}

	if idx.Size() == 0 {
		t.Error("data should not be empty")
	}

	count := 0

	idx.ForEach(func(e Entry) {
		count++
	})

	if count != 1 {
		t.Error("should have iterated over 1 entry")
	}
}
