package main

import(
	"strings"
	"io/ioutil"
	"os/exec"
	"sync"
)

func invokeCommand(cmdString string, pipeInput []byte) ([]byte, []byte) {

	parts := strings.Split(cmdString, " ")

	cmdObject := exec.Command(parts[0], parts[1:]...)

	writer, _ := cmdObject.StdinPipe()
	reader, _ := cmdObject.StdoutPipe()

	var nextInput []byte

	wg := sync.WaitGroup{}
	wg.Add(2)

	cmdObject.Start()

	go func() {
		defer wg.Done()
		writer.Write(pipeInput)
		writer.Close()
	}()

	go func() {
		defer wg.Done()
		nextInput, _ = ioutil.ReadAll(reader)
	}()

	wg.Wait()

	cmdObject.Wait()

	out, _ := cmdObject.CombinedOutput()

	return out, nextInput
}
