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
	outBytes, errBytes, _, err := e.RunOnRawData(nil, []byte("hello grep\ngoodbye grep"))
	fmt.Printf("Stdout: [%s]\n", string(outBytes))
	fmt.Printf("Stderr: [%s]\n", string(errBytes))
	assert.Nil(t, err)
	assert.Equal(t, outBytes, []byte("hello grep\n"))
	assert.Equal(t, errBytes, []byte{})
}

func TestExecutor_invokeCommand_paramIn_stdout(t *testing.T) {
	e, _ := NewExecutor(&ExecutorOptions{
		Command: CommandDescriptor{
			CommandString: "ps",
		},
	})
	out, err, _, _ := e.RunOnRawData(nil, nil)
	fmt.Printf("Stdout: [%s]\n", string(out))
	fmt.Printf("Stderr: [%s]\n", string(err))
}
