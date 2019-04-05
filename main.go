package main

import (
	"github.com/opwire/opwire-agent/shellio"
	"github.com/opwire/opwire-agent/services"
)

func main() {
	manifest := &Manifest{}

	if info, ok := manifest.String(); ok {
		shellio.Println(info)
	}

	args, _ := shellio.ParseArgs()

	_, err := services.NewAgentServer(args, manifest)
	if err != nil {
		shellio.Println("Fatal: %s", err.Error())
	}
}
