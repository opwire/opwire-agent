package main

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"github.com/opwire/opwire-agent/entrypoints"
)

func main() {
	entrypoints.StartRestServer()

	var buff bytes.Buffer
	reader, writer := io.Pipe()

	srcCmd := exec.Command("ps", "-e")
	srcCmd.Stdout = writer

	dstCmd := exec.Command("wc", "-l")
	dstCmd.Stdin = reader
	dstCmd.Stdout = &buff

	srcCmd.Start()
	dstCmd.Start()
	srcCmd.Wait()
	writer.Close()
	dstCmd.Wait()

	io.Copy( os.Stdout, &buff )
}
