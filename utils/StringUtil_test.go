package utils

import(
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestStringPadding(t *testing.T) {
	assert.Equal(t, PadString("Codes", LEFT, 3, "*"), "Codes")
	assert.Equal(t, PadString("Codes", CENTER, 3, "*"), "Codes")
	assert.Equal(t, PadString("Codes", RIGHT, 3, "*"), "Codes")

	assert.Equal(t, PadString("Codes", LEFT, 6, "___"), "_Codes")
	assert.Equal(t, PadString("Codes", CENTER, 6, "___"), "Codes_")
	assert.Equal(t, PadString("Codes", RIGHT, 6, "___"), "Codes_")

	assert.Equal(t, PadString("Codes", LEFT, 10, "*"), "*****Codes")
	assert.Equal(t, PadString("Codes", CENTER, 10, "_"), "__Codes___")
	assert.Equal(t, PadString("Codes", RIGHT, 10, " "), "Codes     ")

	assert.Equal(t, PadString("Codes", LEFT, 10, "-="), "=-=-=Codes")
}
