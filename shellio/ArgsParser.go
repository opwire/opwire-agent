package shellio

import (
	"os"
	"github.com/acegik/cmdflags"
)

type AgentCmdArgs struct {
	Verbose []bool `short:"v" long:"verbose" description:"Show verbose debug information"`
	Host string `short:"h" long:"host" description:"Agent server host" default:"0.0.0.0"`
	Port uint `short:"p" long:"port" description:"Agent server port" default:"17779"`
	DefaultCommand string `short:"d" long:"default-command" description:"Default command string"`
	ConfigPath string `short:"c" long:"config" description:"Explicit configuration file"`
	StaticPath []string `short:"s" long:"static-path" description:"Path of static web resources"`
}

func ParseArgs() (AgentCmdArgs, error) {
	args := AgentCmdArgs{}
	_, err := flags.ParseArgs(&args, os.Args[1:])
	return args, err
}
