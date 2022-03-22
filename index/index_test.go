package index

import (
	"compress/gzip"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/spf13/afero"
	log "github.com/spf13/jwalterweatherman"
)

func TestNew(t *testing.T) {
	_, err := New("")

	if err == nil {
		t.Fatal("should not be able to create Index with empty root")
	}

	idx, err := New("test")

	if err != nil {
		t.Fatal("New failed", err)
	}

	if !strings.HasSuffix(idx.Root(), "/") {
		t.Error("root should end in '/'")
	}

	if idx.rootLen != len(idx.Root()) {
		t.Error("rootLen should be ", len("test/"), "not", idx.rootLen)
	}

	if idx.Timestamp().IsZero() {
		t.Error("timestamp should not be the zero value")
	}

	if idx.Size() != 0 {
		t.Error("data should be empty")
	}

	if idx.GetRelativePath("test/test") != "test" {
		t.Error("relative path should be test not", idx.GetRelativePath("test/test"))
	}

	if idx.GetRelativePath("test1") != "test1" {
		t.Error("relative path should be test1 not", idx.GetRelativePath("test1"))
	}

	idx, err = New(".")

	if err != nil {
		t.Fatal("New failed", err)
	}

	if idx.Root() != "./" {
		t.Error("root of '.' should be stored at './' not", idx.Root())
	}

	idx, err = New("/")

	if err != nil {
		t.Fatal("New failed", err)
	}

	if idx.Root() != "/" {
		t.Error("root of '/' should be stored at '/' not", idx.Root())
	}
}

func TestAdd(t *testing.T) {
	idx, err := New("testdir")

	if err != nil {
		t.Fatal("New failed", err)
	}

	testFs, teardown := setupTestFs()
	defer teardown()

	err = testFs.MkdirAll("testdir", 0755)

	if err != nil {
		t.Fatal("cannot create test dir", err)
	}

	afero.WriteFile(testFs, "testdir/test", []byte("test"), 0644)

	if err != nil {
		t.Fatal("cannot make file", "testdir/test", err)
	}

	info, err := testFs.Stat("testdir/test")

	if err != nil {
		t.Fatal("cannot load FileInfo", err)
	}

	err = idx.Add("testdir/test", info)

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
	afero.WriteFile(testFs, "testdir/testzero", []byte(""), 0644)

	if err != nil {
		t.Fatal("cannot make file", "testdir/testzero", err)
	}

	info, err = testFs.Stat("testdir/testzero")

	if err != nil {
		t.Fatal("cannot load FileInfo", err)
	}

	err = idx.Add("testdir/testzero", info)

	if err != nil {
		t.Error("cannot add Entry", err)
	}

	if idx.Size() != 1 {
		t.Error("data should have 1 Entry")
	}
}

func TestAddBadPath(t *testing.T) {
	idx, err := New("testdir")

	if err != nil {
		t.Fatal("New failed", err)
	}

	// assume buildEntry errors with bad path / info
	err = idx.Add("", nil)

	if err == nil {
		t.Error("should not be able to add a path that does not start with Index root")
	}

	err = idx.Add("testdir/test", nil)

	if err == nil {
		t.Error("should not be able to add with a nil FileInfo")
	}
}

func TestAddEntry(t *testing.T) {
	idx, err := New("testdir")

	if err != nil {
		t.Fatal("New failed", err)
	}

	// valid entry
	e := Entry{path: "testdir/test", lastMod: time.Now(), size: 1, hash: "n4bQgYhMfWWaL+qgxVrQFaO/TxsrC4Is0V1sFbDwCgg"}
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

	if idx.Size() != 1 {
		t.Error("data should contain one entry", idx.Size())
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
	e = Entry{path: "testdir/test"}
	err = idx.AddEntry(e)

	if err == nil {
		t.Error("should not be able to add invalid entry")
	}
}
func TestGetNonExistentEntry(t *testing.T) {
	idx, err := New("testdir")

	if err != nil {
		t.Fatal("New failed", err)
	}

	_, exists := idx.Get("nonexistent")

	if exists {
		t.Error("should not have nonexistent entry")
	}
}

func TestForEach(t *testing.T) {
	idx, err := New("testdir")

	if err != nil {
		t.Fatal("New failed", err)
	}

	// valid entry
	e := Entry{path: "testdir/test", lastMod: time.Now(), size: 1, hash: "n4bQgYhMfWWaL+qgxVrQFaO/TxsrC4Is0V1sFbDwCgg"}
	err = idx.AddEntry(e)

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

func TestStoreAndLoad(t *testing.T) {
	idx, err := New(".")

	if err != nil {
		t.Fatal("New failed", err)
	}

	// use path with comma and spaces to ensure Load / Store handles correctly
	e := Entry{path: "./test  , file ", lastMod: time.Now(), size: 1, hash: "n4bQgYhMfWWaL+qgxVrQFaO/TxsrC4Is0V1sFbDwCgg"}

	err = idx.AddEntry(e)

	// Index stores path without root
	e.path = "test  , file "

	if err != nil {
		t.Error("could not add Entry to Index", err)
	}

	_, teardown := setupTestFs()
	defer teardown()

	err = idx.Store("test")

	if err != nil {
		t.Error("should be able to store", err)
	}

	idx2, err := Load("test")

	if err != nil {
		t.Error("should be able to load", err)
	}

	if idx.Root() != idx2.Root() {
		t.Error("roots should match", idx.Root(), idx2.Root())
	}

	if idx.Timestamp() != idx2.Timestamp() {
		t.Error("timestamps should match", idx.Timestamp(), idx2.Timestamp())
	}

	if idx.Size() != idx2.Size() {
		t.Error("size should be 1", idx.Size(), idx2.Size())
	}

	e2, exists := idx.data["test  , file "]

	if !exists {
		t.Error("loaded Index should contain an entry for 'testfile'", idx.data, idx2.data)
	}

	if (e.Path() != e2.Path()) || (e.LastMod() != e2.LastMod()) || (e.Size() != e2.Size()) || (e.Hash() != e2.Hash()) {
		t.Error("loaded Entity should be identical to stored Entity", e, e2)
	}
}

func TestStoreOnBadFs(t *testing.T) {
	idx, err := New("testdir")

	if err != nil {
		t.Fatal("New failed", err)
	}

	_, teardown := setupReadOnlyTestFs()
	defer teardown()

	err = idx.Store("invalid")

	if err == nil {
		t.Error("should not be able to store on a bad filesystem")
	}
}

func TestLoadMissing(t *testing.T) {
	_, err := Load("missing")

	if err == nil {
		t.Error("should not be able to load a missing file")
	}
}

func TestLoadEmptyFile(t *testing.T) {
	_, err := newIndexFromString("", t)

	if err == nil {
		t.Error("should not be able to load an empty file")
	}
}

func TestLoadEmptyRoot(t *testing.T) {
	_, err := newIndexFromString(fmt.Sprintf(",%d", time.Now().Unix()), t)

	if err == nil {
		t.Error("should not be able to load Index with missing root")
	}

}

func TestLoadNoTimestamp(t *testing.T) {
	_, err := newIndexFromString("testdir,", t)

	if err == nil {
		t.Error("should not be able to load Index with missing timestamp", err)
	}
}

func TestLoadNoEntries(t *testing.T) {
	_, err := newIndexFromString(fmt.Sprintf("testdir,%d", time.Now().Unix()), t)

	if err != nil {
		t.Error("should be able to load Index with no Entries", err)
	}
}

func TestLoadBadEntries(t *testing.T) {
	// blank lines and missing fields should all be skipped
	// bad entries should also be skipped without errors
	// only the last entry is good
	data := `
testdir,1234

,
,,
,,,
,,,,
	,
path,,,
path,-1,,
path,1,-1,
path,1,1,
path,1,1,short
path,1,1,n4bQgYhMfWWaL+qgxVrQFaO/TxsrC4Is0V1sFbDwCgg`

	idx, err := newIndexFromString(data, t)

	if err != nil {
		t.Error("should be able to load Index with corrupt Entries", err)
	}

	if idx.Size() != 1 {
		t.Error("Index should have 1 Entry", idx.Size())
	}

	if idx.Root() != "testdir/" {
		t.Error("Index root should be 'testdir/' even with 0 entries, not", idx.Root())
	}
}

func TestStoreZeroVal(t *testing.T) {
	_, teardown := setupTestFs()
	defer teardown()

	idx := &Index{}
	err := idx.Store("zeroval")

	if err == nil {
		t.Error("should not be able to store a zero value Index")
	}
}

func TestSetIndexFs(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("should not be able to set the index fs to nil")
		}
	}()

	SetIndexFs(nil)
}

func newIndexFromString(indexString string, t *testing.T) (*Index, error) {
	testFs, teardown := setupTestFs()
	defer teardown()

	// Load expects a gzipped file so store the given string
	out, err := testFs.Create("test")

	if err != nil {
		t.Fatal("cannot create Index save file")
		return &Index{}, err
	}

	gz := gzip.NewWriter(out)

	_, err = gz.Write([]byte(indexString))

	if err != nil {
		t.Fatal("cannot gzip Index save file")
		return &Index{}, err
	}

	err = gz.Flush()

	if err != nil {
		t.Fatal("could not flush Index save file", err)
	}

	err = out.Close()

	if err != nil {
		t.Fatal("could not close Index save file", err)
	}

	log.DEBUG.Printf("loading index from '%s'\n", indexString)

	return Load("test")
}
