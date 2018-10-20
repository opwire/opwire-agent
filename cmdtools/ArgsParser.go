package cmdtools

import (
	"os"
	"github.com/acegik/cmdflags"
)

type AgentCmdArgs struct {
	Verbose []bool `short:"v" long:"verbose" description:"Show verbose debug information"`
	Host string `short:"h" long:"host" description:"Agent server host" default:"0.0.0.0"`
	Port uint `short:"p" long:"port" description:"Agent server port" default:"17779"`
	CommandString string `short:"c" long:"default-command" description:"Default command string"`
}

func ParseArgs() (AgentCmdArgs, error) {
	args := AgentCmdArgs{}
	_, err := flags.ParseArgs(&args, os.Args[1:])
	return args, err
}
