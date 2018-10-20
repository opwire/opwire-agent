package main

import (
	"github.com/opwire/opwire-agent/cmdtools"
	"github.com/opwire/opwire-agent/services"
)

func main() {
	if info, ok := getInfoString(); ok {
		cmdtools.Println(info)
	}

	args, _ := cmdtools.ParseArgs()
	
	services.NewAgentServer(&services.ServerOptions{
		Host: args.Host,
		Port: args.Port,
		CommandString: args.CommandString,
	})
}
