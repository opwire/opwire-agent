package shellio

import (
	"fmt"
	"os"
	"github.com/urfave/cli"
	"github.com/opwire/opwire-agent/services"
	"github.com/opwire/opwire-agent/utils"
)

type AgentCommander struct {
	app *cli.App
}

func NewAgentCommander(manifest AgentManifest) (*AgentCommander, error) {
	c := new(AgentCommander)

	app := cli.NewApp()
	app.Name = "opwire-agent"
	app.Usage = "Bring your command line programs to Rest API"
	app.Version = manifest.GetVersion()

	app.Commands = []cli.Command {
		{
			Name: "serve",
			Aliases: []string{"s"},
			Usage: "start the service",
			Action: func(c *cli.Context) error {
				if info, ok := manifest.String(); ok {
					Println(info)
				}
				args, err := ParseArgs(manifest)
				if err != nil {
					return err
				}
				_, err = services.NewAgentServer(args)
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
	return c.app.Run(utils.SliceHead(os.Args, 2))
}
