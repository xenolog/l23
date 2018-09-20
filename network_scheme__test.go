package main

import (
	"reflect"
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

// -----------------------------------------------------------------------------

func NetworkScheme_2() string {
	return `
version: 1.1
provider: lnx
xxx: yyy
interfaces:
    eth0: {}
    eth1:
      mtu: 999
    eth2: {}
transformations:
  - name: eth0
    mtu: 9000
  - name: eth1
    mtu: 2048
    vendor_specific:
      aaa: bbb
      ccc:
        - c1
        - c2
        - c3
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

func NpsStatus_2() *NpsStatus {
	var linkName string
	rv := &NpsStatus{
		Link: make(map[string]*NpLinkStatus),
	}

	linkName = "eth0"
	rv.Link[linkName] = &NpLinkStatus{
		Name:   linkName,
		Online: true,
		L2: L2Status{
			MTU: 9000,
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
			MTU: 2048,
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

// -----------------------------------------------------------------------------

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

// -----------------------------------------------------------------------------

func TestNS__GenerateNpsStatus(t *testing.T) {
	ns := new(NetworkScheme)
	ns_yaml := strings.NewReader(NetworkScheme_1())
	if err := ns.Load(ns_yaml); err != nil {
		t.FailNow()
	}
	nps := ns.NpsStatus()
	ifOrder := []string{"eth0", "eth5", "eth2"}
	if !reflect.DeepEqual(nps.Order, ifOrder) {
		t.Logf("Wrong ordering: %v, instead %v", nps.Order, ifOrder)
		t.Fail()
	}
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

// -----------------------------------------------------------------------------

func TestNS__TransformationsLoad(t *testing.T) {
	t.Logf("Incoming NetworkScheme: %v", NetworkScheme_2())
	ns := new(NetworkScheme)
	ns_yaml := strings.NewReader(NetworkScheme_2())
	if err := ns.Load(ns_yaml); err != nil {
		t.FailNow()
	}
	outyaml, _ := yaml.Marshal(ns)
	t.Logf("Effective NetworkScheme: \n%s", outyaml)
}

func TestNS__TransformationsValues(t *testing.T) {
	ns_yaml := NetworkScheme_2()
	ns := new(NetworkScheme)
	ns_data := strings.NewReader(ns_yaml)
	if err := ns.Load(ns_data); err != nil {
		t.FailNow()
	}
	nps := ns.NpsStatus()
	wantedNps := NpsStatus_2()
	diff := nps.Compare(wantedNps)
	if !diff.IsEqual() {
		t.Logf("Incoming NetworkScheme: %v", ns_yaml)
		t.Logf("NS => NPS:\n%s", nps)
		t.Logf("Wanted NPS:\n%s", wantedNps)
		t.Logf("Diff:\n%s", diff)
		t.Fail()
	}
}

func TestNS__TransformationsOrder(t *testing.T) {
	ns := new(NetworkScheme)
	ns_data := strings.NewReader(`
version: 1.1
provider: lnx
interfaces:
    eth0: {}
    eth1: {}
transformations:
  - name: eth0
    mtu: 9000
  - name: eth0.101
    mtu: 2048
  - name: eth1
    mtu: 9000
`)
	if err := ns.Load(ns_data); err != nil {
		t.FailNow()
	}
	nps := ns.NpsStatus()
	//todo(sv): True order should be
	//wantedOrder := []string{"eth0", "eth0.101", "eth1"}
	// because eth1 found in the transformations list
	wantedOrder := []string{"eth0", "eth1", "eth0.101"}
	if !reflect.DeepEqual(nps.Order, wantedOrder) {
		t.Logf("Wrong ordering: %v, instead %v", nps.Order, wantedOrder)
		t.Fail()
	}
}

// -----------------------------------------------------------------------------
