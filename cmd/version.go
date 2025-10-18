package cmd

import (
	"runtime"

	"github.com/spf13/cobra"
	log "github.com/spf13/jwalterweatherman"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Get program version & info",
	Run:   runVersion,
}

func runVersion(_ *cobra.Command, _ []string) {
	log.INFO.Printf("%s, version %s", rootCmd.Long, version)
	log.INFO.Println("(C) Copyright 2018-2025 https://github.com/hpresnall")
	log.INFO.Println()
	log.INFO.Println("Built with", runtime.Version())
	log.INFO.Println("Apache 2.0 License <https://www.apache.org/licenses/LICENSE-2.0>")
	log.INFO.Println("For more info, see <https://github.com/hpresnall/yabrc>")
}
