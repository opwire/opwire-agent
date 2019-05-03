package config

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestNewManager(t *testing.T) {
	t.Run("all of components should not be nil", func(t *testing.T) {
		options := &ManagerOptionsTest{}
		s := NewManager(options)

		assert.NotNil(t, s)
		assert.Equal(t, options, s.options)
		assert.NotNil(t, s.locator)
		assert.NotNil(t, s.validator)
	})
}

type ManagerOptionsTest struct {
	ConfigPath string
	Host string
	Port uint
	Version string
}

func (o *ManagerOptionsTest) GetConfigPath() string {
	return o.ConfigPath
}

func (o *ManagerOptionsTest) GetHost() string {
	return o.Host
}

func (o *ManagerOptionsTest) GetPort() uint {
	return o.Port
}

func (o *ManagerOptionsTest) GetVersion() string {
	return o.Version
}
