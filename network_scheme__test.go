package main

import (
	"gopkg.in/yaml.v2"
	"strings"
	"testing"
)

func NetworkScheme_1() string {
	return `
version: 1.1
provider: lnx
xxx: yyy
interfaces:
    eth0:
      mtu: 2048
    eth1:
      mtu: 999
    eth2: {}
endpoints:
    eth0:
      gateway: 10.1.3.1
      IP:
        - '10.1.3.11/24'
        - '10.20.30.40/24'
    eth1:
      IP:
        - none
aaa: bbb
`
}

func TestNetworkSchemeLoad(t *testing.T) {
	t.Logf("Incoming NetworkScheme: %v", NetworkScheme_1())
	ns := new(NetworkScheme)
	ns_yaml := strings.NewReader(NetworkScheme_1())
	if err := ns.Load(ns_yaml); err != nil {
		t.FailNow()
	}
	outyaml, _ := yaml.Marshal(ns)
	t.Logf("Effective NetworkScheme: \n%s", outyaml)
}

func TestNetworkSchemeValues(t *testing.T) {
	ns := new(NetworkScheme)
	ns_yaml := strings.NewReader(NetworkScheme_1())
	if err := ns.Load(ns_yaml); err != nil {
		t.FailNow()
	}
	if ns.Interfaces["eth0"].Mtu != 2048 {
		t.Fail()
	}
	if ns.Endpoints["eth0"].Gateway != "10.1.3.1" {
		t.Fail()
	}
	if ns.Endpoints["eth0"].IP[1] != "10.20.30.40/24" {
		t.Fail()
	}
	if ns.Endpoints["eth1"].IP[0] != "none" {
		t.Fail()
	}
}
