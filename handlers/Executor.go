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
	CommandString string
	Name string
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
	if descriptor := e.getCommandDescriptor(opts); descriptor != nil {
		parts := strings.Split(descriptor.CommandString, " ")
		cmdObject := exec.Command(parts[0], parts[1:]...)
		return runSingleCommand(cmdObject, pipeInput)
	}
	return nil, nil, nil
}

func (e *Executor) getCommandDescriptor(opts *CommandInvocation) *CommandDescriptor {
	if opts != nil && len(opts.CommandString) > 0 {
		return &CommandDescriptor{
			CommandString: opts.CommandString,
		}
	} else if len(opts.Name) > 0 {
		if descriptor, ok := e.commands[opts.Name]; ok {
			return &descriptor
		}
	}
	if descriptor, ok := e.commands[DEFAULT_COMMAND]; ok {
		return &descriptor
	}
	return nil
}

func runSingleCommand(cmdObject *exec.Cmd, pipeInput []byte) ([]byte, []byte, error) {
	var cmdOut []byte
	var cmdErr []byte
	
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

func (e *Executor) RunWithPipes(opts *CommandInvocation, ip *io.PipeReader, op *io.PipeWriter, ep *io.PipeWriter) (error) {
	return nil
}
