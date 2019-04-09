package shellio

import (
	"fmt"
	"os"
	clp "github.com/urfave/cli"
	"github.com/opwire/opwire-agent/services"
)

type AgentCommander struct {
	app *clp.App
}

func NewAgentCommander(manifest AgentManifest) (*AgentCommander, error) {
	c := new(AgentCommander)

	app := clp.NewApp()
	app.Name = "opwire-agent"
	app.Usage = "Bring your command line programs to Rest API"
	app.Version = manifest.GetVersion()

	app.Commands = []clp.Command {
		{
			Name: "serve",
			Aliases: []string{"s"},
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
					Name: "host, bind, b",
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
				if info, ok := manifest.String(); ok {
					Println(info)
				}
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
