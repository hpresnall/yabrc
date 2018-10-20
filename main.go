package main

import (
	"os"

	"github.com/hpresnall/yabrc/cmd"
	log "github.com/spf13/jwalterweatherman"
)

func main() {
	if err := cmd.Execute(); err != nil {
		if err.Error() != "" {
			log.ERROR.Println(err)
		}

		os.Exit(1)
	}
}
