package handlers

import(
	"fmt"
	"io/ioutil"
	"os/exec"
	"strings"
	"sync"
)

type Executor struct {
}

func NewExecutor(command string) (*Executor, error) {
	return &Executor{}, nil
}

func (c *Executor) invokeCommand(cmdString string, pipeInput []byte) ([]byte, []byte) {
	var cmdOut []byte
	var cmdErr []byte

	parts := strings.Split(cmdString, " ")

	cmdObject := exec.Command(parts[0], parts[1:]...)

	writer, _ := cmdObject.StdinPipe()
	pipeOut, _ := cmdObject.StdoutPipe()
	pipeErr, _ := cmdObject.StderrPipe()

	cmdObject.Start()

	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()
		if pipeInput != nil {
			writer.Write(pipeInput)
			writer.Close()
		}
	}()

	go func() {
		defer wg.Done()
		cmdOut, _ = ioutil.ReadAll(pipeOut)
		fmt.Sprintf("stdout/pipe: [%s]", string(cmdOut))

		cmdErr, _ = ioutil.ReadAll(pipeErr)
		fmt.Sprintf("stderr/pipe: [%s]", string(cmdErr))

		cmdObject.Wait()
	}()

	wg.Wait()

	return cmdOut, cmdErr
}
