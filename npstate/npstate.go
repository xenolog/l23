package npstate

import (
	"net"
	"reflect"

	"github.com/vishvananda/netlink"
	logger "github.com/xenolog/go-tiny-logger"
	yaml "gopkg.in/yaml.v2"
)

var Log *logger.Logger

type L2State struct {
	Mtu          int
	Bridge       string
	Parent       string
	Slaves       []string
	Vlan_id      int
	Stp          bool
	Bpdu_forward bool
	// Type         string
}

type L3State struct {
	IPv4 []string // in the CIDR notation
	// IPv6 []IpAddr6
}

// Np -- is a acronym for Network Primitive
type NPState struct {
	Name     string
	Action   string
	IfIndex  int
	attrs    *netlink.LinkAttrs
	linkType string
	Provider string
	Online   bool
	L2       L2State
	L3       L3State
}

// Fill LinkStatus data structure by LinkAttrs
func (s *NPState) FillByNetlinkLink(link netlink.Link) {
	s.attrs = link.Attrs()
	s.linkType = link.Type()
	s.Name = s.attrs.Name
	s.IfIndex = s.attrs.Index
	if s.attrs.Flags&net.FlagUp != 0 {
		s.Online = true
	}
	s.fillL2stateByNetlinkLink()
}

func (s *NPState) fillL2stateByNetlinkLink() {
	s.L2.Mtu = s.attrs.MTU
}

func (s *NPState) FillByNetlinkAddrList(addrs *[]netlink.Addr) {
	s.L3.IPv4 = make([]string, len(*addrs))
	for _, addr := range *addrs {
		s.L3.IPv4 = append(s.L3.IPv4, addr.IPNet.String())
	}
}

// Next methods implements netlink.NP interface
func (s *NPState) Attrs() *netlink.LinkAttrs {
	return s.attrs
}
func (s *NPState) Type() string {
	return s.linkType
}

func (s *NPState) String() string {
	rv, _ := yaml.Marshal(s)
	return string(rv)
}

//------------------------------------------------------------------------------

type DiffTopologyStatees struct {
	New       []string
	Waste     []string
	Different []string
}

func (s *DiffTopologyStatees) IsEqual() bool {
	return len(s.New) == 0 && len(s.Waste) == 0 && len(s.Different) == 0
}
func (s *DiffTopologyStatees) String() string {
	rv, _ := yaml.Marshal(s)
	return string(rv)
}

//------------------------------------------------------------------------------

type TopologyState struct {
	NP              map[string]*NPState // Should be renamed to NP
	Order           []string
	DefaultProvider string          // Do we really need this field?
	handle          *netlink.Handle // Should be moved to corresponded plugin
}

// This method allow to compare TopologyState with another
// TopologyState (runtime and wanted, for example)
// and return report about diferences
func (s *TopologyState) Compare(n *TopologyState) *DiffTopologyStatees {
	rv := new(DiffTopologyStatees)

	// check for aded Np
	for key, _ := range n.NP {
		if _, ok := s.NP[key]; !ok {
			rv.New = append(rv.New, key)
		}
	}

	// check for different and removed Np
	for key, np := range s.NP {
		if _, ok := n.NP[key]; !ok {
			rv.Waste = append(rv.Waste, key)
		} else if n.NP[key].Action == "remove" {
			// "remove" is a pseudo-action for force add any network primitive to removal queue
			n.NP[key].Action = ""
			rv.Waste = append(rv.Waste, key)
		} else if !reflect.DeepEqual(np, n.NP[key]) {
			rv.Different = append(rv.Different, key)
		}
	}

	return rv
}

// Setup netlink handler if need.
// if nil given -- netlink handler will be created automatically
func (s *TopologyState) setHandle(hh *netlink.Handle) (err error) {
	if s.handle == nil && hh == nil {
		// generate new handle if need
		if s.handle, err = netlink.NewHandle(); err != nil {
			Log.Error("%v", err)
		}
	} else if hh != nil {
		// setup handle
		s.handle = hh
	}
	return
}

func (s *TopologyState) ObserveRuntime() (err error) {
	var linkList []netlink.Link

	s.setHandle(nil)

	if linkList, err = s.handle.LinkList(); err != nil {
		Log.Error("%v", err)
		return
	}

	//links := reflect.ValueOf(linkList).MapKeys()
	for _, link := range linkList {
		linkName := link.Attrs().Name
		s.NP[linkName] = new(NPState)
		s.NP[linkName].FillByNetlinkLink(link)
		if ipaddrInfo, err := s.handle.AddrList(link, netlink.FAMILY_V4); err == nil {
			s.NP[linkName].FillByNetlinkAddrList(&ipaddrInfo)
		} else {
			Log.Error("Error while fetch L3 info for '%s' %v", linkName, err)
		}
	}
	return nil
}

func (s *TopologyState) String() string {
	rv, _ := yaml.Marshal(s)
	return string(rv)
}

func NewTopologyState() *TopologyState {
	rv := new(TopologyState)
	rv.NP = make(map[string]*NPState)
	return rv
}

func init() {
	Log = new(logger.Logger)
}
