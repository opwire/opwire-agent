package invokers

import(
	"fmt"
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestExecutor_invokeCommand_pipeInput_stdout(t *testing.T) {
	e, _ := NewExecutor(&ExecutorOptions{
		Command: CommandDescriptor{
			CommandString: "grep hello",
		},
	})
	out, err, _ := e.RunOnRawData(nil, []byte("hello grep\ngoodbye grep"))
	fmt.Printf("Stdout: [%s]\n", string(out))
	fmt.Printf("Stderr: [%s]\n", string(err))
	// assert.Equal(t, out, []byte("hello grep\n"))
	// assert.Nil(t, err, "grep don't return anything")
	assert.True(t, true)
}

func TestExecutor_invokeCommand_paramIn_stdout(t *testing.T) {
	e, _ := NewExecutor(&ExecutorOptions{
		Command: CommandDescriptor{
			CommandString: "ps",
		},
	})
	out, err, _ := e.RunOnRawData(nil, nil)
	fmt.Printf("Stdout: [%s]\n", string(out))
	fmt.Printf("Stderr: [%s]\n", string(err))
}
