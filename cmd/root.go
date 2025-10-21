package cmd

import (
	"bufio"
	"io"
	golog "log"
	"os"

	"github.com/spf13/cobra"
	log "github.com/spf13/jwalterweatherman"

	"github.com/hpresnall/yabrc/config"
	"github.com/hpresnall/yabrc/index"
)

var version = "test"

var debug bool
var ext string

// for testing, allow these to be changed
var writer = io.Writer(os.Stdout)
var reader = bufio.NewReader(os.Stdin)

func init() {
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "DEBUG level logging")
	rootCmd.PersistentFlags().StringVarP(&ext, "ext", "e", "_current", "index file extension")

	rootCmd.AddCommand(versionCmd, printCmd, updateCmd, compareCmd)
}

var rootCmd = &cobra.Command{
	Long:          "yabrc - yet another bit rot checker",
	Use:           "yabrc",
	SilenceUsage:  true, // do not print usage on errors
	SilenceErrors: true, // main will log the error via log.ERROR instead of Cobra
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if debug {
			// DEBUG logging, microsecond timestamps and line numbers
			log.SetLogThreshold(log.LevelDebug)
			log.SetStdoutThreshold(log.LevelDebug)

			log.SetFlags(golog.LstdFlags | golog.Lmicroseconds | golog.Lshortfile)
		} else {
			log.SetLogThreshold(log.LevelWarn)
			log.SetStdoutThreshold(log.LevelInfo)

			// remove timestamps from the log output
			log.SetFlags(0)

			// do not output INFO marker; keep WARN & ERROR
			log.INFO.SetPrefix("")
		}
	},
}

// Execute runs the command line application.
func Execute() error {
	return rootCmd.Execute()
}

// helper function to load a Config and Index
func loadIndex(configFile string, ext string) (*index.Index, error) {
	config, err := config.Load(configFile)

	if err != nil {
		return nil, err
	}

	idx, err := index.Load(config, ext)

	if err != nil {
		return nil, err
	}

	return idx, nil
}
