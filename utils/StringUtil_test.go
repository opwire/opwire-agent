package utils

import(
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestStringPadding(t *testing.T) {
	assert.Equal(t, PadString("label", LEFT, 3, "="), "label")
	assert.Equal(t, PadString("label", CENTER, 3, "="), "label")
	assert.Equal(t, PadString("label", RIGHT, 3, "="), "label")

	assert.Equal(t, PadString("label", LEFT, 6, "___"), "label_")
	assert.Equal(t, PadString("label", CENTER, 6, "___"), "label_")
	assert.Equal(t, PadString("label", RIGHT, 6, "___"), "_label")

	assert.Equal(t, PadString("label", LEFT, 10, " "), "label     ")
	assert.Equal(t, PadString("label", CENTER, 10, "_"), "__label___")
	assert.Equal(t, PadString("label", RIGHT, 10, "*"), "*****label")

	assert.Equal(t, PadString("label", RIGHT, 10, ".-"), "-.-.-label")
}
