package main

import (
	"os"

	log "github.com/spf13/jwalterweatherman"

	"github.com/hpresnall/yabrc/cmd"
)

// only used for testing
var shouldExit = true

func main() {
	if err := cmd.Execute(); err != nil {
		if err.Error() != "" {
			log.ERROR.Println(err)
		}

		if shouldExit {
			os.Exit(1)
		}
	}
}
