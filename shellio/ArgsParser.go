package shellio

import (
	"os"
	"github.com/acegik/cmdflags"
	"github.com/opwire/opwire-agent/services"
	"github.com/opwire/opwire-agent/utils"
)

type AgentCmdArgs struct {
	ConfigPath *string `short:"c" long:"config" description:"Explicit configuration file"`
	Host *string `short:"h" long:"host" description:"Agent http server host"`
	Port *uint `short:"p" long:"port" description:"Agent http server port"`
	DirectCommand *string `short:"d" long:"default-command" description:"The command string that will be executed directly"`
	StaticPath []string `short:"s" long:"static-path" description:"Path of static web resources"`
	Verbose []bool `short:"v" long:"verbose" description:"Show verbose debug information"`
}

func ParseArgs() (AgentCmdArgs, error) {
	args := AgentCmdArgs{}
	_, err := flags.ParseArgs(&args, os.Args[1:])
	return args, err
}

func (a *AgentCmdArgs) AgentServerOptions() (*services.ServerOptions) {
	o := &services.ServerOptions{}
	if a.ConfigPath != nil {
		o.ConfigPath = *a.ConfigPath
	}
	if a.Host != nil {
		o.Host = a.Host
	}
	if a.Port != nil {
		o.Port = a.Port
	}
	if a.DirectCommand != nil {
		o.DirectCommand = *a.DirectCommand
	}
	o.StaticPath = utils.ParseDirMappings(a.StaticPath)
	return o
}
