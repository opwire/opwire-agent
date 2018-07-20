package main

import(
	"strings"
	"io/ioutil"
	"os/exec"
	"sync"
)

type Invoker struct {
}

func (c *Invoker) invokeCommand(cmdString string, pipeInput []byte) ([]byte, []byte) {

	parts := strings.Split(cmdString, " ")

	cmdObject := exec.Command(parts[0], parts[1:]...)

	writer, _ := cmdObject.StdinPipe()
	reader, _ := cmdObject.StdoutPipe()

	var output []byte
	var pipeOutput []byte

	cmdObject.Start()

	wg := sync.WaitGroup{}
	wg.Add(3)

	go func() {
		defer wg.Done()
		writer.Write(pipeInput)
		writer.Close()
	}()

	go func() {
		defer wg.Done()
		pipeOutput, _ = ioutil.ReadAll(reader)
	}()

	go func() {
		defer wg.Done()
		output, _ = cmdObject.CombinedOutput()
	}()
	
	wg.Wait()

	return output, pipeOutput
}
