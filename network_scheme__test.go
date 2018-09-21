package l23

import (
	"reflect"
	"strings"
	"testing"

	ifstatus "github.com/xenolog/l23/ifstatus"
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

func NpsStatus_1() *ifstatus.NpsStatus {
	var linkName string
	rv := &ifstatus.NpsStatus{
		Link: make(map[string]*ifstatus.NpLinkStatus),
	}

	linkName = "eth0"
	rv.Link[linkName] = &ifstatus.NpLinkStatus{
		Name:   linkName,
		Online: true,
		L2: ifstatus.L2Status{
			MTU: 2048,
		},
		L3: ifstatus.L3Status{
			IPv4: []string{"10.1.3.11/24", "10.20.30.40/24"},
		},
		Provider: "lnx",
	}

	linkName = "eth1"
	rv.Link[linkName] = &ifstatus.NpLinkStatus{
		Name:   linkName,
		Online: true,
		L2: ifstatus.L2Status{
			MTU: 999,
		},
		L3: ifstatus.L3Status{
			IPv4: nil,
		},
		Provider: "lnx",
	}

	linkName = "eth2"
	rv.Link[linkName] = &ifstatus.NpLinkStatus{
		Name:     linkName,
		Online:   true,
		Provider: "lnx",
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

func NpsStatus_2() *ifstatus.NpsStatus {
	var linkName string
	rv := &ifstatus.NpsStatus{
		Link: make(map[string]*ifstatus.NpLinkStatus),
	}

	linkName = "eth0"
	rv.Link[linkName] = &ifstatus.NpLinkStatus{
		Name:   linkName,
		Online: true,
		L2: ifstatus.L2Status{
			MTU: 9000,
		},
		L3: ifstatus.L3Status{
			IPv4: []string{"10.1.3.11/24", "10.20.30.40/24"},
		},
		Provider: "lnx",
	}

	linkName = "eth1"
	rv.Link[linkName] = &ifstatus.NpLinkStatus{
		Name:   linkName,
		Online: true,
		L2: ifstatus.L2Status{
			MTU: 2048,
		},
		L3: ifstatus.L3Status{
			IPv4: nil,
		},
		Provider: "lnx",
	}

	linkName = "eth2"
	rv.Link[linkName] = &ifstatus.NpLinkStatus{
		Name:     linkName,
		Online:   true,
		Provider: "lnx",
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
	ifOrder := []string{"eth0", "eth1", "eth2"}
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

func TestNS__DefaultProviderEmpty(t *testing.T) {
	ns := new(NetworkScheme)
	ns_data := strings.NewReader(`
version: 1.1
interfaces:
    eth0: {}
`)
	if err := ns.Load(ns_data); err != nil {
		t.FailNow()
	}
	nps := ns.NpsStatus()
	if nps.DefaultProvider != "lnx" {
		t.Logf("Wrong default provider: %v, instead %v", nps.DefaultProvider, "lnx")
		t.Fail()
	}
}

func TestNS__DefaultProviderSetup(t *testing.T) {
	ns := new(NetworkScheme)
	ns_data := strings.NewReader(`
version: 1.1
provider: ovs
interfaces:
    eth0: {}
`)
	if err := ns.Load(ns_data); err != nil {
		t.FailNow()
	}
	nps := ns.NpsStatus()
	if nps.DefaultProvider != "ovs" {
		t.Logf("Wrong default provider: %v, instead %v", nps.DefaultProvider, "lnx")
		t.Fail()
	}
}

func TestNS__DefaultProviderWithMoreSpecific(t *testing.T) {
	ns := new(NetworkScheme)
	ns_data := strings.NewReader(`
version: 1.1
provider: ovs
interfaces:
    eth0:
      provider: aaa
    eth1:
      provider: aaa
    eth2: {}
    eth3:
      provider: ddd
    eth4: {}
transformations:
    - action: port
      name: eth1
      provider: bbb
    - action: port
      name: eth2
      provider: ccc
    - action: port
      name: eth3
`)
	if err := ns.Load(ns_data); err != nil {
		t.FailNow()
	}
	nps := ns.NpsStatus()
	for _, m := range [][]string{
		{"eth0", "aaa"}, // provider defined into interfaces section
		{"eth1", "bbb"}, // provider defined into interfaces section and re-defined into transformations
		{"eth2", "ccc"}, // provider defined into transformation
		{"eth3", "ddd"}, // provider defined into interfaces section and does not touched into transformations
		{"eth4", "ovs"}, // Default Provider used
	} {
		if nps.Link[m[0]].Provider != m[1] {
			t.Logf("Wrong provider for %s: %v, instead %v", m[0], nps.Link[m[0]].Provider, m[1])
			t.Fail()
		}
	}
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
