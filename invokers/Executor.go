package invokers

import(
	"bytes"
	"errors"
	"io/ioutil"
	"os/exec"
	"sync"
	"github.com/opwire/opwire-agent/utils"
)

const DEFAULT_COMMAND string = "opwire-agent-default"

type Executor struct {
	commands map[string]*CommandDescriptor
	pipeChain *PipeChain
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
	Envs []string
}

func NewExecutor(opts *ExecutorOptions) (*Executor, error) {
	e := &Executor{}
	e.pipeChain = NewPipeChain()
	if opts != nil {
		if err := e.Register(DEFAULT_COMMAND, &opts.Command); err != nil {
			return nil, err
		}
	}
	return e, nil
}

func (e *Executor) Register(name string, descriptor *CommandDescriptor) (error) {
	if e.commands == nil {
		e.commands = make(map[string]*CommandDescriptor)
	}
	if descriptor != nil && len(descriptor.CommandString) > 0 {
		if cloned, err := prepareCommandDescriptor(descriptor.CommandString); err == nil {
			e.commands[name] = cloned
		} else {
			return err
		}
	}
	return nil
}

func (e *Executor) RunOnRawData(opts *CommandInvocation, inData []byte) ([]byte, []byte, error) {
	ib := bytes.NewBuffer(inData)
	var ob bytes.Buffer
	var eb bytes.Buffer
	if err := e.Run(ib, opts, &ob, &eb); err != nil {
		return nil, nil, err
	}
	return ob.Bytes(), eb.Bytes(), nil
}

func (e *Executor) Run(ib *bytes.Buffer, opts *CommandInvocation, ob *bytes.Buffer, eb *bytes.Buffer) (err error) {
	if descriptor, err := e.getCommandDescriptor(opts); err == nil {
		if cmds, err := buildExecCmds(descriptor); err == nil {
			if opts.Envs != nil {
				for _, cmd := range cmds {
					cmd.Env = opts.Envs
				}
			}
			count := len(cmds)
			if count > 0 {
				if count == 1 {
					return runCommand(ib, ob, eb, cmds[0])
				}
				return e.pipeChain.Run(ib, ob, eb, cmds...)
			} else {
				return errors.New("Command not found")
			}
		} else {
			return err
		}
	} else {
		return err
	}
}

func (e *Executor) getCommandDescriptor(opts *CommandInvocation) (*CommandDescriptor, error) {
	if opts != nil {
		if len(opts.CommandString) > 0 {
			return prepareCommandDescriptor(opts.CommandString)
		} else if len(opts.Name) > 0 {
			if descriptor, ok := e.commands[opts.Name]; ok {
				return descriptor, nil
			}
		}
	}
	if descriptor, ok := e.commands[DEFAULT_COMMAND]; ok {
		return descriptor, nil
	}
	return nil, errors.New("Default command has not been provided")
}

func prepareCommandDescriptor(cmdString string) (*CommandDescriptor, error) {
	descriptor := &CommandDescriptor{}
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
	if parts, err := utils.ParseCmd(cmdString); err != nil {
		return nil, err
	} else {
		return exec.Command(parts[0], parts[1:]...), nil
	}
}

func runCommand(ib *bytes.Buffer, ob *bytes.Buffer, eb *bytes.Buffer, cmdObject *exec.Cmd) (err error) {
	cmdObject.Stdin = ib
	cmdObject.Stdout = ob
	cmdObject.Stderr = eb
	if err = cmdObject.Start(); err != nil {
		return err
	}
	return cmdObject.Wait()
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
