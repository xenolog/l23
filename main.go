package main

import (
	"fmt"
	"os"

	cli "github.com/urfave/cli"
	logger "github.com/xenolog/go-tiny-logger"
	"github.com/xenolog/l23/lnx"
	"github.com/xenolog/l23/plugin"
	"github.com/xenolog/l23/u1804"
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
			Destination: &Cfg.Debug, // todo: remove Cfg usage
		},
		cli.BoolFlag{
			Name:        "dry-run",
			EnvVar:      "L23_DRY-RUN",
			Usage:       "Dry-run mode. Do non changes in the real network configuration",
			Destination: &Cfg.DryRun, // todo: remove Cfg usage
		},
		cli.StringFlag{
			Name:        "ns",
			EnvVar:      "L23_NS",
			Value:       "/etc/network_scheme.yaml",
			Usage:       "Specify path to network scheme file",
			Destination: &Cfg.NsPath, // todo: remove Cfg usage
		},
		cli.StringFlag{
			Name:   "store-config",
			EnvVar: "L23_STORE_CONFIG",
			Value:  "/etc/netplan/999-l23network.yaml",
			Usage:  "Specify path for generate network config file. (use 'stdout' if need)",
		},
		cli.BoolFlag{
			Name:   "generate",
			EnvVar: "L23_GENERATE",
			Usage:  "Generate network config",
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
	}, {
		Name:    "store",
		Aliases: []string{"st"},
		Usage:   "Re-write network config, correspond to network scheme",
		Action:  StoreNetConfig,
		Before: func(c *cli.Context) error {
			Log.Debug("Check network scheme exists.")
			_, err := os.Stat(Cfg.NsPath)
			return err
		},
	}}
	App.Before = func(c *cli.Context) error {
		var suffix = ""
		if c.GlobalBool("debug") {
			Log.SetMinimalFacility(logger.LOG_D)
		} else {
			Log.SetMinimalFacility(logger.LOG_I)
		}
		if c.GlobalBool("dry-run") {
			suffix = " (dry-run)"
		}
		Log.Debug("L23network started.%s", suffix)
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
	wantedNetState := ns.TopologyState()
	Log.Debug("NetworkScheme processed")
	Log.Debug("Planned resources ordering is: %s", wantedNetState.Order)

	// initialize and configure LnxRtPlugin
	lnxRtPlugin := lnx.NewLnxRtPlugin()
	lnxRtPlugin.Init(Log, nil)
	lnxRtPlugin.Observe()
	Log.Debug("LnxRtPlugin initialized")

	// generate diff betwen current and wanted network topology
	diffNetState := lnxRtPlugin.Topology().Compare(wantedNetState)
	Log.Debug("NetworkState DIFF ready: \n%v", diffNetState)

	NSoperators := lnxRtPlugin.Operators()

	// // report
	// npCreated := []string{}
	// npRemoved := []string{}
	// npModifyed := []string{}

	// walk through Wanted.NetState network primitives and implement changes
	// into network configuration
	for _, npName := range wantedNetState.Order {
		Log.Debug("Processing '%v'", npName)
		action, ok := NSoperators[wantedNetState.NP[npName].Action]
		if !ok {
			Log.Warn("Unsupported action '%s' for '%s', skipped", action, npName)
			continue
		}
		oper := action.(func() plugin.NpOperator)()
		oper.Init(wantedNetState.NP[npName])

		if IndexString(diffNetState.Waste, npName) >= 0 {
			// this NP should be removed
			oper.Remove(c.GlobalBool("dry-run"))
			// npRemoved = append(npRemoved, npName)
		} else if IndexString(diffNetState.New, npName) >= 0 {
			// this NP shoujld be created
			oper.Create(c.GlobalBool("dry-run"))
			// npCreated = append(npCreated, npName)
		} else if IndexString(diffNetState.Different, npName) >= 0 {
			oper.Modify(c.GlobalBool("dry-run"))
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

	if c.GlobalBool("gnerate") {
		err = StoreNetConfig(c)
	}

	return err
}

func StoreNetConfig(c *cli.Context) (err error) {
	// Load and Process Network Scheme
	var (
		rr *os.File
		ww *os.File
	)
	Log.Debug("Run StoreNetConfig with network scheme: '%s'", c.GlobalString("ns"))
	ns := new(NetworkScheme)
	if rr, err = os.Open(c.GlobalString("ns")); err != nil {
		Log.Error("Can't open file '%s'", c.GlobalString("ns"))
		Log.Error("%v", err)
		return err
	}
	// defer rr.Close()
	if err = ns.Load(rr); err != nil {
		Log.Error("Can't process network scheme from '%s'", c.GlobalString("ns"))
		Log.Error("%v", err)
		return err
	}
	Log.Debug("NetworkScheme loaded")

	// generate wanted network topology
	wantedNetState := ns.TopologyState()
	Log.Debug("NetworkScheme processed")

	// Generate Netplan YAML and store it
	savedConfig := u1804.NewSavedConfig(Log)
	savedConfig.SetWantedState(&wantedNetState.NP)
	if err = savedConfig.Generate(); err != nil {
		Log.Error("Error while Netplan YAML generation: '%s'", err)
		return err
	}
	actualYaml := savedConfig.String()

	configFileName := c.GlobalString("store-config")
	if configFileName == "stdout" || configFileName == "tty" {
		fmt.Printf("---\n%s", actualYaml)
	} else {
		ww, err = os.Create(configFileName)
		if err != nil {
			Log.Error("Can't create file '%s'", configFileName)
			Log.Error("%v", err)
			return err
		}
		defer ww.Close()
		ww.WriteString(actualYaml)
	}
	return err
}

// -----------------------------------------------------------------------------
