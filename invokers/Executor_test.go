package invokers

import(
	"fmt"
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestExecutor_Run(t *testing.T) {
	t.Run("provide input data through stdin", func(t *testing.T) {
		e, _ := NewExecutor(&ExecutorOptions{
			DefaultCommand: &CommandDescriptor{
				CommandString: "grep hello",
			},
		})
		outBytes, errBytes, state, err := e.RunOnRawData(nil, []byte("hello grep\ngoodbye grep"))
		fmt.Printf("Stdout: [%s]\n", string(outBytes))
		fmt.Printf("Stderr: [%s]\n", string(errBytes))
		assert.Nil(t, err)
		assert.Equal(t, outBytes, []byte("hello grep\n"))
		assert.Equal(t, errBytes, []byte{})
		assert.NotNil(t, state)
		assert.True(t, state.Duration.Seconds() > 0)
	})
	t.Run("run without input data", func(t *testing.T) {
		e, _ := NewExecutor(&ExecutorOptions{
			DefaultCommand: &CommandDescriptor{
				CommandString: "pwd",
			},
		})
		outBytes, errBytes, state, err := e.RunOnRawData(nil, nil)
		fmt.Printf("Stdout: [%s]\n", string(outBytes))
		fmt.Printf("Stderr: [%s]\n", string(errBytes))
		assert.Nil(t, err)
		assert.True(t, len(outBytes) > 0)
		assert.True(t, len(errBytes) == 0)
		assert.NotNil(t, state)
		assert.True(t, state.Duration.Seconds() > 0)
	})
}
