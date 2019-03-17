package utils

import (
	"path"
	"os"
	"strings"
)

func IsFileExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func IsDir(name string) bool {
	if stat, err := os.Stat(name); !os.IsNotExist(err) {
		return stat.IsDir()
	}
	return false
}

func standardizeUrlPath(dir string) string {
	p := path.Join("/", dir, "/")
	if p != "/" {
		p += "/"
	}
	return p
}

func standardizeDirPath(dir string) string {
	p := path.Join(dir, "/")
	if p != "/" {
		p += "/"
	}
	return p
}

func ParseDirMappings(paths []string) map[string]string {
	mapping := make(map[string]string)
	for _, pairStr := range paths {
		pair := strings.Split(pairStr, ":")
		pairLen := len(pair)
		if pairLen > 0 {
			urlPath := "/"
			if pairLen > 1 {
				urlPath = standardizeUrlPath(pair[1])
			}
			mapping[urlPath] = standardizeDirPath(pair[0])
		}
	}
	return mapping
}
