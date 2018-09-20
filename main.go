package main

import (
	// "log"
	// "io/ioutil"
	// "gopkg.in/gin-gonic/gin.v1"
	// yaml "gopkg.in/yaml.v2"
	"os"

	cli "github.com/urfave/cli"
	logger "github.com/xenolog/go-tiny-logger"
)

const (
	Version = "0.0.1"
)

var (
	Log *logger.Logger
	App *cli.App
	err error
)

func init() {
	// // Setup logger
	// Log := logger.New()

	// Configure CLI flags and commands
	App = cli.NewApp()
	App.Name = "L23network -- host network topology configurator"
	App.Version = Version
	App.EnableBashCompletion = true
	// App.Usage = "Specify entry point of tree and got subtree for simple displaying"
	App.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "debug",
			Usage: "Enable debug mode. Show more output",
		},
		// cli.StringFlag{
		//     Name:  "url, u",
		//     Value: "http://127.0.0.1:4001",
		//     Usage: "Specify URL for connect to ETCD",
		// },
	}
	App.Commands = []cli.Command{{
		Name:    "utility",
		Aliases: []string{"u", "util"},
		Usage:   "Execute utility, instead implementing network scheme",
		Subcommands: []cli.Command{{
			Name:   "list-np",
			Usage:  "add a new template",
			Action: UtilityListNetworkPrimitives,
		}, {
			Name:   "list-np-old",
			Usage:  "add a new template",
			Action: UtilityListNetworkPrimitivesOld,
		}},
	}}
	App.Before = func(c *cli.Context) error {
		if c.GlobalBool("debug") {
			Log.SetMinimalFacility(logger.LOG_D)
		} else {
			Log.SetMinimalFacility(logger.LOG_I)
		}
		Log.Debug("L23network started.")
		return nil
	}
	App.CommandNotFound = func(c *cli.Context, cmd string) {
		Log.Printf("Wrong command '%s'", cmd)
		os.Exit(1)
	}
}

func init() {
	// Setup logger
	Log = logger.New()
}

func main() {
	App.Run(os.Args)
}
