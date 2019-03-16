package main

import (
	"fmt"
)

const artifactId string = "opwire-agent"

var (
	gitCommit string
	gitTag string
	builtAt string
	builtBy string
)

func getInfoString() (string, bool) {
	ok := false
	s := artifactId + " |"

	position := gitTag
	if len(position) == 0 {
		position = gitCommit
	}

	if len(position) > 0 {
		ok = true
		s += fmt.Sprintf(" revision[%s]", position)
	}

	if len(builtAt) > 0 {
		ok = true
		s += fmt.Sprintf(" built @ %s", builtAt)
	}

	if len(builtBy) > 0 {
		ok = true
		s += fmt.Sprintf(" by '%s'", builtBy)
	}

	return s, ok
}
