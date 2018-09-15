package handlers

import(
	"errors"
	"io"
	"io/ioutil"
	"os/exec"
	"strings"
	"sync"
	"github.com/opwire/opwire-agent/utils"
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
	subCommands []string
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
		if cloned, err := prepareCommandDescriptor(descriptor.CommandString); err == nil {
			e.commands[name] = cloned
		}
	}
	return nil
}

func (e *Executor) Run(opts *CommandInvocation, inData []byte) ([]byte, []byte, error) {
	if descriptor := e.getCommandDescriptor(opts); descriptor != nil {
		count := len(descriptor.subCommands)
		if count <= 0 {
			return nil, nil, errors.New("Command not found")
		} else if count == 1 {
			if cmd, err := buildExecCmd(descriptor.subCommands[0]); err == nil {
				return runSingleCommand(cmd, inData)
			} else {
				return nil, nil, err
			}
		} else {
			
		}
	}
	return nil, nil, nil
}

func (e *Executor) getCommandDescriptor(opts *CommandInvocation) *CommandDescriptor {
	if opts != nil && len(opts.CommandString) > 0 {
		if descriptor, err := prepareCommandDescriptor(opts.CommandString); err == nil {
			return &descriptor
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

func prepareCommandDescriptor(cmdString string) (CommandDescriptor, error) {
	descriptor := CommandDescriptor{}
	if len(cmdString) == 0 {
		return descriptor, errors.New("Command must not be empty")
	}
	descriptor.CommandString = cmdString
	descriptor.subCommands = utils.Split(descriptor.CommandString, "|")
	return descriptor, nil
}

func buildExecCmds(d *CommandDescriptor) ([]*exec.Cmd, error) {
	procs := make([]*exec.Cmd, 0)
	for _, proc := range d.subCommands {
		if cmd, err := buildExecCmd(proc); err == nil {
			procs = append(procs, cmd)
		} else {
			return nil, err
		}
	}
	return procs, nil
}

func buildExecCmd(cmdString string) (*exec.Cmd, error) {
	if len(cmdString) == 0 {
		return nil, errors.New("Sub-command must not be empty")
	}
	parts := strings.Split(cmdString, " ")
	return exec.Command(parts[0], parts[1:]...), nil
}

func runSingleCommand(cmdObject *exec.Cmd, inData []byte) ([]byte, []byte, error) {
	var outData []byte
	var errData []byte
	
	inPipe, _ := cmdObject.StdinPipe()
	outPipe, _ := cmdObject.StdoutPipe()
	errPipe, _ := cmdObject.StderrPipe()

	cmdObject.Start()

	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()
		if inData != nil {
			inPipe.Write(inData)
			inPipe.Close()
		}
	}()

	go func() {
		defer wg.Done()
		outData, _ = ioutil.ReadAll(outPipe)
		errData, _ = ioutil.ReadAll(errPipe)
		cmdObject.Wait()
	}()

	wg.Wait()

	return outData, errData, nil
}

func (e *Executor) RunWithPipes(opts *CommandInvocation, ip *io.PipeReader, op *io.PipeWriter, ep *io.PipeWriter) (error) {
	return nil
}
