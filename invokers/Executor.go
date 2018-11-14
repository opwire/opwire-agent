package invokers

import(
	"bytes"
	"errors"
	"io/ioutil"
	"os/exec"
	"sync"
	"github.com/opwire/opwire-agent/utils"
)

const BLANK string = ""
const DEFAULT_COMMAND string = "opwire-agent-default"

type Executor struct {
	commands map[string]*CommandEntrypoint
	pipeChain *PipeChain
}

type ExecutorOptions struct {
	Command CommandDescriptor
}

type CommandEntrypoint struct {
	Default *CommandDescriptor
	Method map[string]*CommandDescriptor
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
		if err := e.Register(&opts.Command, DEFAULT_COMMAND); err != nil {
			return nil, err
		}
	}
	return e, nil
}

func extractNames(names []string) (string, string, error) {
	num := len(names)
	switch num {
	case 0:
		return DEFAULT_COMMAND, BLANK, nil
	case 1:
		if len(names[0]) == 0 {
			return BLANK, BLANK, errors.New("Resource name must not be empty")
		}
		return names[0], BLANK, nil
	default:
		if len(names[1]) == 0 || len(names[0]) == 0 {
			return BLANK, BLANK, errors.New("Resource/Action names must not be empty")
		}
		return names[0], names[1], nil
	}
}

func (e *Executor) Register(descriptor *CommandDescriptor, names ...string) (error) {
	if descriptor == nil {
		return errors.New("Descriptor must not be nil")
	}

	if  len(descriptor.CommandString) == 0 {
		return errors.New("Command must not be empty")
	}

	preparedCmd, err := prepareCommandDescriptor(descriptor.CommandString)
	if err != nil {
		return err
	}

	resource, action, err := extractNames(names)

	if err != nil {
		return err
	}

	if e.commands == nil {
		e.commands = make(map[string]*CommandEntrypoint)
	}

	resourceEp, ok := e.commands[resource];
	if !ok {
		resourceEp = &CommandEntrypoint{}
		resourceEp.Method = make(map[string]*CommandDescriptor)
		e.commands[resource] = resourceEp
	}

	if action == BLANK {
		resourceEp.Default = preparedCmd
		for k := range resourceEp.Method {
			delete(resourceEp.Method, k)
		}
	} else {
		resourceEp.Method[action] = preparedCmd
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
			if opts != nil && opts.Envs != nil {
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
		} else {
			resourceName := DEFAULT_COMMAND
			if len(opts.Name) > 0 {
				resourceName = opts.Name
			}
			if entrypoint, ok := e.commands[resourceName]; ok {
				return entrypoint.Default, nil
			}
		}
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
