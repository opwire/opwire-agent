package config

import(
	_ "fmt"
	"os"
	"testing"
	"github.com/stretchr/testify/assert"
)

var DEFAULT_SERIES []string = []string { "arg", "env", "bin", "cwd", "xdg", "home", "etc" }

func Test_getConfigSeries(t *testing.T) {
	t.Run("default series", func(t *testing.T) {
		assert.Equal(t, getConfigSeries(), DEFAULT_SERIES)
	})

	t.Run("empty series", func(t *testing.T) {
		os.Setenv("OPWIRE_AGENT_CONFIG_SERIES", " ")
		assert.Equal(t, getConfigSeries(), DEFAULT_SERIES)
	})

	t.Run("customized series", func(t *testing.T) {
		os.Setenv("OPWIRE_AGENT_CONFIG_SERIES", "env, arg, cwd, home")
		assert.Equal(t, getConfigSeries(), []string { "env", "arg", "cwd", "home" })
	})
}
