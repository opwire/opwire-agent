package main

import (
	"github.com/opwire/opwire-agent/services"
)

func main() {
	services.NewAgentServer(&services.ServerOptions{
		Port: 8888,
	})
}
