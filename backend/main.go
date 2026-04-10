package main

import (
	"log"
	"os"

	"github.com/potibm/billedapparat/cmd"
)

var version = "0.0.0"

func startup() int {
	if err := cmd.Execute(); err != nil {
		log.Printf("Fatal error while starting: %v", err)

		return int(cmd.Exit_Software)
	}

	return int(cmd.Exit_OK)
}

func main() {
	os.Exit(startup())
}
