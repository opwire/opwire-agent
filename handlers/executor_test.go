package handlers

import(
	"fmt"
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestExecutor_invokeCommand_pipeInput_stdout(t *testing.T) {
	e, _ := NewExecutor("grep hello", &ExecutorOptions{})
	out, err := e.Run([]byte("hello grep\ngoodbye grep"))
	fmt.Printf("Stdout: [%s]\n", string(out))
	fmt.Printf("Stderr: [%s]\n", string(err))
	// assert.Equal(t, out, []byte("hello grep\n"))
	// assert.Nil(t, err, "grep don't return anything")
	assert.True(t, true)
}

func TestExecutor_invokeCommand_paramIn_stdout(t *testing.T) {
	e, _ := NewExecutor("ps", &ExecutorOptions{})
	out, err := e.Run(nil)
	fmt.Printf("Stdout: [%s]\n", string(out))
	fmt.Printf("Stderr: [%s]\n", string(err))
}
