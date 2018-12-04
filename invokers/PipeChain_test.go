package invokers

import(
	"bytes"
	"log"
	"os/exec"
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestPipeChain_Run_Ok(t *testing.T) {
	pc := &PipeChain{}
	var i bytes.Buffer
	var o bytes.Buffer
	var e bytes.Buffer
	err := pc.Run(&i, &o, &e,
		exec.Command("echo", "Hello Opwire"),
		exec.Command("wc"),
	)
	output := string(o.Bytes())
	log.Printf("Stdout: [%s]\n", output)
	assert.Equal(t, output, "      1       2      13\n")
	assert.Nil(t, err)
}

func TestPipeChain_Run_Failed(t *testing.T) {
	pc := &PipeChain{}
	var i bytes.Buffer
	var o bytes.Buffer
	var e bytes.Buffer
	err := pc.Run(&i, &o, &e,
		exec.Command("ls", "-7"),
		exec.Command("wc"),
	)
	log.Printf("ErrMsg: [%s]\n", err.Error())
	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), "exit status 2")
}
