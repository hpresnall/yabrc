//go:build !release
// +build !release

package util

import (
	"testing"

	"github.com/hpresnall/yabrc/config"
	"github.com/hpresnall/yabrc/index"
	"github.com/hpresnall/yabrc/test"
)

// creates a file layout and builds the Index from those files.
func IndexForTest(t *testing.T) *index.Index {
	config := config.ForTest(t)
	root := config.Root()

	test.MakeDir(t, root+"/test1")
	test.MakeDir(t, root+"/test2")
	test.MakeDir(t, root+"/test3")
	test.MakeDir(t, root+"/test2/sub1")
	test.MakeDir(t, root+"/test2/sub2") // sub directory with no files
	test.MakeDir(t, root+"/test2/ignored")
	test.MakeFile(t, root+"/test1/"+"test1_1", "data1_1", 0644)
	test.MakeFile(t, root+"/test2/"+"test2_1", "data2_1", 0644)
	test.MakeFile(t, root+"/test2/sub1/"+"test2_sub1_1", "data2_sub1_1", 0644)
	test.MakeFile(t, root+"/test2/sub1/"+"test2_sub1_2", "data2_1_2", 0644)
	test.MakeFile(t, root+"/test2/ignored/ignored1", "ignored1", 0644) // should be skipped via Config.IgnoreDir()
	// should not add 0 byte files
	test.MakeFile(t, root+"/test2/sub1/"+"test2_3", "", 0644)
	test.MakeFile(t, root+"/test3/"+"test3", "data3", 0644) // file for deletion in compare test
	// should not index non-files but cannot test because afero's in-memory file system disallows non-file mode bits to be set
	// should not index file manager metadata
	test.MakeFile(t, root+"/.DS_Store", ".DS_Store", 0644)
	test.MakeFile(t, root+"/desktop.ini", "desktop.ini", 0644)

	idx, err := BuildIndex(&config, nil)

	if err != nil {
		t.Fatal("should be able to build an Index;", err)
	}

	if idx.Size() != 5 {
		t.Fatal("Index should have 4 entries, not", idx.Size(), idx.StringWithEntries())
	}

	return idx
}
