package cmd

import (
	"errors"

	"github.com/hpresnall/yabrc/util"
	"github.com/spf13/cobra"
	log "github.com/spf13/jwalterweatherman"
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
	newIdx, err := loadIndex(args[0], ext)

	if err != nil {
		return err
	}

	log.INFO.Println()

	var oldIdxConfig string

	// one arg => use the same config
	if len(args) > 1 {
		oldIdxConfig = args[1]
	} else {
		oldIdxConfig = args[0]
	}

	oldIdx, err := loadIndex(oldIdxConfig, ext2)

	if err != nil {
		return err
	}

	log.INFO.Println()

	same := util.Compare(newIdx, oldIdx, ignoreMissing)

	if !same {
		return errors.New("") // empty error message => no error logged in main()
	}

	log.INFO.Println("no differences!")
	return nil
}
