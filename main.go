package main

import (
	"log"
	"os"
	"github.com/opwire/opwire-agent/cli"
)

func main() {
	log.SetFlags(log.Ldate | log.Lmicroseconds | log.LUTC)

	manifest := &Manifest{}

	cmd, err := cli.NewCommander(manifest)
	if err != nil {
		cli.Println("Cannot create Commander, error: %s", err.Error())
		os.Exit(2)
	}

	err = cmd.Run()
	if err != nil {
		cli.Println("Cannot process command, error: %s", err.Error())
		os.Exit(1)
	}
}
