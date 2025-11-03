package index

import (
	"compress/gzip"
	"fmt"
	"path"
	"reflect"
	"testing"
	"time"

	"github.com/spf13/afero"
	log "github.com/spf13/jwalterweatherman"

	"github.com/hpresnall/yabrc/config"
	"github.com/hpresnall/yabrc/file"
)

func TestLoadMissingIndex(t *testing.T) {
	config := config.ForTest(t)

	_, err := Load(&config, "_test")

	if err == nil {
		t.Error("should not be able to load index from a missing file")
	}
}

func TestStoreOnBadFs(t *testing.T) {
	idx := ForTest(t)

	rofs := afero.NewReadOnlyFs(file.GetFs())
	file.SetFs(rofs) // ForTest already had a Cleanup handler to reset to originally fs

	err := idx.Store("invalid")

	if err == nil {
		t.Error("should not be able to store on a bad filesystem")
	}
}

func TestLoadAndStore(t *testing.T) {
	idx := ForTest(t)

	e := Entry{path: idx.Config().Root() + "/test", lastMod: time.Now(), size: 1, hash: "n4bQgYhMfWWaL+qgxVrQFaO/TxsrC4Is0V1sFbDwCgg"}
	err := idx.AddEntry(e)

	if err != nil {
		t.Fatal("should be able to add entry", err)
	}

	err = idx.Store("_test")

	if err != nil {
		t.Fatal("should be able to store index", err)
	}

	idx2, err := Load(idx.Config(), "_test")

	if err != nil {
		t.Fatal("should be able to load index", err)
	}

	if idx.Config().Root() != idx2.Config().Root() {
		t.Error("root not the same")
	}

	if idx.rootLen != idx2.rootLen {
		t.Error("rootLen not the same")
	}

	if idx.timestamp != idx2.timestamp {
		t.Error("timestamp not the same")
	}

	if reflect.DeepEqual(idx, idx2) {
		t.Error("data not the same")
	}
}

func TestStoreAndLoad(t *testing.T) {
	idx := ForTest(t)

	// use path with comma and spaces to ensure Load / Store handles correctly
	e := Entry{path: "test  , file ", lastMod: time.Now(), size: 1, hash: "n4bQgYhMfWWaL+qgxVrQFaO/TxsrC4Is0V1sFbDwCgg"}

	err := idx.AddEntry(e)

	if err != nil {
		t.Error("could not add Entry to Index", err)
	}

	// Index stores path without root
	e, exists := idx.data["test  , file "]

	if !exists {
		t.Error("loaded Index should contain an entry for 'testfile'", idx.data)
	}

	err = idx.Store("test")

	if err != nil {
		t.Error("should be able to store", err)
	}

	idx2, err := Load(idx.Config(), "test")

	if err != nil {
		t.Error("should be able to load", err)
	}

	if idx.Config().Root() != idx2.Config().Root() {
		t.Error("roots should match", idx.Config().Root(), idx2.Config().Root())
	}

	if idx.Timestamp() != idx2.Timestamp() {
		t.Error("timestamps should match", idx.Timestamp(), idx2.Timestamp())
	}

	if idx.Size() != idx2.Size() {
		t.Error("size should be 1", idx.Size(), idx2.Size())
	}

	e2, exists := idx.data["test  , file "]

	if !exists {
		t.Error("loaded Index should contain an entry for 'testfile'", idx2.data)
	}

	if (e.Path() != e2.Path()) || (e.LastMod() != e2.LastMod()) || (e.Size() != e2.Size()) || (e.Hash() != e2.Hash()) {
		t.Error("loaded Entity should be identical to stored Entity", e, e2)
	}
}

func TestLoadMissing(t *testing.T) {
	config := config.ForTest(t)
	_, err := Load(&config, "missing")

	if err == nil {
		t.Error("should not be able to load a missing file")
	}
}

func TestLoadEmptyFile(t *testing.T) {
	_, err := fromString(t, "")

	if err == nil {
		t.Error("should not be able to load an empty file")
	}
}

func TestLoadEmptyRoot(t *testing.T) {
	_, err := fromString(t, fmt.Sprintf(",%d", time.Now().Unix()))

	if err == nil {
		t.Error("should not be able to load Index with missing root")
	}

}

func TestLoadNoTimestamp(t *testing.T) {
	_, err := fromString(t, "testRoot,")

	if err == nil {
		t.Error("should not be able to load Index with missing timestamp", err)
	}
}

func TestLoadNoEntries(t *testing.T) {
	_, err := fromString(t, fmt.Sprintf("testRoot,%d", time.Now().Unix()))

	if err != nil {
		t.Error("should be able to load Index with no Entries", err)
	}
}

func TestLoadBadEntries(t *testing.T) {
	// blank lines and missing fields should all be skipped
	// bad entries should also be skipped without errors
	// only the last entry is good
	data := `
testRoot,1234

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

	idx, err := fromString(t, data)

	if err != nil {
		t.Error("should be able to load Index with corrupt Entries", err)
	}

	if idx.Size() != 1 {
		t.Error("Index should have 1 Entry", idx.Size())
	}

	if idx.Config().Root() != "testRoot" {
		t.Error("Index root should be 'testRoot' even with 0 entries, not", idx.Config().Root())
	}
}

func TestStoreEmpty(t *testing.T) {
	idx := ForTest(t)
	err := idx.Store("empty")

	if err == nil {
		t.Error("should not be able to store an empty Index")
	}
}

func fromString(t *testing.T, indexString string) (*Index, error) {
	config := config.ForTest(t)
	testFs := file.GetFs()

	// mimic Index.getFile()
	indexFile := path.Join(config.SavePath(), config.BaseName()+"test")

	// Load expects a gzipped file so store the given string
	out, err := testFs.Create(indexFile)

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

	return Load(&config, "test")
}
