package main

import (
	"os"

	"github.com/hpresnall/yabrc/cmd"
	log "github.com/spf13/jwalterweatherman"
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
