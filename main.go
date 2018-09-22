package main

import (
	"os"

	cli "github.com/urfave/cli"
	logger "github.com/xenolog/go-tiny-logger"
	. "github.com/xenolog/l23/utils"
)

const (
	Version = "0.0.1"
)

var (
	Log *logger.Logger
	App *cli.App
	Cfg *AppConfig
	err error
)

func init() {
	// Setup logger
	Log = logger.New()
	Cfg = new(AppConfig)

	// Configure CLI flags and commands
	App = cli.NewApp()
	App.Name = "L23network -- host network topology configurator"
	App.Version = Version
	App.EnableBashCompletion = true
	// App.Usage = "Specify entry point of tree and got subtree for simple displaying"
	App.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:        "debug",
			EnvVar:      "L23_DEBUG",
			Usage:       "Enable debug mode. Show more output",
			Destination: &Cfg.Debug,
		},
		cli.BoolFlag{
			Name:        "dry-run",
			EnvVar:      "L23_DRY-RUN",
			Usage:       "Dry-run mode. Do non changes in the real network configuration",
			Destination: &Cfg.DryRun,
		},
		cli.StringFlag{
			Name:        "ns",
			EnvVar:      "L23_NS",
			Value:       "/etc/network_scheme.yaml",
			Usage:       "Specify path to network scheme file",
			Destination: &Cfg.NsPath,
		},
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
	}, {
		Name:    "netconfig",
		Aliases: []string{"nc", "run"},
		Usage:   "Re-configure network, correspond to network scheme",
		Action:  RunNetConfig,
		Before: func(c *cli.Context) error {
			Log.Debug("Check network scheme exists.")
			_, err := os.Stat(Cfg.NsPath)
			return err
		},
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

func main() {
	App.Run(os.Args)
}

// -----------------------------------------------------------------------------
func RunNetConfig(c *cli.Context) error {
	// runtimeNps := RuntimeNpStatuses__1__exists()

	// wantedNps := RuntimeNpStatuses__1__wanted()
	// diff := runtimeNps.Compare(wantedNps)

	// lnxRtPlugin := NewLnxRtPlugin()
	// lnxRtPlugin.Init(log, nil)
	// operators := lnxRtPlugin.Operators()
	// t.Logf("Diff: %s", diff)

	// // report
	// npCreated := []string{}
	// npRemoved := []string{}
	// npModifyed := []string{}
	// // walk ordr and implement diffs
	// for _, npName := range wantedNps.Order {
	//     action, ok := operators[wantedNps.Link[npName].Action]
	//     if !ok {
	//         t.Logf("Unsupported actiom '%s' for '%s', skipped", action, npName)
	//         t.Fail()
	//         continue
	//     }
	//     oper := action.(func() NpOperator)()
	//     oper.Init(log, lnxRtPlugin.GetHandle(), wantedNps.Link[npName])

	//     t.Logf(npName)
	//     if IndexString(diff.Waste, npName) >= 0 {
	//         // this NP should be removed
	//         oper.Remove(true)
	//         npRemoved = append(npRemoved, npName)
	//     } else if IndexString(diff.New, npName) >= 0 {
	//         // this NP shoujld be created
	//         oper.Create(true)
	//         npCreated = append(npCreated, npName)
	//     } else if IndexString(diff.Different, npName) >= 0 {
	//         oper.Modify(true)
	//         npModifyed = append(npModifyed, npName)
	//     }
	// }

	// // evaluate report
	// if !reflect.DeepEqual(npCreated, []string{"br4", "eth1.101"}) {
	//     t.Logf("Problen while creating resources: %v", npCreated)
	//     t.Fail()
	// }
	// if !reflect.DeepEqual(npModifyed, []string{"eth0", "eth1"}) {
	//     t.Logf("Problen while modifying resources: %v", npModifyed)
	//     t.Fail()
	// }
	return nil
}

// -----------------------------------------------------------------------------
