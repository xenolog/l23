package lnx

import (
	"github.com/vishvananda/netlink"
	logger "github.com/xenolog/go-tiny-logger"
	npstate "github.com/xenolog/l23/npstate"
	. "github.com/xenolog/l23/plugin"
	. "github.com/xenolog/l23/utils"
	yaml "gopkg.in/yaml.v2"
)

const (
	MsgPrefix = "LNX plugin"
)

var LnxRtPluginEntryPoint *LnxRtPlugin

type LnxRtPlugin struct {
	log    *logger.Logger
	handle *netlink.Handle
	nps    *npstate.TopologyState
}

// -----------------------------------------------------------------------------

type OpBase struct {
	plugin      *LnxRtPlugin
	log         *logger.Logger
	handle      *netlink.Handle
	wantedState *npstate.NPState
	rtState     *npstate.NPState
}

func (s *OpBase) Init(wantedState *npstate.NPState) error {
	s.wantedState = wantedState
	s.rtState = nil
	return nil
}
func (s *OpBase) setupGlobals() {
	s.plugin = LnxRtPluginEntryPoint
	s.handle = s.plugin.handle
	s.log = s.plugin.log
}

func (s *OpBase) Name() string {
	return s.wantedState.Name
}

func (s *OpBase) Link() netlink.Link {
	linkName := s.Name()
	link, err := netlink.LinkByName(linkName)
	if err != nil {
		s.log.Error("%s Can't get attributes for interface '%s' : %v", MsgPrefix, linkName, err)
		return nil
	}
	return link

}

func (s *OpBase) IfIndex() int {
	link := s.Link()
	if link == nil {
		return 0
	}
	return link.Attrs().Index

}

// returns address list in the original order
func (s *OpBase) IPv4addrList() []string {

	rv := []string{}
	if addrs, err := s.handle.AddrList(s.Link(), netlink.FAMILY_V4); err == nil {
		for _, addr := range addrs {
			rv = append(rv, addr.IPNet.String())
		}
	} else {
		s.log.Error("%s Error while fetch L3 info for '%s' %v", MsgPrefix, s.Name(), err)
	}

	return rv
}

func (s *OpBase) allignIPv4list() {
	runtimeIPs := s.IPv4addrList()

	// plan to add non-existing IPs
	toAdd := []string{}
	for _, addr := range s.wantedState.L3.IPv4 {
		if IndexString(runtimeIPs, addr) < 0 {
			toAdd = append(toAdd, addr)
		}
	}

	// plan to remove unwanted IPs
	toRemove := []string{}
	for _, addr := range runtimeIPs {
		if IndexString(s.wantedState.L3.IPv4, addr) < 0 {
			toRemove = append(toRemove, addr)
		}
	}

	// add required IPs
	for _, addr := range toAdd {
		s.log.Debug("%s Adding IPv4 addr '%s' to interface '%s'", MsgPrefix, addr, s.Name())
		if a, err := netlink.ParseAddr(addr); err == nil {
			if err := s.handle.AddrAdd(s.Link(), a); err != nil {
				s.log.Error("%s Can't add IPv4 addr '%s' to interface '%s': %v", MsgPrefix, addr, s.Name(), err)
			}
		} else {
			s.log.Error("%s Can't parse IPv4 addr '%s' while addition: %v", MsgPrefix, addr, err)
		}
	}

	// remove unwanted IPs
	for _, addr := range toRemove {
		s.log.Debug("%s Removing IPv4 addr '%s' on interface '%s'", MsgPrefix, addr, s.Name())
		if a, err := netlink.ParseAddr(addr); err == nil {
			if err := s.handle.AddrDel(s.Link(), a); err != nil {
				s.log.Error("%s Can't remove IPv4 addr '%s' on interface '%s': %v", MsgPrefix, addr, s.Name(), err)
			}
		} else {
			s.log.Error("%s Can't parse IPv4 addr '%s' while addition: %v", MsgPrefix, addr, err)
		}
	}

}

// -----------------------------------------------------------------------------

type L2Port struct {
	OpBase
}

func (s *L2Port) Create(dryrun bool) error {
	if dryrun {
		s.log.Info("%s dryrun: Port '%s' created.", MsgPrefix, s.Name())
		return nil
	}

	s.log.Info("%s Creating port '%s'", MsgPrefix, s.Name())

	// check whether this port is HW device
	link, err := netlink.LinkByName(s.Name())
	if err == nil {
		s.log.Error("%s found existed port", MsgPrefix)
		report, _ := yaml.Marshal(link.Attrs())
		s.log.Debug("%s", report)
		return err
	}

	if s.wantedState.L2.Parent != "" && s.wantedState.L2.Vlan_id > 0 {
		// vlan over parent
		parent_id := 0
		if parent, err := netlink.LinkByName(s.wantedState.L2.Parent); err != nil {
			s.log.Error("%s Can't find interface '%s' as parent for '%s': %v", MsgPrefix, s.wantedState.L2.Parent, s.Name(), err)
			return err
		} else {
			parent_id = parent.Attrs().Index
		}
		vlan := netlink.Vlan{
			VlanId: s.wantedState.L2.Vlan_id,
		}
		vlan.Name = s.Name()
		vlan.ParentIndex = parent_id
		if err := s.handle.LinkAdd(&vlan); err != nil {
			s.log.Error("%s Can't create vlan '%s': %v", MsgPrefix, s.Name(), err)
			return err
		}
		err := s.Modify(false)
		return err
	} else {
		s.log.Error("%s Not implementing, because TBD", MsgPrefix)
	}

	return nil
}

func (s *L2Port) Remove(dryrun bool) error {
	if dryrun {
		s.log.Info("%s dryrun: Port '%s' removed.", MsgPrefix, s.Name())
		return nil
	}

	s.log.Info("%s: Removing port '%s'", MsgPrefix, s.Name())
	link, _ := netlink.LinkByName(s.Name())
	if err := s.handle.LinkSetDown(link); err != nil {
		s.log.Error("%s: error while port removing: %v", MsgPrefix, err)
		return err
	}
	if err := s.handle.LinkAdd(link); err != nil {
		s.log.Error("%s: error while port removing: %v", MsgPrefix, err)
	} else {
		s.log.Info("%s: port removed.", MsgPrefix)
	}
	return nil
}

func (s *L2Port) Modify(dryrun bool) error {
	if dryrun {
		s.log.Info("%s dryrun: Port '%s' modifyed.", MsgPrefix, s.Name())
		return nil
	}
	s.log.Info("%s: Modifying port '%s'", MsgPrefix, s.Name())
	link, _ := netlink.LinkByName(s.Name())
	attrs := link.Attrs()

	if err := s.handle.LinkSetDown(link); err != nil {
		s.log.Error("%s: error while port set to DOWN state: %v", MsgPrefix, err)
	}

	if s.wantedState.L2.Mtu > 0 && s.wantedState.L2.Mtu != attrs.MTU {
		s.log.Debug("%s: setting MTU to: %v", MsgPrefix, s.wantedState.L2.Mtu)
		if err := s.handle.LinkSetMTU(link, s.wantedState.L2.Mtu); err != nil {
			s.log.Error("%s: error while port set MTU: %v", MsgPrefix, err)
		}
	}

	if s.wantedState.L2.Bridge != "" {
		// attach to bridge
		br, err := netlink.LinkByName(s.wantedState.L2.Bridge)
		if br == nil || err != nil {
			s.log.Debug("%s: bridge '%s' can't be located: %v", MsgPrefix, s.wantedState.L2.Bridge, err)
		}

		if err := s.handle.LinkSetMasterByIndex(link, br.Attrs().Index); err != nil {
			s.log.Debug("%s: port can't be became a member of bridge '%s': %v", MsgPrefix, s.wantedState.L2.Bridge, err)
			return err
		}
	} else {
		// remove from bridge
		if attrs.MasterIndex != 0 {
			if err := s.handle.LinkSetNoMaster(link); err != nil {
				s.log.Debug("%s: port can't removed from bridge: %v", MsgPrefix, err)
			}
		}
	}

	if s.wantedState.Online {
		s.log.Debug("%s: setting to UP state", MsgPrefix)
		if err := s.handle.LinkSetUp(link); err != nil {
			s.log.Error("%s: error while port set to UP state: %v", MsgPrefix, err)
		}
	}

	s.allignIPv4list()

	return nil
}

func NewPort() NpOperator {
	rv := new(L2Port)
	rv.setupGlobals()
	return rv
}

// -----------------------------------------------------------------------------

type L2Bridge struct {
	OpBase
}

func (s *L2Bridge) Create(dryrun bool) (err error) {
	if dryrun {
		s.log.Info("%s dryrun: Bridge '%s' created.", MsgPrefix, s.Name())
		return nil
	}

	s.log.Info("%s Creating bridge '%s'", MsgPrefix, s.Name())
	br := netlink.Bridge{}
	br.Name = s.Name()
	if err = s.handle.LinkAdd(&br); err != nil {
		s.log.Error("%s: error while bridge creating: %v", MsgPrefix, err)
		return err
	} else {
		s.log.Info("%s: bridge created.", MsgPrefix)
	}

	err = s.Modify(false)

	return err
}

func (s *L2Bridge) Remove(dryrun bool) (err error) {
	if dryrun {
		s.log.Info("%s: dryrun: Bridge '%s' removed.", MsgPrefix, s.Name())
		return nil
	}
	s.log.Info("%s: Removing bridge '%s'", MsgPrefix, s.Name())
	link, _ := netlink.LinkByName(s.Name())
	if err = s.handle.LinkSetDown(link); err != nil {
		s.log.Error("%s: error while bridge removing: %v", MsgPrefix, err)
		return err
	}
	if err = s.handle.LinkAdd(link); err != nil {
		s.log.Error("%s: error while bridge removing: %v", MsgPrefix, err)
	} else {
		s.log.Info("%s: bridge removed.", MsgPrefix)
	}
	return err
}

func (s *L2Bridge) Modify(dryrun bool) (err error) {
	if dryrun {
		s.log.Info("%s dryrun: Bridge '%s' modifyed.", MsgPrefix, s.Name())
		return nil
	}

	s.log.Info("%s: Modifying bridge '%s'", MsgPrefix, s.Name())
	link, _ := netlink.LinkByName(s.Name())
	attrs := link.Attrs()

	if err = s.handle.LinkSetDown(link); err != nil {
		s.log.Error("%s: error while bridge set to DOWN state: %v", MsgPrefix, err)
	}

	if s.wantedState.L2.Mtu > 0 && s.wantedState.L2.Mtu != attrs.MTU {
		s.log.Debug("%s: setting MTU to: %v", MsgPrefix, s.wantedState.L2.Mtu)
		if err = s.handle.LinkSetMTU(link, s.wantedState.L2.Mtu); err != nil {
			s.log.Error("%s: error while bridge set MTU: %v", MsgPrefix, err)
		}
	}

	if s.wantedState.Online {
		s.log.Debug("%s: setting to UP state", MsgPrefix)
		if err = s.handle.LinkSetUp(link); err != nil {
			s.log.Error("%s: error while bridge set to UP state: %v", MsgPrefix, err)
		}
	}

	s.allignIPv4list()

	return err
}

func NewBridge() NpOperator {
	rv := new(L2Bridge)
	rv.setupGlobals()
	return rv
}

// -----------------------------------------------------------------------------

// func NewIPv4() NpOperator {
// 	rv := new(NewIPv4)
// 	return rv
// }

// -----------------------------------------------------------------------------

func (s *LnxRtPlugin) Init(log *logger.Logger, hh *netlink.Handle) (err error) {
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
	s.log = log
	LnxRtPluginEntryPoint = s
	return nil
}

func (s *LnxRtPlugin) Operators() NpOperators {
	return NpOperators{
		"port":   NewPort,
		"bridge": NewBridge,
		// "endpoint":   NewIPv4,
	}
}

func (s *LnxRtPlugin) Version() string {
	return "LNX RUNTIME PLUGIN: v0.0.1"
}

func (s *LnxRtPlugin) Observe() error {
	s.nps = npstate.NewTopologyState()
	return s.nps.ObserveRuntime()
}

func (s *LnxRtPlugin) NetworkState() *npstate.TopologyState {
	return s.nps
}

func (s *LnxRtPlugin) GetNp(name string) *npstate.NPState {
	rv, ok := s.nps.Link[name]
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

func NewLnxRtPlugin() *LnxRtPlugin {
	rv := new(LnxRtPlugin)
	return rv
}
