package main

import (
	"os"

	cli "github.com/urfave/cli"
	logger "github.com/xenolog/go-tiny-logger"
	"github.com/xenolog/l23/lnx"
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
func RunNetConfig(c *cli.Context) (err error) {
	// Load and Process Network Scheme
	var rr *os.File
	Log.Debug("Run NetworkConfig with network scheme: '%s'", c.GlobalString("ns"))
	ns := new(NetworkScheme)
	if rr, err = os.Open(c.GlobalString("ns")); err != nil {
		Log.Error("Can't open file '%s'", c.GlobalString("ns"))
		Log.Error("%v", err)
		return err
	}
	if err = ns.Load(rr); err != nil {
		Log.Error("Can't process network scheme from '%s'", c.GlobalString("ns"))
		Log.Error("%v", err)
		return err
	}
	Log.Debug("NetworkScheme loaded")

	// generate wanted network topology
	wantedNetState := ns.NpsStatus()
	Log.Debug("NetworkScheme processed")

	// initialize and cinfigure LnxRtPlugin
	lnxRtPlugin := lnx.NewLnxRtPlugin()
	lnxRtPlugin.Init(Log, nil)
	lnxRtPlugin.Observe()
	Log.Debug("LnxRtPlugin initialized")

	// generate diff betwen current and wanted network topology
	diffNetState := lnxRtPlugin.NetworkState().Compare(wantedNetState)
	Log.Debug("NetworkState DIFF ready: \n%v", diffNetState)

	NSoperators := lnxRtPlugin.Operators()

	// // report
	// npCreated := []string{}
	// npRemoved := []string{}
	// npModifyed := []string{}

	// walk through Wanted.NetState network primitives and implement snanges
	// into network configuration
	for _, npName := range wantedNetState.Order {
		Log.Debug("Processing '%v'", npName)
		action, ok := NSoperators[wantedNetState.Link[npName].Action]
		if !ok {
			Log.Warn("Unsupported actiom '%s' for '%s', skipped", action, npName)
			continue
		}
		oper := action.(func() lnx.NpOperator)() // should be blugin_base.NpOperator
		oper.Init(wantedNetState.Link[npName])

		if IndexString(diffNetState.Waste, npName) >= 0 {
			// this NP should be removed
			oper.Remove(true)
			// npRemoved = append(npRemoved, npName)
		} else if IndexString(diffNetState.New, npName) >= 0 {
			// this NP shoujld be created
			oper.Create(true)
			// npCreated = append(npCreated, npName)
		} else if IndexString(diffNetState.Different, npName) >= 0 {
			oper.Modify(true)
			// npModifyed = append(npModifyed, npName)
		}
	}

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
