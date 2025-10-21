package cmd

import (
	"testing"

	log "github.com/spf13/jwalterweatherman"

	"github.com/hpresnall/yabrc/config"
	"github.com/hpresnall/yabrc/file"
	"github.com/hpresnall/yabrc/index"
)

func TestPrint(t *testing.T) {
	defer setup(t)()

	counter := &log.Counter{}
	log.SetLogListeners(log.LogCounter(counter, log.LevelDebug))

	// pass config twice to cover printing more than one index
	rootCmd.SetArgs([]string{"print", config.TestFile, config.TestFile})
	err := rootCmd.Execute()

	if err != nil {
		t.Error("should not error on print", err)
	}

	// call with debug for coverage
	rootCmd.SetArgs([]string{"print", "--debug", config.TestFile})
	err = rootCmd.Execute()

	if err != nil {
		t.Error("should not error on print", err)
	}

	if counter.Count() == 0 {
		t.Error("should output DEBUG for --debug")
	}
}

func TestPrintJson(t *testing.T) {
	defer setup(t)()

	// pass config twice to cover printing more than one index
	args = []string{config.TestFile, config.TestFile}
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

func TestPrintBadIndex(t *testing.T) {
	defer setup(t)()

	if err := file.GetFs().Remove(index.GetPath(cfg, ext)); err != nil {
		t.Fatalf("cannot remove index from file system")
	}

	if err := runPrint(nil, args); err == nil {
		t.Error("should error on invalid config", err)
	}
}
