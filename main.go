package main

import (
	"os"
	"github.com/opwire/opwire-agent/shellio"
)

func main() {
	manifest := &Manifest{}

	cmd, err := shellio.NewAgentCommander(manifest)
	if err != nil {
		shellio.Println("Cannot create Commander, error: %s", err.Error())
		os.Exit(2)
	}

	err = cmd.Run()
	if err != nil {
		shellio.Println("Cannot process command, error: %s", err.Error())
		os.Exit(1)
	}
}
