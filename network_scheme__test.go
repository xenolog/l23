package main

import (
	"strings"
	"testing"

	yaml "gopkg.in/yaml.v2"
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
func NpsStatus_1() *NpsStatus {
	var linkName string
	rv := &NpsStatus{
		Link: make(map[string]*NpLinkStatus),
	}

	linkName = "eth0"
	rv.Link[linkName] = &NpLinkStatus{
		Name:   linkName,
		Online: true,
		L2: L2Status{
			MTU: 2048,
		},
		L3: L3Status{
			IPv4: []string{"10.1.3.11/24", "10.20.30.40/24"},
		},
	}

	linkName = "eth1"
	rv.Link[linkName] = &NpLinkStatus{
		Name:   linkName,
		Online: true,
		L2: L2Status{
			MTU: 999,
		},
		L3: L3Status{
			IPv4: nil,
		},
	}

	linkName = "eth2"
	rv.Link[linkName] = &NpLinkStatus{
		Name:   linkName,
		Online: true,
	}
	return rv
}

func TestNS__Load(t *testing.T) {
	t.Logf("Incoming NetworkScheme: %v", NetworkScheme_1())
	ns := new(NetworkScheme)
	ns_yaml := strings.NewReader(NetworkScheme_1())
	if err := ns.Load(ns_yaml); err != nil {
		t.FailNow()
	}
	outyaml, _ := yaml.Marshal(ns)
	t.Logf("Effective NetworkScheme: \n%s", outyaml)
}

func TestNS__Values(t *testing.T) {
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

func TestNS__GenerateNpsStatus(t *testing.T) {
	ns := new(NetworkScheme)
	ns_yaml := strings.NewReader(NetworkScheme_1())
	if err := ns.Load(ns_yaml); err != nil {
		t.FailNow()
	}
	nps := ns.NpsStatus()
	wantedNps := NpsStatus_1()
	diff := nps.Compare(wantedNps)
	if !diff.IsEqual() {
		t.Logf("Incoming NetworkScheme: %v", NetworkScheme_1())
		t.Logf("NS => NPS:\n%s", nps)
		t.Logf("Wanted NPS:\n%s", wantedNps)
		t.Logf("Diff:\n%s", diff)
		t.Fail()
	}
}
