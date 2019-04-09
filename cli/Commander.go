package cli

import (
	"fmt"
	"os"
	clp "github.com/urfave/cli"
	"github.com/opwire/opwire-agent/services"
	"github.com/opwire/opwire-agent/utils"
)

type AgentCommander struct {
	app *clp.App
}

func NewAgentCommander(manifest AgentManifest) (*AgentCommander, error) {
	if manifest == nil {
		return nil, fmt.Errorf("Manifest must not be nil")
	}

	c := new(AgentCommander)

	clp.HelpFlag = clp.BoolFlag{
		Name: "help",
	}
	clp.VersionFlag = clp.BoolFlag{
		Name: "version",
	}
	if info, ok := manifest.String(); ok {
		clp.AppHelpTemplate = fmt.Sprintf("%s\nNOTES:\n   %s\n\n", clp.AppHelpTemplate, info)
	}

	app := clp.NewApp()
	app.Name = "opwire-agent"
	app.Usage = "Bring your command line programs to Rest API"
	app.Version = manifest.GetVersion()

	app.Commands = []clp.Command {
		{
			Name: "serve",
			Aliases: []string{"start"},
			Usage: "start the service",
			Flags: []clp.Flag{
				clp.StringFlag{
					Name: "config, c",
					Usage: "Explicit configuration file",
				},
				clp.StringFlag{
					Name: "direct-command, default-command, d",
					Usage: "The command string that will be executed directly",
				},
				clp.StringFlag{
					Name: "host, bind-addr, h",
					Usage: "Agent http server host",
				},
				clp.UintFlag{
					Name: "port, p",
					Usage: "Agent http server port",
				},
				clp.StringSliceFlag{
					Name: "static-path, s",
					Usage: "Path of static web resources",
				},
			},
			Action: func(c *clp.Context) error {
				f := new(AgentCmdFlags)
				f.ConfigPath = c.String("config-path")
				f.DirectCommand = c.String("direct-command")
				f.Host = c.String("host")
				f.Port = c.Uint("port")
				f.StaticPath = c.StringSlice("static-path")
				f.manifest = manifest
				_, err := services.NewAgentServer(f)
				return err
			},
		},
		{
			Name: "help",
		},
	}
	c.app = app
	return c, nil
}

func (c *AgentCommander) Run() error {
	if c.app == nil {
		return fmt.Errorf("Agent commander has not initialized properly")
	}
	return c.app.Run(os.Args)
}

type AgentManifest interface {
	GetRevision() string
	GetVersion() string
	String() (string, bool)
}

type AgentCmdFlags struct {
	ConfigPath string
	Host string
	Port uint
	DirectCommand string
	StaticPath []string
	manifest AgentManifest
}

func (a *AgentCmdFlags) GetConfigPath() string {
	return a.ConfigPath
}

func (a *AgentCmdFlags) GetDirectCommand() string {
	return a.DirectCommand
}

func (a *AgentCmdFlags) GetHost() string {
	return a.Host
}

func (a *AgentCmdFlags) GetPort() uint {
	return a.Port
}

func (a *AgentCmdFlags) GetStaticPath() map[string]string {
	return utils.ParseDirMappings(a.StaticPath)
}

func (a *AgentCmdFlags) SuppressAutoStart() bool {
	return false
}

func (a *AgentCmdFlags) GetRevision() string {
	if a.manifest == nil {
		return ""
	}
	return a.manifest.GetRevision()
}

func (a *AgentCmdFlags) GetVersion() string {
	if a.manifest == nil {
		return ""
	}
	return a.manifest.GetVersion()
}
