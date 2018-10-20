package cmd

import (
	"errors"
	"fmt"

	"github.com/hpresnall/yabrc/index"
	"github.com/spf13/cobra"
	log "github.com/spf13/jwalterweatherman"
)

var entries bool
var json bool

func init() {
	printCmd.Flags().BoolVarP(&entries, "entries", "", false, "print all entries in the index")
	printCmd.Flags().BoolVarP(&json, "json", "j", false, "JSON output; no logging")
}

var printCmd = &cobra.Command{
	Use:   "print <config_file>",
	Short: "Print index data",
	Args:  cobra.ExactArgs(1), // config file
	RunE:  runPrint,
}

func runPrint(_ *cobra.Command, args []string) error {
	if entries && json {
		return errors.New("entries and json flags are mutually exclusive")
	}

	if json {
		// reset log so JSON is the only output
		log.SetLogThreshold(log.LevelWarn)
		log.SetStdoutThreshold(log.LevelError)
	}

	idx, err := loadIndex(args[0], ext)

	if err != nil {
		return err
	}

	if entries {
		idx.ForEach(func(e index.Entry) {
			log.INFO.Println(e)
		})
	}

	if json {
		fmt.Fprintln(writer, idx.StringWithEntries())
	}

	return nil
}
