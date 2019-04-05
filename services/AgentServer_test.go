package services

import (
	"testing"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestNewAgentServer(t *testing.T) {
	t.Run("all of components should not be nil", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		options := NewMockAgentServerOptions(ctrl)
		options.EXPECT().GetConfigPath().Return("").AnyTimes()
		options.EXPECT().GetStaticPath().Return(map[string]string{}).AnyTimes()
		options.EXPECT().GetHost().Return("").AnyTimes()
		options.EXPECT().GetPort().Return(uint(0)).AnyTimes()
		options.EXPECT().SuppressAutoStart().Return(true).AnyTimes()

		edition := NewMockAgentServerEdition(ctrl)
		edition.EXPECT().GetRevision().Return("g586f436").AnyTimes()
		edition.EXPECT().GetVersion().Return("v1.0.6-9-g586f436").AnyTimes()

		s, err := NewAgentServer(options, edition)

		assert.NotNil(t, s)
		assert.Nil(t, err)

		assert.Nil(t, s.httpServer)
		assert.NotNil(t, s.httpRouter)
		assert.NotNil(t, s.reqSerializer)
		assert.NotNil(t, s.stateStore)
		assert.NotNil(t, s.executor)
	})
}
