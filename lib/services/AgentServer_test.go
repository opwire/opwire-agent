package services

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/opwire/opwire-agent/lib/utils"
)

func TestNewAgentServer(t *testing.T) {
	t.Run("all of components should not be nil", func(t *testing.T) {
		options := &AgentServerOptionsTest{}

		s, err := NewAgentServer(options)

		assert.NotNil(t, s)
		assert.Nil(t, err)

		assert.Nil(t, s.httpServer)
		assert.NotNil(t, s.httpRouter)
		assert.NotNil(t, s.reqSerializer)
		assert.NotNil(t, s.stateStore)
		assert.NotNil(t, s.executor)
	})
}

type AgentServerOptionsTest struct {
	ConfigPath string
	DirectCommand string
	Host string
	Port uint
	StaticPath []string
	Revision string
	Version string
}

func (o *AgentServerOptionsTest) GetConfigPath() string {
	return o.ConfigPath
}

func (o *AgentServerOptionsTest) GetDirectCommand() string {
	return o.DirectCommand
}

func (o *AgentServerOptionsTest) GetHost() string {
	return o.Host
}

func (o *AgentServerOptionsTest) GetPort() uint {
	return o.Port
}

func (o *AgentServerOptionsTest) GetStaticPath() map[string]string {
	return utils.ParseDirMappings(o.StaticPath)
}

func (o *AgentServerOptionsTest) SuppressAutoStart() bool {
	return true
}

func (o *AgentServerOptionsTest) GetRevision() string {
	return o.Revision
}

func (o *AgentServerOptionsTest) GetVersion() string {
	return o.Version
}
