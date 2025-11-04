package cmd

import (
	"errors"

	"github.com/spf13/cobra"
	log "github.com/spf13/jwalterweatherman"

	"github.com/hpresnall/yabrc/config"
	"github.com/hpresnall/yabrc/index"
	"github.com/hpresnall/yabrc/util"
)

var ext2 string
var ignoreMissing bool

func init() {
	// default to _current to compare current values of 2 indexes (i.e. 2 filesystems)
	compareCmd.Flags().StringVar(&ext2, "ext2", "_current", "extension for the second index")
	compareCmd.Flags().BoolVar(&ignoreMissing, "ignore_missing", false, "ignore missing files in the _first_ index")
}

var compareCmd = &cobra.Command{
	Use:   "compare <config_file_1> [config_file_2]",
	Short: "Compare an existing index to another / updated index",
	Args:  cobra.RangeArgs(1, 2), // 1 or 2 config files
	RunE:  runCompare,
}

func runCompare(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load(args[0])

	if err != nil {
		return err
	}

	newIdx, err := index.Load(&cfg, ext)

	if err != nil {
		return err
	}

	log.INFO.Println()

	var otherCfg config.Config

	// one arg => use the same config
	if len(args) > 1 {
		otherCfg, err = config.Load(args[1])

		if err != nil {
			return err
		}
	} else {
		otherCfg = cfg
	}

	oldIdx, err := index.Load(&otherCfg, ext2)

	if err != nil {
		return err
	}

	log.INFO.Println()

	same := util.Compare(newIdx, oldIdx, ignoreMissing)

	if !same {
		// empty error message => no error logged in main()
		// but _will_ trigger an exit code of 1
		return errors.New("")
	}

	log.INFO.Println("no differences!")
	return nil
}
