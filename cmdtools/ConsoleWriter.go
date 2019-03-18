package cmdtools

import (
	"fmt"
)

func Println(format string, a ...interface{}) {
	fmt.Printf(format, a...)
	fmt.Println()
}
