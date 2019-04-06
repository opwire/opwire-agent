package shellio

import (
	"os"
	"github.com/acegik/cmdflags"
	"github.com/opwire/opwire-agent/utils"
)

type AgentCmdArgs struct {
	ConfigPath *string `short:"c" long:"config" description:"Explicit configuration file"`
	Host *string `short:"h" long:"host" description:"Agent http server host"`
	Port *uint `short:"p" long:"port" description:"Agent http server port"`
	DirectCommand *string `short:"d" long:"default-command" description:"The command string that will be executed directly"`
	StaticPath []string `short:"s" long:"static-path" description:"Path of static web resources"`
	Verbose []bool `short:"v" long:"verbose" description:"Show verbose debug information"`
	manifest AgentManifest
}

type AgentManifest interface {
	GetRevision() string
	GetVersion() string
}

func ParseArgs(manifest AgentManifest) (*AgentCmdArgs, error) {
	args := &AgentCmdArgs{}
	_, err := flags.ParseArgs(args, os.Args[1:])
	args.manifest = manifest
	return args, err
}

func (a *AgentCmdArgs) GetConfigPath() string {
	if a.ConfigPath == nil {
		return ""
	}
	return *a.ConfigPath
}

func (a *AgentCmdArgs) GetDirectCommand() string {
	if a.DirectCommand == nil {
		return ""
	}
	return *a.DirectCommand
}

func (a *AgentCmdArgs) GetHost() string {
	if a.Host == nil {
		return ""
	}
	return *a.Host
}

func (a *AgentCmdArgs) GetPort() uint {
	if a.Port == nil {
		return 0
	}
	return *a.Port
}

func (a *AgentCmdArgs) GetStaticPath() map[string]string {
	return utils.ParseDirMappings(a.StaticPath)
}

func (a *AgentCmdArgs) SuppressAutoStart() bool {
	return false
}

func (a *AgentCmdArgs) GetRevision() string {
	if a.manifest == nil {
		return ""
	}
	return a.manifest.GetRevision()
}

func (a *AgentCmdArgs) GetVersion() string {
	if a.manifest == nil {
		return ""
	}
	return a.manifest.GetVersion()
}
