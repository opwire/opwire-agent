package services

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestNewAgentServer(t *testing.T) {
	t.Run("all of components should not be nil", func(t *testing.T) {
		s, err := NewAgentServer(&ServerOptions{
			SuppressAutoStart: true,
		})

		assert.NotNil(t, s)
		assert.Nil(t, err)

		assert.Nil(t, s.httpServer)
		assert.NotNil(t, s.httpRouter)
		assert.NotNil(t, s.reqSerializer)
		assert.NotNil(t, s.stateStore)
		assert.NotNil(t, s.executor)
	})
}
