package config

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestNewManager(t *testing.T) {
	t.Run("all of components should not be nil", func(t *testing.T) {
		s := NewManager("v0.1.2", "/home/opwire/demo.cfg")

		assert.NotNil(t, s)
		assert.Equal(t, s.currentVersion, "v0.1.2")
		assert.Equal(t, s.defaultCfgFile, "/home/opwire/demo.cfg")
		assert.NotNil(t, s.locator)
		assert.NotNil(t, s.validator)
	})
}
