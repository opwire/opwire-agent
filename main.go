package main

import (
	"github.com/opwire/opwire-agent/shellio"
	"github.com/opwire/opwire-agent/services"
)

func main() {
	if info, ok := getInfoString(); ok {
		shellio.Println(info)
	}

	args, _ := shellio.ParseArgs()
	opts := args.AgentServerOptions()
	opts.Edition = services.ServerEdition {
		Revision: gitCommit,
		Version: gitTag,
	}

	_, err := services.NewAgentServer(opts)
	if err != nil {
		shellio.Println("Fatal: %s", err.Error())
	}
}
