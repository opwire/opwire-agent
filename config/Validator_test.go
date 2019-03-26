package config

import(
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestValidator_Validate(t *testing.T) {
	validator := NewValidator()
	t.Run("simple configuration (valid)", func(t *testing.T) {
		cfg := &Configuration{
			Version: "0.0.1",
		}
		result, err := validator.Validate(cfg)
		assert.Nil(t, err)
		assert.True(t, result.Valid())
	})

	t.Run("simple configuration (invalid)", func(t *testing.T) {
		cfg := &Configuration{
			Version: "a.b.c",
		}
		result, err := validator.Validate(cfg)
		assert.Nil(t, err)
		assert.False(t, result.Valid())
	})
}
