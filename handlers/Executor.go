package handlers

import(
	"fmt"
	"io"
	"io/ioutil"
	"os/exec"
	"strings"
	"sync"
)

const DEFAULT_COMMAND string = "opwire-agent-default"

type Executor struct {
	commands map[string]CommandDescriptor
}

type ExecutorOptions struct {
	Command CommandDescriptor
}

type CommandDescriptor struct {
	CommandString string
}

type CommandInvocation struct {
}

func NewExecutor(opts *ExecutorOptions) (*Executor, error) {
	e := &Executor{}
	if opts != nil {
		e.Register(DEFAULT_COMMAND, &opts.Command)
	}
	return e, nil
}

func (e *Executor) Register(name string, descriptor *CommandDescriptor) (error) {
	if e.commands == nil {
		e.commands = make(map[string]CommandDescriptor)
	}
	if descriptor != nil && len(descriptor.CommandString) > 0 {
		cloned := CommandDescriptor{
			CommandString: descriptor.CommandString,
		}
		e.commands[name] = cloned
	}
	return nil
}

func (e *Executor) Run(opts *CommandInvocation, pipeInput []byte) ([]byte, []byte, error) {
	if cmdObject, ok := e.commands[DEFAULT_COMMAND]; ok {
		var cmdString string = cmdObject.CommandString
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

		return cmdOut, cmdErr, nil
	}
	return nil, nil, nil
}

func (e *Executor) RunWithPipes(opts *CommandInvocation, ip *io.PipeReader, op *io.PipeWriter, ep *io.PipeWriter) (error) {
	return nil
}
