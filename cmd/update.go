package cmd

import (
	"fmt"
	"strings"

	humanize "github.com/dustin/go-humanize"
	"github.com/spf13/cobra"
	log "github.com/spf13/jwalterweatherman"

	"github.com/hpresnall/yabrc/config"
	"github.com/hpresnall/yabrc/file"
	"github.com/hpresnall/yabrc/index"
	"github.com/hpresnall/yabrc/util"
)

var fast bool
var autosave bool
var overwrite bool
var oldExt string

func init() {
	updateCmd.Flags().BoolVarP(&fast, "fast", "f", false, "only hash new or updated files")
	updateCmd.Flags().BoolVarP(&autosave, "autosave", "a", false, "save the updated index without user confirmation")
	updateCmd.Flags().BoolVarP(&overwrite, "overwrite", "o", false, "overwrite the existing index")
	updateCmd.Flags().StringVar(&oldExt, "old_ext", "", "extension for storing the old Index; ignored with --overwrite")
}

var updateCmd = &cobra.Command{
	Use:   "update <config_file>",
	Short: "Rescan the filesystem and update the index",
	Args:  cobra.ExactArgs(1), // config file
	RunE:  runUpdate,
}

func runUpdate(cmd *cobra.Command, args []string) error {
	config, err := config.Load(args[0])

	if err != nil {
		return err
	}

	indexFile := index.GetPath(config, ext)

	log.INFO.Println()
	log.INFO.Printf("loading existing Index '%s'\n", indexFile)

	existingIdx, err := index.Load(config, ext)

	if err != nil {
		existingIdx = nil

		log.WARN.Printf("cannot open index '%s'; assuming new index creation\n", indexFile)

		if fast {
			fast = false
			log.WARN.Println("ignoring --fast on new index")
		}
	}

	log.INFO.Println()

	var newIdx *index.Index

	if fast {
		newIdx, err = util.BuildIndex(config, existingIdx)
	} else {
		newIdx, err = util.BuildIndex(config, nil)
	}

	if err != nil {
		return err
	}

	if existingIdx != nil {
		log.INFO.Println()
		log.INFO.Printf("comparing '%s' %s vs %s\n", newIdx.Root(), humanize.Time(newIdx.Timestamp()), humanize.Time(existingIdx.Timestamp()))
		same := util.Compare(newIdx, existingIdx, false)

		if same {
			log.INFO.Println("Indexes are the same")
			return nil
		}
	}

	log.INFO.Println()

	// move old index to file with a different extension
	if !overwrite && (existingIdx != nil) {
		if oldExt == "" {
			oldExt = "_" + existingIdx.Timestamp().Format("20060102_150405")
		}
		movedFile := index.GetPath(config, oldExt)

		if autosave {
			log.INFO.Printf("moving '%s' to '%s'\n", indexFile, movedFile)
		} else {
			if !confirm(fmt.Sprintf("move '%s' to '%s'", indexFile, movedFile)) {
				return nil
			}
		}

		err = file.GetFs().Rename(indexFile, movedFile)

		if err != nil {
			return fmt.Errorf("cannot move existing Index to '%s': %v", movedFile, err)
		}
	}

	// save the new index, possibly to the same file name
	if autosave {
		if overwrite && (existingIdx != nil) {
			log.INFO.Printf("overwriting Index '%s'\n", indexFile)
		} else {
			log.INFO.Printf("saving Index to '%s'\n", indexFile)
		}
	} else {
		var prompt string

		if overwrite && (existingIdx != nil) {
			prompt = fmt.Sprintf("overwrite Index '%s'", indexFile)
		} else {
			prompt = fmt.Sprintf("save Index to '%s'", indexFile)
		}

		if !confirm(prompt) {
			return nil
		}
	}

	err = newIdx.Store(indexFile)

	if err != nil {
		return fmt.Errorf("cannot save Index to '%s': %v", indexFile, err)
	}

	return nil
}

func confirm(prompt string) bool {
	for {
		fmt.Fprintf(writer, "%s? (y/n) ", prompt)
		input, err := reader.ReadString('\n')

		if err != nil {
			log.ERROR.Println("cannot read input", err)
			return false
		}

		input = strings.ToLower(input)

		if strings.HasPrefix(input, "y") {
			return true
		}

		if strings.HasPrefix(input, "n") {
			return false
		}
	}
}
