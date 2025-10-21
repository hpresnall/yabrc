package cmd

import (
	"io"
	"testing"

	log "github.com/spf13/jwalterweatherman"

	"github.com/hpresnall/yabrc/config"
	"github.com/hpresnall/yabrc/index"
	"github.com/hpresnall/yabrc/util"
)

var args []string

var cfg config.Config
var idx *index.Index

func setup(t *testing.T) func() {
	// silence output
	log.SetStdoutThreshold(log.LevelError)

	originalReader := reader
	writer = io.Discard

	var teardown func()
	cfg, teardown = config.ForTest(t)

	idx = util.IndexForTest(t, cfg)
	index.Store(idx, cfg, "_current")

	args = []string{config.TestFile}

	return func() {
		teardown()

		reader = originalReader

		args = nil

		// reset all command flags to default
		ext = "_current"

		// from print
		entries = false
		json = false

		// from compare
		ext2 = "_current"

		// from update
		fast = false
		autosave = false
		overwrite = false
	}
}

func TestEmptyCmd(t *testing.T) {
	// call Execute instead of rootCmd.Execute() for coverage
	if err := Execute(); err != nil {
		t.Error("should not error on empty command", err)
	}
}

func TestVersion(t *testing.T) {
	// increase test coverage
	runVersion(nil, nil)
}
