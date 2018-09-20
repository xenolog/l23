package main

import (
	"net"
	"reflect"

	"github.com/vishvananda/netlink"
	yaml "gopkg.in/yaml.v2"
)

type L2Status struct {
	MTU int
}

type L3Status struct {
	IPv4 []string // in the CIDR notation
	// IPv6 []IpAddr6
}

// Np -- is a acronym for Network Primitive
type NpLinkStatus struct {
	Name     string
	IfIndex  int
	attrs    *netlink.LinkAttrs
	linkType string
	Provider string
	Online   bool
	L2       L2Status
	L3       L3Status
}

// Fill LinkStatus data structure by LinkAttrs
func (s *NpLinkStatus) FillByNetlinkLink(link netlink.Link) {
	s.attrs = link.Attrs()
	s.linkType = link.Type()
	s.Name = s.attrs.Name
	s.IfIndex = s.attrs.Index
	if s.attrs.Flags&net.FlagUp != 0 {
		s.Online = true
	}
	s.fillL2statusByNetlinkLink()
}

func (s *NpLinkStatus) fillL2statusByNetlinkLink() {
	s.L2.MTU = s.attrs.MTU
}

func (s *NpLinkStatus) FillByNetlinkAddrList(addrs *[]netlink.Addr) {
	s.L3.IPv4 = make([]string, len(*addrs))
	for _, addr := range *addrs {
		s.L3.IPv4 = append(s.L3.IPv4, addr.IPNet.String())
	}
}

// Next methods implements netlink.Link interface
func (s *NpLinkStatus) Attrs() *netlink.LinkAttrs {
	return s.attrs
}
func (s *NpLinkStatus) Type() string {
	return s.linkType
}

func (s *NpLinkStatus) String() string {
	rv, _ := yaml.Marshal(s)
	return string(rv)
}

//------------------------------------------------------------------------------

type DiffNpsStatuses struct {
	New       []string
	Waste     []string
	Different []string
}

func (s *DiffNpsStatuses) IsEqual() bool {
	return len(s.New) == 0 && len(s.Waste) == 0 && len(s.Different) == 0
}
func (s *DiffNpsStatuses) String() string {
	rv, _ := yaml.Marshal(s)
	return string(rv)
}

//------------------------------------------------------------------------------

type NpsStatus struct {
	Link            map[string]*NpLinkStatus
	Order           []string
	DefaultProvider string
	handle          *netlink.Handle
}

// This method allow to compare NpsStatus with another
// NpsStatus (runtime and wanted, for example)
// and return report about diferences
func (s *NpsStatus) Compare(n *NpsStatus) *DiffNpsStatuses {
	rv := new(DiffNpsStatuses)

	// check for aded Np
	for key, _ := range n.Link {
		if _, ok := s.Link[key]; !ok {
			rv.New = append(rv.New, key)
		}
	}

	// check for different and removed Np
	for key, np := range s.Link {
		if _, ok := n.Link[key]; !ok {
			rv.Waste = append(rv.Waste, key)
			continue
		}
		if !reflect.DeepEqual(np, n.Link[key]) {
			rv.Different = append(rv.Different, key)
		}
	}

	return rv
}

// Setup netlink handler if need.
// set args[0] to true for force re-init handler
func (s *NpsStatus) setHandle(args ...bool) (err error) {
	need := false
	if len(args) > 0 && args[0] == true {
		need = true
	} else {
		if s.handle == nil {
			need = true
		}
	}

	if need {
		if s.handle, err = netlink.NewHandle(); err != nil {
			Log.Error("%v", err)
			return
		}
	}

	return nil
}

func (s *NpsStatus) ObserveRuntime() (err error) {
	var linkList []netlink.Link

	s.setHandle()

	if linkList, err = s.handle.LinkList(); err != nil {
		Log.Error("%v", err)
		return
	}

	//links := reflect.ValueOf(linkList).MapKeys()
	for _, link := range linkList {
		linkName := link.Attrs().Name
		s.Link[linkName] = new(NpLinkStatus)
		s.Link[linkName].FillByNetlinkLink(link)
		if ipaddrInfo, err := s.handle.AddrList(link, netlink.FAMILY_V4); err == nil {
			s.Link[linkName].FillByNetlinkAddrList(&ipaddrInfo)
		} else {
			Log.Error("Error while fetch L3 info for '%s' %v", linkName, err)
		}
	}
	return nil
}

func (s *NpsStatus) String() string {
	rv, _ := yaml.Marshal(s)
	return string(rv)
}

func NewNpsStatus() *NpsStatus {
	rv := new(NpsStatus)
	rv.Link = make(map[string]*NpLinkStatus)
	return rv
}
