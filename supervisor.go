package main

import(
	"strings"
	"io/ioutil"
	"os/exec"
)

func invokeCommand(cmdString string, pipeInput []byte) ([]byte, []byte) {

	parts := strings.Split(cmdString, " ")

	targetCmd := exec.Command(parts[0], parts[1:]...)

	writer, _ := targetCmd.StdinPipe()
	reader, _ := targetCmd.StdoutPipe()

	targetCmd.Start()
	writer.Write(pipeInput)
	writer.Close()
	nextInput, _ := ioutil.ReadAll(reader)
	targetCmd.Wait()

	out, _ := targetCmd.CombinedOutput()

	return out, nextInput
}
