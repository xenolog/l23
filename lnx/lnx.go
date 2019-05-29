package lnx

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"sort"
	"strings"

	"github.com/vishvananda/netlink"
	logger "github.com/xenolog/go-tiny-logger"
	npstate "github.com/xenolog/l23/npstate"
	. "github.com/xenolog/l23/plugin"
	. "github.com/xenolog/l23/utils"
	"golang.org/x/sys/unix"
	yaml "gopkg.in/yaml.v3"
)

const (
	MsgPrefix = "LNX plugin"
)

var LnxRtPluginEntryPoint *LnxRtPlugin

type LnxRtPlugin struct {
	log      *logger.Logger
	handle   *netlink.Handle
	topology *npstate.TopologyState
}

type BondSlavesDiffType struct {
	toAdd    []string
	toRemove []string
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

func (s *OpBase) AddToBridge(brName string) error {
	// attach to bridge
	br, err := netlink.LinkByName(brName)
	if br == nil || err != nil {
		s.log.Debug("%s: bridge '%s' can't be located: %v", MsgPrefix, brName, err)
		return err
	}

	if err := s.handle.LinkSetMasterByIndex(s.Link(), br.Attrs().Index); err != nil {
		s.log.Debug("%s: '%s' can't be became a member of bridge '%s': %v", MsgPrefix, s.Name(), brName, err)
		return err
	}
	return nil
}

func (s *OpBase) RemoveFromBridge() error {
	// remove from bridge
	// todo(sv): Check if master is bridge, not bond !!!
	link := s.Link()
	if link.Attrs().MasterIndex != 0 {
		if err := s.handle.LinkSetNoMaster(link); err != nil {
			s.log.Debug("%s: '%s' can't be removed from bridge: %v", MsgPrefix, s.Name(), err)
			return err
		}
	}
	return nil
}

// returns address list in the original order
func (s *OpBase) IPv4addrList() []string {

	rv := []string{}
	if addrs, err := s.handle.AddrList(s.Link(), unix.AF_INET); err == nil { // unix.AF_INET === netlink.FAMILY_V4 , but operable under OSX
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

	s.log.Debug("%s %s: IPv4 addresses found: %s", MsgPrefix, s.Name(), runtimeIPs)
	s.log.Debug("%s %s: IPv4 addresses to add: %s", MsgPrefix, s.Name(), toAdd)
	s.log.Debug("%s %s: IPv4 addresses to remove: %s", MsgPrefix, s.Name(), toRemove)

	// add required IPs
	for _, addr := range toAdd {
		s.log.Debug("%s %s: Adding IPv4 addr '%s'", MsgPrefix, s.Name(), addr)
		if a, err := netlink.ParseAddr(addr); err == nil {
			if err := s.handle.AddrAdd(s.Link(), a); err != nil {
				s.log.Error("%s %sCan't add IPv4 addr '%s': %v", MsgPrefix, s.Name(), addr, err)
			}
		} else {
			s.log.Error("%s Can't parse IPv4 addr '%s' while addition: %v", MsgPrefix, addr, err)
		}
	}

	// remove unwanted IPs
	for _, addr := range toRemove {
		s.log.Debug("%s %s Removing IPv4 addr '%s'", MsgPrefix, s.Name(), addr)
		if a, err := netlink.ParseAddr(addr); err == nil {
			if err := s.handle.AddrDel(s.Link(), a); err != nil {
				s.log.Error("%s %s Can't remove IPv4 addr '%s': %v", MsgPrefix, s.Name(), addr, err)
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
		parentID := 0
		if parent, err := netlink.LinkByName(s.wantedState.L2.Parent); err != nil {
			s.log.Error("%s Can't find interface '%s' as parent for '%s': %v", MsgPrefix, s.wantedState.L2.Parent, s.Name(), err)
			return err
		} else {
			parentID = parent.Attrs().Index
		}
		vlan := netlink.Vlan{
			VlanId: s.wantedState.L2.Vlan_id,
		}
		vlan.Name = s.Name()
		vlan.ParentIndex = parentID
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
	if err := s.handle.LinkDel(link); err != nil {
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
	// needUp := false
	link, _ := netlink.LinkByName(s.Name())
	attrs := link.Attrs()

	// if err := s.handle.LinkSetDown(link); err != nil {
	// 	s.log.Error("%s: error while port set to DOWN state: %v", MsgPrefix, err)
	// }

	if s.wantedState.L2.Mtu > 0 && s.wantedState.L2.Mtu != attrs.MTU {
		s.log.Debug("%s: setting MTU to: %v", MsgPrefix, s.wantedState.L2.Mtu)
		if err := s.handle.LinkSetMTU(link, s.wantedState.L2.Mtu); err != nil {
			s.log.Error("%s: error while port set MTU: %v", MsgPrefix, err)
		}
	}

	if s.wantedState.L2.Bridge != "" {
		s.AddToBridge(s.wantedState.L2.Bridge)
	} else {
		s.RemoveFromBridge()
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
	if err = s.handle.LinkDel(link); err != nil {
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

func (s *L2Bridge) AddToBridge(brName string) error {
	s.log.Error("%s: There are no able to add the bridge to another bridge.", MsgPrefix)
	return nil
}
func (s *L2Bridge) RemoveFromBridge() error {
	s.log.Error("%s: There are no able to remove the bridge from another bridge.", MsgPrefix)
	return nil
}

func NewBridge() NpOperator {
	rv := new(L2Bridge)
	rv.setupGlobals()
	return rv
}

// -----------------------------------------------------------------------------

type L2Bond struct {
	OpBase
}

func (s *L2Bond) Create(dryrun bool) (err error) {
	if dryrun {
		s.log.Info("%s dryrun: Bond '%s' created.", MsgPrefix, s.Name())
		return nil
	}

	s.log.Info("%s Creating bond '%s'", MsgPrefix, s.Name())
	bnd := netlink.NewLinkBond(netlink.LinkAttrs{Name: s.Name()})
	if err = s.handle.LinkAdd(bnd); err != nil {
		s.log.Error("%s: error while bond creating: %v", MsgPrefix, err)
		return err
	} else {
		s.log.Info("%s: bond created.", MsgPrefix)
	}

	err = s.Modify(false)

	return err
}

func (s *L2Bond) Remove(dryrun bool) (err error) {
	if dryrun {
		s.log.Info("%s: dryrun: Bond '%s' removed.", MsgPrefix, s.Name())
		return nil
	}
	s.log.Info("%s: Removing Bond '%s'", MsgPrefix, s.Name())
	link, _ := netlink.LinkByName(s.Name())
	if err = s.handle.LinkSetDown(link); err != nil {
		s.log.Error("%s: error while Bond removing: %v", MsgPrefix, err)
		return err
	}
	if err = s.handle.LinkDel(link); err != nil {
		s.log.Error("%s: error while Bond removing: %v", MsgPrefix, err)
	} else {
		s.log.Info("%s: Bond removed.", MsgPrefix)
	}
	return err
}

func (s *L2Bond) getSlaves(dryrun bool) (rv []string) {
	if dryrun {
		rv = []string{"eth1", "eth2"}
		s.log.Info("%s dryrun: slaves for Bond '%s' is %v", MsgPrefix, s.Name(), rv)
		return
	}

	// todo(sv): Research why this code does not work properly (race condition, may be)
	// bondIfIndex := s.Link().Attrs().Index

	// linkList, err := netlink.LinkList()
	// if err != nil {
	// 	s.log.Error("%v", err)
	// 	return
	// }
	// for _, link := range linkList {
	// 	linkAttrs := link.Attrs()
	// 	s.log.Info("%s: FFF name='%s' master='%d'", MsgPrefix, linkAttrs.Name, linkAttrs.MasterIndex)
	// 	if linkAttrs.MasterIndex == bondIfIndex {
	// 		rv = append(rv, linkAttrs.Name)
	// 	}
	// }

	bondSlavesFileName := fmt.Sprintf("/sys/class/net/%s/bonding/slaves", s.Name())
	var (
		rr   *os.File
		err  error
		data []byte
	)
	// Log.Debug("Run NetworkConfig with network scheme: '%s'", c.GlobalString("ns"))
	if rr, err = os.Open(bondSlavesFileName); err != nil {
		s.log.Error("Can't open file '%s': %v", bondSlavesFileName, err)
		return
	}
	if data, err = ioutil.ReadAll(rr); err != nil {
		s.log.Error("Can't process file '%s': %v", bondSlavesFileName, err)
		return
	}
	rv = strings.Fields(string(data))
	sort.Strings(rv)
	return rv
}

// diffSlaves -- genrate diff between current and wanted Bond slaves list
// returns (diff, ok), where ok==true if differnces found
func (s *L2Bond) diffSlaves(dryrun bool, wantedSlaves []string) (*BondSlavesDiffType, bool) {
	actualSlaves := s.getSlaves(dryrun) // already sorted by design
	sort.Strings(wantedSlaves)          // should be sorted
	s.log.Debug("%s: Bond's wanted slaves: %v", MsgPrefix, wantedSlaves)
	s.log.Debug("%s: Bond's actual slaves: %v", MsgPrefix, actualSlaves)

	ok := false
	rv := &BondSlavesDiffType{}

	// check for new addition
	for _, ifname := range wantedSlaves {
		if n := IndexString(actualSlaves, ifname); n == -1 {
			rv.toAdd = append(rv.toAdd, ifname)
			ok = true
		}
	}

	// check for unwanted member
	for _, ifname := range actualSlaves {
		if n := IndexString(wantedSlaves, ifname); n == -1 {
			rv.toRemove = append(rv.toRemove, ifname)
			ok = true
		}
	}
	return rv, ok
}

func (s *L2Bond) Modify(dryrun bool) (err error) {
	if dryrun {
		s.log.Info("%s dryrun: Bond '%s' modifyed.", MsgPrefix, s.Name())
		return nil
	}

	s.log.Info("%s: Modifying Bond '%s'", MsgPrefix, s.Name())
	bondLink := s.Link()
	bondAttrs := bondLink.Attrs()

	// if err = s.handle.LinkSetDown(bondLink); err != nil {
	// 	s.log.Error("%s: error while Bond set to DOWN state: %v", MsgPrefix, err)
	// }

	if s.wantedState.L2.Mtu > 0 && s.wantedState.L2.Mtu != bondAttrs.MTU {
		s.log.Debug("%s: setting MTU to: %v", MsgPrefix, s.wantedState.L2.Mtu)
		if err = s.handle.LinkSetMTU(bondLink, s.wantedState.L2.Mtu); err != nil {
			s.log.Error("%s: error while Bond set MTU: %v", MsgPrefix, err)
		}
	}

	diff, need := s.diffSlaves(dryrun, s.wantedState.L2.Slaves)
	if need {
		for _, slaveName := range diff.toRemove {
			s.log.Debug("%s: Removing '%s' from bond", MsgPrefix, slaveName)
			slaveLink, err := s.handle.LinkByName(slaveName)
			if err == nil {
				err = s.handle.LinkSetDown(slaveLink)
			}
			if err == nil {
				err = s.handle.LinkSetNoMaster(slaveLink)
			}
			if err != nil {
				s.log.Error("%s: error while Bond removing slave '%s': %v", MsgPrefix, slaveName, err)
			}
		}
		for _, slaveName := range diff.toAdd {
			s.log.Debug("%s: Enslaving '%s' to bond", MsgPrefix, slaveName)
			slaveLink, err := s.handle.LinkByName(slaveName)
			if err == nil {
				err = s.handle.LinkSetDown(slaveLink)
			}
			if err == nil {
				err = netlink.LinkSetBondSlave(slaveLink, &netlink.Bond{LinkAttrs: *bondAttrs})
				// whi does not wokr??? err = s.handle.LinkSetMasterByIndex(slaveLink, bondAttrs.MasterIndex)
			}
			if err != nil {
				s.log.Error("%s: error while Bond adding slave '%s': %v", MsgPrefix, slaveName, err)
			}
		}
	}

	if s.wantedState.L2.Bridge != "" {
		s.AddToBridge(s.wantedState.L2.Bridge)
	} else {
		s.RemoveFromBridge()
	}

	if s.wantedState.Online {
		s.log.Debug("%s: setting to UP state", MsgPrefix)
		if err = s.handle.LinkSetUp(bondLink); err != nil {
			s.log.Error("%s: error while Bond set to UP state: %v", MsgPrefix, err)
		}
	}

	s.allignIPv4list()

	return err
}

func NewBond() NpOperator {
	rv := new(L2Bond)
	rv.setupGlobals()
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
