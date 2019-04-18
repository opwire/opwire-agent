package invokers

import(
	"context"
	"fmt"
	"bytes"
	"io"
	"os/exec"
	"time"
	loq "github.com/opwire/opwire-agent/logging"
	"github.com/opwire/opwire-agent/utils"
)

const BLANK string = ""
const MAIN_RESOURCE string = ":default-resource:"

type TimeSecond float64

type Executor struct {
	newPipeChain func(logger *loq.Logger) (PipeChainRunner)
	resources map[string]*CommandEntrypoint
	logger *loq.Logger
}

type ExecutorOptions struct {
	DefaultCommand *CommandDescriptor
	Logger *loq.Logger
}

type CommandEntrypoint struct {
	Default *CommandDescriptor `json:"default"`
	Methods map[string]*CommandDescriptor `json:"methods"`
	Pattern *string `json:"pattern"`
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
	RequestId string
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

func NewExecutor(opts *ExecutorOptions) (e *Executor, err error) {
	e = &Executor{}
	var defaultCommand *CommandDescriptor
	if opts != nil {
		e.logger = opts.Logger
		defaultCommand = opts.DefaultCommand
	}
	if e.logger == nil {
		e.logger, err = loq.NewLogger(nil)
		if err != nil {
			return nil, err
		}
	}
	if defaultCommand != nil {
		if err := e.Register(defaultCommand, MAIN_RESOURCE); err != nil {
			return nil, err
		}
	}
	return e, nil
}

func GetExecutionTimeout(cd *CommandDescriptor, ci *CommandInvocation) TimeSecond {
	var timeout TimeSecond
	if cd != nil && cd.ExecutionTimeout > 0 {
		timeout = cd.ExecutionTimeout
	}
	if ci != nil && ci.ExecutionTimeout > 0 {
		timeout = ci.ExecutionTimeout
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

func (e *Executor) GetNewPipeChain() (func(logger *loq.Logger) PipeChainRunner) {
	if e.newPipeChain == nil {
		e.newPipeChain = func(logger *loq.Logger) (PipeChainRunner) {
			if logger == nil {
				logger = e.logger
			}
			return NewPipeChain(logger)
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

	runLogger := e.logger
	if opts != nil && len(opts.RequestId) > 0 {
		runLogger = e.logger.With(loq.String("requestId", opts.RequestId))
		defer runLogger.Sync()
	}

	if descriptor, _, _, err := e.ResolveCommandDescriptor(opts); err == nil {
		if cmds, err := buildExecCmds(descriptor); err == nil {
			count := len(cmds)
			if count == 0 {
				return nil, fmt.Errorf("Command not found")
			}

			if opts != nil && opts.Envs != nil {
				envs := e.buildEnvs(opts)
				for _, cmd := range cmds {
					cmd.Env = envs
				}
			}

			state := &ExecutionState{}
			constructor := e.GetNewPipeChain()
			pipeChain := constructor(runLogger)

			timeout := GetExecutionTimeout(descriptor, opts)

			if opts != nil && opts.Context != nil {
				var ctx context.Context
				var cancel context.CancelFunc

				if timeout > 0 {
					ctx, cancel = context.WithTimeout(opts.Context, ConvertSecondToDuration(timeout))
				} else {
					ctx, cancel = context.WithCancel(opts.Context)
				}
				defer cancel() // call cancel as soon as the operations running in this Context complete

				c := make(chan error, 1)

				go func() {
					c <- pipeChain.Run(ib, ob, eb, cmds...)
				}()

				select {
				case <-ctx.Done():
					runLogger.Log(loq.InfoLevel, fmt.Sprintf("Context is timeout after %f seconds.", timeout), loq.Error(ctx.Err()))
					pipeChain.Stop()
					err := <-c
					state.IsTimeout = true
					state.Duration = time.Since(startTime)
					return state, err
				case err := <-c:
					state.Duration = time.Since(startTime)
					return state, err
				}
			}

			// Run without Context
			var timer *time.Timer
			if timeout > 0 {
				timer = time.AfterFunc(ConvertSecondToDuration(timeout), func() {
					runLogger.Log(loq.InfoLevel, fmt.Sprintf("Execution is timeout after %f seconds", timeout))
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

func ConvertSecondToDuration(t TimeSecond) time.Duration {
	return time.Duration(TimeSecond(time.Second) * t)
}

func ConvertDurationToSecond(t time.Duration) TimeSecond {
	return TimeSecond(t.Seconds())
}
