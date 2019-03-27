package utils

import(
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestFlatten(t *testing.T) {
	t.Run("empty map is converted to empty list", func(t *testing.T) {
		treeObj := make(map[string]interface{}, 0)
		expected := make(map[string]interface{}, 0)
		output, _ := Flatten("namespace", treeObj)
		assert.Equal(t, output, expected)
	})
}