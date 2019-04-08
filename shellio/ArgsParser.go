package shellio

import (
	"github.com/opwire/opwire-agent/utils"
)

type AgentManifest interface {
	GetRevision() string
	GetVersion() string
	String() (string, bool)
}

type AgentCmdFlags struct {
	ConfigPath string
	Host string
	Port uint
	DirectCommand string
	StaticPath []string
	manifest AgentManifest
}

func (a *AgentCmdFlags) GetConfigPath() string {
	return a.ConfigPath
}

func (a *AgentCmdFlags) GetDirectCommand() string {
	return a.DirectCommand
}

func (a *AgentCmdFlags) GetHost() string {
	return a.Host
}

func (a *AgentCmdFlags) GetPort() uint {
	return a.Port
}

func (a *AgentCmdFlags) GetStaticPath() map[string]string {
	return utils.ParseDirMappings(a.StaticPath)
}

func (a *AgentCmdFlags) SuppressAutoStart() bool {
	return false
}

func (a *AgentCmdFlags) GetRevision() string {
	if a.manifest == nil {
		return ""
	}
	return a.manifest.GetRevision()
}

func (a *AgentCmdFlags) GetVersion() string {
	if a.manifest == nil {
		return ""
	}
	return a.manifest.GetVersion()
}
