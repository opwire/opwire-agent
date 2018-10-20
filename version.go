package main

import (
	"fmt"
)

var (
	gitCommit string
	builtAt string
	builtBy string
)

func getInfoString() (string, bool) {
	ok := false
	s := "opwire-agent |"

	if len(gitCommit) > 0 {
		ok = true
		s += fmt.Sprintf(" git-commit[%s]", gitCommit)
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
