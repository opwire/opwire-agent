package utils

import (
	"fmt"
	"testing"
	"github.com/stretchr/testify/assert"
)

func Test_convertSettingsToFlatEnvs(t *testing.T) {
	t.Run("empty settings object", func(t *testing.T) {
		envs, err := convertSettingsToFlatEnvs("namespace", map[string]interface{}{})
		assert.Nil(t, err)
		assert.ElementsMatch(t, envs, []string{})
	})
	t.Run("normal settings object", func(t *testing.T) {
		settings := map[string]interface{}{
			"version": "0.1.1",
			"db": map[string]interface{}{
				"username": "e",
				"password": "g",
			},
			"keepalive": true,
			"timeout": 1024,
			"options": nil,
			"rate": 1.4567,
		}
		expected := []string{
			"namespace_version=0.1.1",
			"namespace_db_username=e",
			"namespace_db_password=g",
			"namespace_keepalive=true",
			"namespace_timeout=1024",
			"namespace_rate=1.4567",
		}
		envslist, err := convertSettingsToFlatEnvs("namespace", settings)
		assert.Nil(t, err)
		for _, str := range envslist {
			fmt.Printf("\"%s\",\n", str)
		}
		assert.ElementsMatch(t, envslist, expected)
	})
}
