package cmd

import (
	"bufio"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/hpresnall/yabrc/index"
	"github.com/hpresnall/yabrc/util"
)

var idxTime time.Time

func setupUpdate(t *testing.T) func() {
	teardown := setup(t)

	// set file time earlier for test validation
	idxTime = idx.Timestamp().Add(time.Second * -3600)
	index.GetIndexFs().Chtimes(util.GetIndexFile(config, "_current"), idxTime, idxTime)

	// update index with a new file
	path := config.Root() + "/another"
	f := util.MakeFile(t, path, "another", 0644)
	idx.Add(path, f)

	return teardown
}

func TestUpdate(t *testing.T) {
	defer setupUpdate(t)()

	// move and save
	reader = bufio.NewReader(strings.NewReader("y\ny\n"))

	overwrite = false
	autosave = false

	runAndValidate(t)
	currentUpdated(t)
	oldExists(t)

	allInputRead(t)
}

func TestUpdateOverwrite(t *testing.T) {
	defer setupUpdate(t)()

	// save only
	reader = bufio.NewReader(strings.NewReader("y\n"))

	overwrite = true
	autosave = false

	runAndValidate(t)
	currentUpdated(t)
	oldDoesNotExist(t)

	allInputRead(t)
}

func TestUpdateOverwriteAutosave(t *testing.T) {
	defer setupUpdate(t)()

	overwrite = true
	autosave = true

	runAndValidate(t)
	currentUpdated(t)
	oldDoesNotExist(t)
}

func TestUpdateFast(t *testing.T) {
	defer setupUpdate(t)()

	fast = true

	// cover 4th combo too
	overwrite = false
	autosave = true

	runAndValidate(t)
	currentUpdated(t)
	oldExists(t)
}

func TestUpdateNoInput(t *testing.T) {
	defer setupUpdate(t)()

	// EOF on confirm; should quit and not save
	reader = bufio.NewReader(strings.NewReader(""))

	runAndValidate(t)
	currentExists(t)
	oldDoesNotExist(t)

	allInputRead(t)
}

func TestUpdateNoMove(t *testing.T) {
	defer setupUpdate(t)()

	// no move, should quit and not save
	reader = bufio.NewReader(strings.NewReader("n\ny\n"))

	currentExists(t)
	oldDoesNotExist(t)

	if _, err := reader.ReadByte(); err != nil {
		t.Error("all input read")
	}
}

func TestUpdateMoveNoSave(t *testing.T) {
	defer setupUpdate(t)()

	// move, should quit and not save
	reader = bufio.NewReader(strings.NewReader("y\nn\n"))

	runAndValidate(t)
	_, err := index.GetIndexFs().Stat(util.GetIndexFile(config, ext))

	if err == nil {
		t.Fatal("current should not exist")
	}
	oldExists(t)

	allInputRead(t)
}

func TestUpdateSame(t *testing.T) {
	defer setup(t)() // do not add file

	// should short circuit and not read any input
	reader = bufio.NewReader(strings.NewReader("y\ny\n"))

	runAndValidate(t)
	// current exists but has not been updated
	_, err := index.GetIndexFs().Stat(util.GetIndexFile(config, ext))

	if err != nil {
		t.Fatal("current should exist")
	}
	oldDoesNotExist(t)

	if b, _ := reader.Peek(4); len(b) != 4 {
		t.Error("should not read any input", len(b))
	}
}

func TestUpdateNew(t *testing.T) {
	defer setupUpdate(t)()

	// should only need to confirm for new file and not moving existing
	reader = bufio.NewReader(strings.NewReader("y\ny\n"))

	fast = true // increase coverage; should be ignored

	index.GetIndexFs().Remove(util.GetIndexFile(config, ext))

	runAndValidate(t)
	currentUpdated(t)
	oldDoesNotExist(t)

	if reader.Buffered() != 2 {
		t.Error("should not read all input", reader.Buffered())
	}
}

// same code path as TestUpdateNew; overwrite should have no effect
func TestUpdateNewOverwrite(t *testing.T) {
	defer setupUpdate(t)()

	// should only need to confirm for new file and not moving existing
	reader = bufio.NewReader(strings.NewReader("y\ny\n"))

	overwrite = true

	index.GetIndexFs().Remove(util.GetIndexFile(config, ext))

	runAndValidate(t)
	currentUpdated(t)
	oldDoesNotExist(t)

	if reader.Buffered() != 2 {
		t.Error("should not read all input", reader.Buffered())
	}
}

func runAndValidate(t *testing.T) {
	if err := runUpdate(nil, args); err != nil {
		t.Fatal("should not error on update", err)
	}
}

func currentExists(t *testing.T) {
	info, err := index.GetIndexFs().Stat(util.GetIndexFile(config, ext))

	if err != nil {
		t.Fatal("current should exist")
	}

	if !info.ModTime().Equal(idxTime) {
		t.Fatal("current should not be updated")
	}
}

func currentUpdated(t *testing.T) {
	info, err := index.GetIndexFs().Stat(util.GetIndexFile(config, ext))

	if err != nil {
		t.Fatal("current should exist")
	}

	if info.ModTime().Equal(idxTime) {
		fmt.Println(info.ModTime(), idxTime)

		t.Fatal("current should be updated")
	}
}

func oldExists(t *testing.T) {
	//  assumes oldExt is set by runUpdate and not reset afterwards
	_, err := index.GetIndexFs().Stat(util.GetIndexFile(config, oldExt))

	if err != nil {
		t.Fatal("old should exist")
	}

	// no need to check old time; if the file exists it had to have been moved
}

func oldDoesNotExist(t *testing.T) {
	//  assumes oldExt is set by runUpdate and not reset afterwards
	_, err := index.GetIndexFs().Stat(util.GetIndexFile(config, oldExt))

	if err == nil {
		t.Fatal("old should not exist")
	}
}

func allInputRead(t *testing.T) {
	if _, err := reader.ReadByte(); err == nil {
		t.Error("did not read all input")
	}
}
