
// +build plan9 windows

package utils

import (
	"os"
	"syscall"
)

func ShutdownSignals() []os.Signal {
	return []os.Signal{ syscall.SIGTERM }
}
