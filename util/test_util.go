//go:build !release
// +build !release

package util

import (
	"os"
	"testing"

	"github.com/hpresnall/yabrc/index"
	"github.com/spf13/afero"
	log "github.com/spf13/jwalterweatherman"
	"github.com/spf13/viper"
)

// ConfigFile is the name of the config file used for testing.
var ConfigFile = "config.properties"

// LoadTestConfig sets up a fake config file for use in tests.
// It links the index file system into Viper, saves 'config.properties'
// from the given string and then loads that file into a Config.
// The unction returned should is for test teardown and should be called via defer.
func LoadTestConfig(t *testing.T) (index.Config, func()) {
	oldFs := index.GetIndexFs()

	testFs := afero.NewMemMapFs()
	index.SetIndexFs(testFs)

	configString := `root=testRoot
baseName=testBaseName
savePath=testSavePath
ignoredDirs=.*ignored.*`
	MakeFile(t, ConfigFile, configString, 0644)
	// if MakeFile fails, index fs probably will not be cleaned up

	index.ConfigViperHook = func(v *viper.Viper) {
		v.SetFs(index.GetIndexFs())
	}

	c, err := LoadConfig("config.properties")

	if err != nil {
		// Fatal stops the goroutine before the caller can defer the teardown function
		// run it manually now
		index.SetIndexFs(oldFs)
		index.ConfigViperHook = func(v *viper.Viper) {}

		t.Fatal("should  be able to load config", err)
	}

	return c, func() {
		// reset file system for index and Viper
		index.SetIndexFs(oldFs)
		index.ConfigViperHook = func(v *viper.Viper) {}
	}
}

func makeDir(t *testing.T, dir string) {
	err := index.GetIndexFs().MkdirAll("testRoot", 0755)

	if err != nil {
		t.Fatal("cannot make directory", dir, err)
	}
}

// MakeFile creates a file with the given data.
func MakeFile(t *testing.T, path string, data string, perm os.FileMode) os.FileInfo {
	err := afero.WriteFile(index.GetIndexFs(), path, []byte(data), perm)

	if err != nil {
		t.Fatal("cannot make file", path, err)
	}

	info, err := index.GetIndexFs().Stat(path)

	if err != nil {
		t.Fatal("cannot stat file", path, err)
	}

	return info
}

//BuildTestIndex creates a file layout and builds the Index from those files.
func BuildTestIndex(t *testing.T, config index.Config) *index.Index {
	makeDir(t, "testRoot/test1")
	makeDir(t, "testRoot/test2")
	makeDir(t, "testRoot/test3")
	makeDir(t, "testRoot/test2/sub1")
	makeDir(t, "testRoot/test2/sub2") // sub directory with no files
	makeDir(t, "testRoot/test2/ignored")
	MakeFile(t, "testRoot/test1/"+"test1_1", "data1_1", 0644)
	MakeFile(t, "testRoot/test2/"+"test2_1", "data2_1", 0644)
	MakeFile(t, "testRoot/test2/sub1/"+"test2_sub1_1", "data2_sub1_1", 0644)
	MakeFile(t, "testRoot/test2/sub1/"+"test2_sub1_2", "data2_1_2", 0644)
	MakeFile(t, "testRoot/test2/ignored/ignored1", "ignored1", 0644) // should be skipped via Config.IgnoreDir()
	// should not add 0 byte files
	MakeFile(t, "testRoot/test2/sub1/"+"test2_3", "", 0644)
	// should not index non-files but cannot test because afero's in-memory file system disallows non-file mode bits to be set
	// should not index file manager metadata
	MakeFile(t, "testRoot/.DS_Store", ".DS_Store", 0644)
	MakeFile(t, "testRoot/desktop.ini", "desktop.ini", 0644)

	log.SetLogThreshold(log.LevelTrace)
	idx, err := BuildIndex(config, nil)

	if err != nil {
		t.Fatal("should be able to build an Index;", err)
	}

	if idx.Size() != 4 {
		t.Fatal("Index should have 4 entries, not", idx.Size(), idx.StringWithEntries())
	}

	return idx
}
