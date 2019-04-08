package lnx

import (
	"fmt"
	"net"
	"strings"

	"github.com/vishvananda/netlink"
	logger "github.com/xenolog/go-tiny-logger"
	npstate "github.com/xenolog/l23/npstate"
	. "github.com/xenolog/l23/plugin"
	. "github.com/xenolog/l23/utils"
	"golang.org/x/sys/unix"
)

const (
	MsgPrefix = "Ubuntu 1804 netplan plugin"
)

// var LnxRtPluginEntryPoint *LnxRtPlugin

// type U1804Plugin struct {
// 	log      *logger.Logger
// 	handle   *netlink.Handle
// 	topology *npstate.TopologyState
// }

// -----------------------------------------------------------------------------

type SCVlan struct {
	Id        int
	Link      string
	Addresses []string
	Dhcp4     bool
	Dhcp6     bool
}
type SCVlans map[string]SCVlan

type SCBridge struct {
	Interfaces []string
	Addresses  []string
	Dhcp4      bool
	Dhcp6      bool
}
type SCBridges map[string]SCBridge

type SCBond struct {
	Interfaces []string
	Addresses  []string
	Dhcp4      bool
	Dhcp6      bool
}
type SCBonds map[string]SCBond

type SCEthernet struct {
	Match     map[string]string
	Addresses []string
	Dhcp4     bool
	Dhcp6     bool
}
type SCEthernets map[string]SCEthernet

type SavedConfig struct {
	log         *logger.Logger
	wantedState *npstate.NPState
	Version     string
	Renderer    string
	Ethernets   SCEthernets
	Bonds       SCBonds
	Vlans       SCVlans
	Bridges     SCBridges
}

// -----------------------------------------------------------------------------

func NewSavedConfig(wantedState *npstate.NPState) *SavedConfig {
	rv := &SavedConfig{
		wantedState: wantedState,
		Version:     "2",
		Renderer:    "networkd",
	}
	// rv.setupGlobals()
	return rv
}

// -----------------------------------------------------------------------------
// -----------------------------------------------------------------------------

// Init -- Runtime linux plugin entry point
func (s *LnxRtPlugin) Init(log *logger.Logger, hh *netlink.Handle) (err error) {
	s.log = log
	if s.handle == nil && hh == nil {
		// generate new handle if need
		if s.handle, err = netlink.NewHandle(); err != nil {
			s.log.Error("%v", err)
			return err
		}
	} else if hh != nil {
		// setup handle
		s.handle = hh
	}
	LnxRtPluginEntryPoint = s
	return nil
}

func (s *LnxRtPlugin) Operators() NpOperators {
	return NpOperators{
		"port":   NewPort,
		"bridge": NewBridge,
		"bond":   NewBond,
		// "endpoint":   NewIPv4,
	}
}

func (s *LnxRtPlugin) Version() string {
	return "LNX RUNTIME PLUGIN: v0.0.1"
}

// Setup netlink handler if need.
// if nil given -- netlink handler will be created automatically
func (s *LnxRtPlugin) setHandle(hh *netlink.Handle) (err error) {
	if s.handle == nil && hh == nil {
		// generate new handle if need
		if s.handle, err = netlink.NewHandle(); err != nil {
			s.log.Error("%v", err)
		}
	} else if hh != nil {
		// setup handle
		s.handle = hh
	}
	return
}

func (s *LnxRtPlugin) Observe() error {
	s.topology = npstate.NewTopologyState()

	var (
		linkList []netlink.Link
		err      error
		attrs    *netlink.LinkAttrs
	)

	s.log.Info("%s: Gathering current network topology", MsgPrefix)

	s.setHandle(nil)

	s.log.Debug("%s: Fetching LinkList from netlink.", MsgPrefix)
	if linkList, err = s.handle.LinkList(); err != nil {
		s.log.Error("%v", err)
		return err
	}

	for _, link := range linkList {
		attrs = link.Attrs()
		linkName := attrs.Name
		s.log.Debug("%s: Processing link '%s'", MsgPrefix, linkName)
		s.topology.NP[linkName] = &npstate.NPState{
			Name:     attrs.Name,
			IfIndex:  attrs.Index,
			LinkType: link.Type(),
		}
		s.topology.NP[linkName].CacheAttrs(attrs)
		if attrs.Flags&net.FlagUp != 0 {
			s.topology.NP[linkName].Online = true
		}
		// s.fillL2stateByNetlinkLink()
		mtu := attrs.MTU
		if mtu == 1500 {
			// workaround for default MTU value
			mtu = 0
		}
		s.topology.NP[linkName].L2 = npstate.L2State{
			// bridge, vlan, bond information should be catched here
			Mtu: mtu,
		}

		if ipaddrs, err := s.handle.AddrList(link, unix.AF_INET); err == nil { // unix.AF_INET === netlink.FAMILY_V4 , but operable under OSX
			// s.topology.NP[linkName].FillByNetlinkAddrList(&ipaddrInfo)
			tmpString := ""
			for _, addr := range ipaddrs {
				// collect IP addresses. This livehack required to prevent empty IPs in the result.
				// it happens :(
				tmpString = fmt.Sprintf("%s %s", tmpString, addr.IPNet.String())
			}
			s.topology.NP[linkName].L3.IPv4 = strings.Fields(tmpString)

		} else {
			s.log.Error("Error while fetch L3 info for '%s' %v", linkName, err)
		}
	}
	s.log.Debug("%s: gathering done.", MsgPrefix)
	return nil
}

func (s *LnxRtPlugin) Topology() *npstate.TopologyState {
	return s.topology
}

func (s *LnxRtPlugin) GetNp(name string) *npstate.NPState {
	rv, ok := s.topology.NP[name]
	if !ok {
		s.log.Error("Network primitive '%s' not found in the stored base", name)
		return nil
	}
	return rv
}

func (s *LnxRtPlugin) GetLogger() *logger.Logger {
	return s.log
}

func (s *LnxRtPlugin) GetHandle() *netlink.Handle {
	return s.handle
}

func NewLnxRtPlugin() RtPlugin {
	rv := new(LnxRtPlugin)
	return rv
}
