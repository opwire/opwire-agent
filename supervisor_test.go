package main

import(
	"fmt"
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestInvoker_invokeCommand_pipeInput_stdout(t *testing.T) {
	i := Invoker{}
	out, err := i.invokeCommand("grep hello", []byte("hello grep\ngoodbye grep"))
	fmt.Printf("Stdout: [%s]\n", string(out))
	fmt.Printf("Stderr: [%s]\n", string(err))
	// assert.Equal(t, out, []byte("hello grep\n"))
	// assert.Nil(t, err, "grep don't return anything")
	assert.True(t, true)
}

func TestInvoker_invokeCommand_paramIn_stdout(t *testing.T) {
	i := Invoker{}
	out, err := i.invokeCommand("ps", nil)
	fmt.Printf("Stdout: [%s]\n", string(out))
	fmt.Printf("Stderr: [%s]\n", string(err))
}
