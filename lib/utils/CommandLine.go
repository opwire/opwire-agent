package utils

import (
	"github.com/mattn/go-shellwords"
)

func ParseCmd(cmd string) ([]string, error) {
	return shellwords.Parse(cmd)
}
