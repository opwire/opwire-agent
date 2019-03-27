package utils

import (
	"github.com/jeremywohl/flatten"
)

func Flatten(docName string, tree map[string]interface{}) (map[string]interface{}, error) {
	root := make(map[string]interface{})
	root[docName] = tree
	return flatten.Flatten(root, "", flatten.UnderscoreStyle)
}
