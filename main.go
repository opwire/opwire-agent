package main

import (
	"github.com/opwire/opwire-agent/shellio"
	"github.com/opwire/opwire-agent/services"
	"github.com/opwire/opwire-agent/utils"
)

func main() {
	if info, ok := getInfoString(); ok {
		shellio.Println(info)
	}

	args, _ := shellio.ParseArgs()

	_, err := services.NewAgentServer(&services.ServerOptions{
		Host: args.Host,
		Port: args.Port,
		ConfigPath: args.ConfigPath,
		StaticPath: utils.ParseDirMappings(args.StaticPath),
		DefaultCommand: args.DefaultCommand,
		Edition: services.ServerEdition {
			Revision: gitCommit,
			Version: gitTag,
		},
	})

	if err != nil {
		shellio.Println("Fatal: %s", err.Error())
	}
}
