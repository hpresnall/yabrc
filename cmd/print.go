package cmd

import (
	"errors"
	"fmt"

	"github.com/hpresnall/yabrc/config"
	"github.com/hpresnall/yabrc/index"
	"github.com/spf13/cobra"
	log "github.com/spf13/jwalterweatherman"
)

var entries bool
var json bool

func init() {
	printCmd.Flags().BoolVarP(&entries, "entries", "", false, "print all entries in the index")
	printCmd.Flags().BoolVarP(&json, "json", "j", false, "JSON output of all entries in the index")
}

var printCmd = &cobra.Command{
	Use:   "print <config_file>",
	Short: "Print index data",
	Args:  cobra.MinimumNArgs(1), // at least one config file
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

	if json {
		fmt.Println("[")
	}

	for n, configFile := range args {
		config, err := config.Load(configFile)

		if err != nil {
			return err
		}

		idx, err := index.Load(&config, ext)

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

		if n < len(args)-1 {
			if json {
				fmt.Println(",")
			} else {
				fmt.Println()
			}
		}
	}

	if json {
		fmt.Println("]")
	}

	return nil
}
