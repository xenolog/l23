package main

import (
	// "gopkg.in/yaml.v2"
	// "strings"
	"github.com/vishvananda/netlink"

	"net"
	"testing"
)

func RuntimeNpStatuses() *NpsStatus {
	var linkName string
	rv := &NpsStatus{
		Link: make(map[string]*NpLinkStatus),
	}

	linkName = "lo"
	rv.Link[linkName] = &NpLinkStatus{
		Name:   &linkName,
		Online: true,
		L3: L3Status{
			IPv4: []netlink.Addr{
				netlink.Addr{
					IPNet: &net.IPNet{
						IP:   net.ParseIP("127.0.0.1"),
						Mask: net.CIDRMask(8, 32),
					},
				},
			},
		},
	}

	linkName = "eth1"
	rv.Link[linkName] = &NpLinkStatus{
		Name:   &linkName,
		Online: true,
		L3: L3Status{
			IPv4: []netlink.Addr{
				netlink.Addr{
					IPNet: &net.IPNet{
						IP:   net.ParseIP("10.20.30.40"),
						Mask: net.CIDRMask(24, 32),
					},
				},
				netlink.Addr{
					IPNet: &net.IPNet{
						IP:   net.ParseIP("20.30.40.50"),
						Mask: net.CIDRMask(25, 32),
					},
				},
			},
		},
	}
	// rv.Link[linkName].L3.IPv4[0].IPNet = &net.IPNet{
	// 	IP:   net.ParseIP("10.20.30.40"),
	// 	Mask: net.CIDRMask(24, 32),
	// }
	// rv.Link[linkName].L3.IPv4[1].IPNet = &net.IPNet{
	// 	IP:   net.ParseIP("20.30.40.50"),
	// 	Mask: net.CIDRMask(25, 32),
	// }

	return rv
}

func TestEqualNpStatuses(t *testing.T) {
	t.Logf("Incoming NPS: %v", RuntimeNpStatuses())
	// ns := new(NetworkScheme)
	// ns_yaml := strings.NewReader(NetworkScheme_1())
	// if err := ns.Load(ns_yaml); err != nil {
	// 	t.FailNow()
	// }
	// outyaml, _ := yaml.Marshal(ns)
	// t.Logf("Effective NetworkScheme: \n%s", outyaml)
}
