package npstate

import (
	"reflect"
	"sort"

	"github.com/vishvananda/netlink"
	logger "github.com/xenolog/go-tiny-logger"
	yaml "gopkg.in/yaml.v3"
)

var Log *logger.Logger

type L2State struct {
	Mtu    int
	Bridge string
	Parent string
	Slaves []string
	VlanID int
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

// CompareL2 -- A method, allows to compare L2 properties of NetworkPrimitive
func (s *NPState) CompareL2(n *NPState) bool {
	// fmt.Printf("*** Comparing L2 '%s' and '%s':\n", s.Name, n.Name)
	// sl2, _ := yaml.Marshal(s.L2)
	// sn2, _ := yaml.Marshal(n.L2)
	// fmt.Printf("*** L2:\n%s\n%s\n", sl2, sn2)
	rv := reflect.DeepEqual(s.L2, n.L2)
	// fmt.Printf(">>> %v\n", rv)
	return rv
}

// CompareL3 -- A method, allows to compare L3 properties of NetworkPrimitive
func (s *NPState) CompareL3(n *NPState) bool {
	// fmt.Printf("*** Comparing L3 '%s' and '%s':\n", s.Name, n.Name)
	// I do not known why comparing xxx.L3.something by DeepEq are failed for equal lists.
	// such comparing are works
	s4 := make([]string, len(s.L3.IPv4))
	copy(s4, s.L3.IPv4)
	sort.Strings(s4)
	n4 := make([]string, len(n.L3.IPv4))
	copy(n4, n.L3.IPv4)
	sort.Strings(n4)
	rv := reflect.DeepEqual(s4, n4)
	// sl3, _ := yaml.Marshal(s4)
	// sn2, _ := yaml.Marshal(n4)
	// fmt.Printf("*** L3:\n%s\n%s\n", sl3, sn2)
	// fmt.Printf(">>> %v\n", rv)
	return rv
}

// CompareL23 -- A method, allows to compare L2 and L3 Properties together of
// NetworkPrimitive
func (s *NPState) CompareL23(n *NPState) bool {
	l2 := s.CompareL2(n)
	l3 := s.CompareL3(n)
	oo := (s.Online == n.Online)
	//fmt.Printf("*** '%s-%s': %v %v %v\n", s.Name, n.Name, l2, l3, oo)
	return l2 && l3 && oo
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

type NPStates map[string]*NPState
type TopologyState struct {
	NP              NPStates
	Order           []string
	DefaultProvider string
}

// Compare -- compare TopologyState with another
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
		} else if !np.CompareL23(n.NP[key]) {
			// fmt.Printf("CMP result>>> %v", np.CompareL23(n.NP[key]))
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
