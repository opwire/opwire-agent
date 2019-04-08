package utils

import (
	"fmt"
	"strings"
)

func CombineErrors(label string, messages []string) error {
	errstrs := []string {}
	for _, msg := range messages {
		if len(msg) > 0 {
			errstrs = append(errstrs, msg)
		}
	}
	if len(errstrs) > 0 {
		errstrs = append([]string {label}, errstrs...)
		return fmt.Errorf(strings.Join(errstrs, "\n - "))
	}
	return nil
}
