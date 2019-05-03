package services

import (
	"fmt"
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestStateStore_GetAsJSON(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		ss, err := NewStateStore()
		assert.Nil(t, err)
		assert.NotNil(t, ss)

		expected := `{"id":"1234567890","metadata":{"fieldName":["id"]},"types":{"vehicles":["car","bike"]}}`

		ss.Store("config", map[string]interface{} {
			"id": "1234567890",
			"types": map[string]interface{} {
				"vehicles": []string {"car", "bike"},
			},
			"metadata": &struct {
				Field []string `json:"fieldName"`
			}{
				Field: []string {"id"},
			},
		})

		enc, enc_err := ss.GetAsJSON("config")
		assert.Nil(t, enc_err)
		assert.NotNil(t, enc)

		fmt.Printf("JSON: %v\n", string(enc))
		assert.Equal(t, string(enc), expected)
	})
}
