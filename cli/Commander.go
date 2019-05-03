package cli

import (
	"fmt"
	"os"
	clp "github.com/urfave/cli"
	"github.com/opwire/opwire-agent/lib/services"
	"github.com/opwire/opwire-agent/lib/utils"
)

type Commander struct {
	app *clp.App
}

func NewCommander(manifest Manifest) (*Commander, error) {
	if manifest == nil {
		return nil, fmt.Errorf("Manifest must not be nil")
	}

	c := new(Commander)

	clp.HelpFlag = clp.BoolFlag{
		Name: "help",
	}
	if info, ok := manifest.String(); ok {
		clp.AppHelpTemplate = fmt.Sprintf("%s\nNOTES:\n   %s\n\n", clp.AppHelpTemplate, info)
	}
	clp.VersionFlag = clp.BoolFlag{
		Name: "version",
	}
	clp.VersionPrinter = func(c *clp.Context) {
		fmt.Fprintf(c.App.Writer, "%s\n", c.App.Version)
	}

	app := clp.NewApp()
	app.Name = "opwire-agent"
	app.Usage = "Bring your command line programs to Rest API"
	app.Version = manifest.GetVersion()

	app.Commands = []clp.Command {
		{
			Name: "serve",
			Aliases: []string{"start"},
			Usage: "Start the service",
			Flags: []clp.Flag{
				clp.StringFlag{
					Name: "config-path, config, c",
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
				f := new(CmdServeFlags)
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
			Usage: "Shows a list of commands or help for one command",
		},
	}
	c.app = app
	return c, nil
}

func (c *Commander) Run() error {
	if c.app == nil {
		return fmt.Errorf("Agent commander has not initialized properly")
	}
	return c.app.Run(os.Args)
}

type Manifest interface {
	GetRevision() string
	GetVersion() string
	String() (string, bool)
}

type CmdServeFlags struct {
	ConfigPath string
	Host string
	Port uint
	DirectCommand string
	StaticPath []string
	manifest Manifest
}

func (a *CmdServeFlags) GetConfigPath() string {
	return a.ConfigPath
}

func (a *CmdServeFlags) GetDirectCommand() string {
	return a.DirectCommand
}

func (a *CmdServeFlags) GetHost() string {
	return a.Host
}

func (a *CmdServeFlags) GetPort() uint {
	return a.Port
}

func (a *CmdServeFlags) GetStaticPath() map[string]string {
	return utils.ParseDirMappings(a.StaticPath)
}

func (a *CmdServeFlags) SuppressAutoStart() bool {
	return false
}

func (a *CmdServeFlags) GetRevision() string {
	if a.manifest == nil {
		return ""
	}
	return a.manifest.GetRevision()
}

func (a *CmdServeFlags) GetVersion() string {
	if a.manifest == nil {
		return ""
	}
	return a.manifest.GetVersion()
}
