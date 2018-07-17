package main

import(
	"fmt"
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestInvokeCommand_pipeInput_pipeOutput(t *testing.T) {
	out, pipe := invokeCommand("grep hello", []byte("hello grep\ngoodbye grep"))
	fmt.Printf("Stdout: %s\n", string(out))
	assert.Nil(t, out, "grep don't return anything")
	assert.Equal(t, pipe, []byte("hello grep\n"))
	fmt.Printf("pipe: %s\n", string(pipe))
}
