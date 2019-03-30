package utils

import (
	"math"
	"strings"
)

type AlignmentType int

const (
	_ AlignmentType = iota
	LEFT
	CENTER
	RIGHT
)

func PadString(text string, align AlignmentType, lineLength int, seed string) string {
	textLength := len(text)
	if textLength >= lineLength {
		return text
	}

	var line string

	seedLength := len(seed)
	switch align {
	case LEFT:
		loop := math.Ceil(float64(1) + (float64(lineLength-seedLength))/float64(seedLength))
		line = text + strings.Repeat(seed, int(loop))
		line = line[:lineLength]
	case CENTER:
		half := (float64(lineLength - textLength)) / float64(2)
		loop := math.Ceil(half / float64(seedLength))
		line = strings.Repeat(seed, int(loop))[:int(math.Floor(float64(half)))] + text +
				strings.Repeat(seed, int(loop))[:int(math.Ceil(float64(half)))]
	case RIGHT:
		loop := math.Ceil(float64(1) + (float64(lineLength-seedLength))/float64(seedLength))
		line = strings.Repeat(seed, int(loop)) + text
		line = line[len(line)-lineLength:]
	}

	return line
}
