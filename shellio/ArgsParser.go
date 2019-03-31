package shellio

import (
	"os"
	"github.com/acegik/cmdflags"
	"github.com/opwire/opwire-agent/services"
	"github.com/opwire/opwire-agent/utils"
)

type AgentCmdArgs struct {
	Host *string `short:"h" long:"host" description:"Agent server host" default:"0.0.0.0"`
	Port *uint `short:"p" long:"port" description:"Agent server port" default:"17779"`
	DirectCommand *string `short:"d" long:"default-command" description:"The command string that will be executed directly"`
	ConfigPath *string `short:"c" long:"config" description:"Explicit configuration file"`
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
	if a.Host != nil {
		o.Host = *a.Host
	}
	if a.Port != nil {
		o.Port = *a.Port
	}
	if a.DirectCommand != nil {
		o.DirectCommand = *a.DirectCommand
	}
	if a.ConfigPath != nil {
		o.ConfigPath = *a.ConfigPath
	}
	o.StaticPath = utils.ParseDirMappings(a.StaticPath)
	return o
}
