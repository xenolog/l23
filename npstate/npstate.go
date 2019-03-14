package npstate

import (
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
	LinkType string
	Provider string
	Online   bool
	L2       L2State
	L3       L3State
}

// // Next methods implements netlink.NP interface
func (s *NPState) Attrs() *netlink.LinkAttrs {
	return s.attrs
}

func (s *NPState) Type() string {
	return s.LinkType
}

func (s *NPState) Master() int {
	return s.attrs.MasterIndex
}

func (s *NPState) Parent() int {
	return s.attrs.ParentIndex
}

func (s *NPState) CacheAttrs(a *netlink.LinkAttrs) {
	s.attrs = a
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
	DefaultProvider string
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
			Log.Debug("*** Comparing '%s':\n%v \n%v", np, n.NP[key])
			rv.Different = append(rv.Different, key)
		}
	}

	return rv
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
