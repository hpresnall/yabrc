package main

import (
	"os"
	"testing"

	log "github.com/spf13/jwalterweatherman"
)

func TestMainNoArgs(t *testing.T) {
	t.Cleanup(silence())

	os.Args = []string{"version"}

	main()
}

func TestMainInvalidArgs(t *testing.T) {
	t.Cleanup(silence())

	os.Args = append(os.Args, "invalid")
	shouldExit = false

	main()
}

func silence() func() {
	out := os.Stdout
	err := os.Stderr

	os.Stdout, _ = os.Open(os.DevNull)
	os.Stderr = os.Stdout

	log.SetStdoutThreshold(log.LevelCritical)

	return func() {
		os.Stdout.Close()
		os.Stdout = out
		os.Stderr = err
	}
}
