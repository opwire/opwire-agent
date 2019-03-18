package main

import (
	"github.com/opwire/opwire-agent/cmdtools"
	"github.com/opwire/opwire-agent/services"
	"github.com/opwire/opwire-agent/utils"
)

func main() {
	if info, ok := getInfoString(); ok {
		cmdtools.Println(info)
	}

	args, _ := cmdtools.ParseArgs()

	_, err := services.NewAgentServer(&services.ServerOptions{
		Host: args.Host,
		Port: args.Port,
		ConfigPath: args.ConfigPath,
		StaticPath: utils.ParseDirMappings(args.StaticPath),
		CommandString: args.CommandString,
		Edition: services.ServerEdition {
			Revision: gitCommit,
			Version: gitTag,
		},
	})

	if err != nil {
		cmdtools.Println("Fatal: %s", err.Error())
	}
}
