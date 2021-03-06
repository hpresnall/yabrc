package cmd

import (
	"io/ioutil"
	"testing"

	"github.com/hpresnall/yabrc/index"
	"github.com/hpresnall/yabrc/util"
	log "github.com/spf13/jwalterweatherman"
)

var args = []string{util.ConfigFile}

var config index.Config
var idx *index.Index

func setup(t *testing.T) func() {
	// silence output
	log.SetStdoutThreshold(log.LevelError)

	originalReader := reader
	writer = ioutil.Discard

	var teardown func()
	config, teardown = util.LoadTestConfig(t)

	idx = util.BuildTestIndex(t, config)
	util.StoreIndex(idx, config, "_current")

	return func() {
		teardown()

		reader = originalReader

		entries = false
		json = false

		ext = "_current"
		ext2 = "_current"

		overwrite = false
		autosave = false
		fast = false
	}
}

func TestEmptyCmd(t *testing.T) {
	// call Execute instead of rootCmd.Execute() for coverage
	if err := Execute(); err != nil {
		t.Error("should not error on empty command", err)
	}
}

func TestPrint(t *testing.T) {
	defer setup(t)()

	rootCmd.SetArgs([]string{"print", "config.properties"})
	err := rootCmd.Execute()

	if err != nil {
		t.Error("should not error on print", err)
	}

	log.ResetLogCounters()

	// call with debug for coverage
	rootCmd.SetArgs([]string{"print", "--debug", "config.properties"})
	err = rootCmd.Execute()

	if err != nil {
		t.Error("should not error on print", err)
	}

	n := log.LogCountForLevel(log.LevelDebug)

	if n == 0 {
		t.Error("should output DEBUG for --debug")
	}
}

func TestPrintJson(t *testing.T) {
	defer setup(t)()

	entries = false
	json = true

	if err := runPrint(nil, args); err != nil {
		t.Error("should not error on JSON print", err)
	}
}

func TestPrintEntries(t *testing.T) {
	defer setup(t)()

	entries = true
	json = false

	if err := runPrint(nil, args); err != nil {
		t.Error("should not error on print with entries", err)
	}
}

func TestPrintBoth(t *testing.T) {
	entries = true
	json = true

	if err := runPrint(nil, args); err == nil {
		t.Error("should error on print with entries & json")
	}
}

func TestVersion(t *testing.T) {
	runVersion(nil, nil)
}
