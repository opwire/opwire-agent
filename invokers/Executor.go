package invokers

import(
	"context"
	"fmt"
	"bytes"
	"io"
	"log"
	"os/exec"
	"time"
	"github.com/opwire/opwire-agent/utils"
)

const BLANK string = ""
const MAIN_RESOURCE string = ":default-resource:"

type TimeSecond int

type Executor struct {
	resources map[string]*CommandEntrypoint
	newPipeChain func() (PipeChainRunner)
}

type ExecutorOptions struct {
	DefaultCommand *CommandDescriptor
}

type CommandEntrypoint struct {
	Default *CommandDescriptor `json:"default"`
	Methods map[string]*CommandDescriptor `json:"methods"`
	Settings map[string]interface{} `json:"settings"`
	SettingsFormat *string `json:"settings-format"`
	settingsEnvs []string
}

type CommandDescriptor struct {
	CommandString string `json:"command"`
	ExecutionTimeout TimeSecond `json:"timeout"`
	subCommands []string
}

type CommandInvocation struct {
	Context context.Context
	Envs []string
	DirectCommand string
	ResourceName string
	MethodName string
	ExecutionTimeout TimeSecond
}

type ExecutionState struct {
	IsTimeout bool
	Duration time.Duration
}

type PipeChainRunner interface {
	Run(ib io.Reader, ob io.Writer, eb io.Writer, chain ...*exec.Cmd) error
	Stop()
}

func NewExecutor(opts *ExecutorOptions) (*Executor, error) {
	if opts == nil {
		opts = &ExecutorOptions{}
	}
	e := &Executor{}
	if opts.DefaultCommand != nil {
		if err := e.Register(opts.DefaultCommand, MAIN_RESOURCE); err != nil {
			return nil, err
		}
	}
	return e, nil
}

func GetExecutionTimeout(cd *CommandDescriptor, ci *CommandInvocation) TimeSecond {
	var timeout TimeSecond
	if cd != nil {
		timeout = cd.ExecutionTimeout
		if ci != nil && ci.ExecutionTimeout > 0 {
			timeout = ci.ExecutionTimeout
		}
	}
	return timeout
}

func extractNames(names []string) (string, string, error) {
	num := len(names)
	switch num {
	case 0:
		return MAIN_RESOURCE, BLANK, nil
	case 1:
		if len(names[0]) == 0 {
			return BLANK, BLANK, fmt.Errorf("Resource name must not be empty")
		}
		return names[0], BLANK, nil
	default:
		if len(names[1]) == 0 || len(names[0]) == 0 {
			return BLANK, BLANK, fmt.Errorf("Resource/Method names must not be empty")
		}
		return names[0], names[1], nil
	}
}

func (e *Executor) GetNewPipeChain() (func() PipeChainRunner) {
	if e.newPipeChain == nil {
		e.newPipeChain = func() (PipeChainRunner) {
			return &PipeChain{}
		}
	}
	return e.newPipeChain
}

func (e *Executor) Register(descriptor *CommandDescriptor, names ...string) (error) {
	if descriptor == nil {
		return fmt.Errorf("Descriptor must not be nil")
	}

	if  len(descriptor.CommandString) == 0 {
		return fmt.Errorf("Command must not be empty")
	}

	preparedCmd, err := prepareCommandDescriptor(descriptor.CommandString)
	if err != nil {
		return err
	}

	preparedCmd.ExecutionTimeout = descriptor.ExecutionTimeout

	resourceName, methodName, err := extractNames(names)

	if err != nil {
		return err
	}

	if e.resources == nil {
		e.resources = make(map[string]*CommandEntrypoint)
	}

	entrypoint, ok := e.resources[resourceName]
	if !ok {
		entrypoint = &CommandEntrypoint{}
		entrypoint.Methods = make(map[string]*CommandDescriptor)
		e.resources[resourceName] = entrypoint
	}

	if methodName == BLANK {
		entrypoint.Default = preparedCmd
		for k := range entrypoint.Methods {
			delete(entrypoint.Methods, k)
		}
	} else {
		entrypoint.Methods[methodName] = preparedCmd
	}

	return nil
}

func (e *Executor) GetSettings(resourceName string) []string {
	if entrypoint, ok := e.resources[resourceName]; ok {
		return entrypoint.settingsEnvs
	}
	return nil
}

func (e *Executor) StoreSettings(prefix string, settings map[string]interface{}, format string, resourceName string) (error) {
	envs, err := utils.TransformSettingsToEnvs(prefix, settings, format)
	if err == nil {
		if entrypoint, ok := e.resources[resourceName]; ok {
			entrypoint.settingsEnvs = envs
		}
	}
	return err
}

func (e *Executor) RunOnRawData(opts *CommandInvocation, inData []byte) ([]byte, []byte, *ExecutionState, error) {
	ib := bytes.NewBuffer(inData)
	var ob bytes.Buffer
	var eb bytes.Buffer
	if state, err := e.Run(ib, opts, &ob, &eb); err != nil {
		return nil, nil, nil, err
	} else {
		return ob.Bytes(), eb.Bytes(), state, err
	}
}

func (e *Executor) Run(ib io.Reader, opts *CommandInvocation, ob io.Writer, eb io.Writer) (*ExecutionState, error) {
	startTime := time.Now()
	if descriptor, _, _, err := e.ResolveCommandDescriptor(opts); err == nil {
		if cmds, err := buildExecCmds(descriptor); err == nil {
			if opts != nil && opts.Envs != nil {
				envs := e.buildEnvs(opts)
				for _, cmd := range cmds {
					cmd.Env = envs
				}
			}
			count := len(cmds)
			if count > 0 {
				state := &ExecutionState{}
				constructor := e.GetNewPipeChain()
				pipeChain := constructor()

				var timer *time.Timer
				timeout := GetExecutionTimeout(descriptor, opts)
				if timeout > 0 {
					timer = time.AfterFunc(time.Second * time.Duration(timeout), func() {
						log.Printf("Execution is timeout after %d seconds\n", timeout)
						pipeChain.Stop()
						state.IsTimeout = true
					})
				}

				err := pipeChain.Run(ib, ob, eb, cmds...)

				if timer != nil {
					timer.Stop()
				}

				state.Duration = time.Since(startTime)

				return state, err
			} else {
				return nil, fmt.Errorf("Command not found")
			}
		} else {
			return nil, err
		}
	} else {
		return nil, err
	}
}

func (e *Executor) ResolveCommandDescriptor(opts *CommandInvocation) (*CommandDescriptor, *string, *string, error) {
	if opts != nil && len(opts.DirectCommand) > 0 {
		descriptor, err := prepareCommandDescriptor(opts.DirectCommand)
		return descriptor, nil, nil, err
	}
	resourceName := getResourceName(opts)
	if entrypoint, ok := e.resources[resourceName]; ok {
		if opts != nil && len(opts.MethodName) > 0 {
			if methodCmd, found := entrypoint.Methods[opts.MethodName]; found {
				return methodCmd, &resourceName, &opts.MethodName, nil
			}
		}
		return entrypoint.Default, &resourceName, nil, nil
	}
	return nil, nil, nil, fmt.Errorf("Command [%s] not found", resourceName)
}

func prepareCommandDescriptor(cmdString string) (*CommandDescriptor, error) {
	descriptor := &CommandDescriptor{}
	if len(cmdString) == 0 {
		return descriptor, fmt.Errorf("Command must not be empty")
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
		return nil, fmt.Errorf("Sub-command must not be empty")
	}
	if parts, err := utils.ParseCmd(cmdString); err != nil {
		return nil, err
	} else {
		return exec.Command(parts[0], parts[1:]...), nil
	}
}

func (e *Executor) buildEnvs(opts *CommandInvocation) []string {
	envs := make([]string, len(opts.Envs))
	copy(envs, opts.Envs)

	resourceName := getResourceName(opts)
	if entrypoint, ok := e.resources[resourceName]; ok {
		settings := entrypoint.settingsEnvs
		if settings != nil {
			envs = append(envs, settings...)
		}
	}

	return envs
}

func getResourceName(opts *CommandInvocation) (string) {
	resourceName := MAIN_RESOURCE
	if opts != nil && len(opts.ResourceName) > 0 {
		resourceName = opts.ResourceName
	}
	return resourceName
}

func runCommand(ib io.Reader, ob io.Writer, eb io.Writer, cmdObject *exec.Cmd) error {
	cmdObject.Stdin = ib
	cmdObject.Stdout = ob
	cmdObject.Stderr = eb
	if err := cmdObject.Start(); err != nil {
		return err
	}
	return cmdObject.Wait()
}
