// +build !windows

package utils

import (
	"path"
	"strings"
)

func standardizeDirPath(dir string) string {
	p := path.Join(dir, "/")
	if !strings.HasSuffix(p, "/") {
		p += "/"
	}
	return p
}
