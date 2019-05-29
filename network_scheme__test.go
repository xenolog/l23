package main

import (
	"reflect"
	"strings"
	"testing"

	npstate "github.com/xenolog/l23/npstate"
	yaml "gopkg.in/yaml.v3"
)

func NetworkScheme_1() string {
	return `
version: 1.2
provider: qwe
xxx: yyy
interfaces:
    eth0:
      mtu: 2048
    eth1:
      mtu: 999
      provider: xxx
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

func TopologyState_1() *npstate.TopologyState {
	var linkName string
	rv := &npstate.TopologyState{
		NP: make(map[string]*npstate.NPState),
	}

	linkName = "eth0"
	rv.NP[linkName] = &npstate.NPState{
		Name:   linkName,
		Action: "port",
		Online: true,
		L2: npstate.L2State{
			Mtu: 2048,
		},
		L3: npstate.L3State{
			IPv4: []string{"10.1.3.11/24", "10.20.30.40/24"},
		},
		Provider: "qwe",
	}

	linkName = "eth1"
	rv.NP[linkName] = &npstate.NPState{
		Name:   linkName,
		Action: "port",
		Online: true,
		L2: npstate.L2State{
			Mtu: 999,
		},
		L3: npstate.L3State{
			IPv4: nil,
		},
		Provider: "xxx",
	}

	linkName = "eth2"
	rv.NP[linkName] = &npstate.NPState{
		Name:     linkName,
		Action:   "port",
		Online:   true,
		Provider: "qwe",
	}
	return rv
}

// -----------------------------------------------------------------------------

func NetworkScheme_2() string {
	// global Provider value should be empty !!!
	return `
version: 1.1
xxx: yyy
interfaces:
    eth0: {}
    eth1:
      mtu: 999
    eth2: {}
transformations:
  - name: eth0
    action: port
    mtu: 9000
  - name: eth1
    action: port
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

func TopologyState_2() *npstate.TopologyState {
	var linkName string
	rv := &npstate.TopologyState{
		NP: make(map[string]*npstate.NPState),
	}

	linkName = "eth0"
	rv.NP[linkName] = &npstate.NPState{
		Name:   linkName,
		Action: "port",
		Online: true,
		L2: npstate.L2State{
			Mtu: 9000,
		},
		L3: npstate.L3State{
			IPv4: []string{"10.1.3.11/24", "10.20.30.40/24"},
		},
		Provider: "lnx",
	}

	linkName = "eth1"
	rv.NP[linkName] = &npstate.NPState{
		Name:   linkName,
		Action: "port",
		Online: true,
		L2: npstate.L2State{
			Mtu: 2048,
		},
		L3: npstate.L3State{
			IPv4: nil,
		},
		Provider: "lnx",
	}

	linkName = "eth2"
	rv.NP[linkName] = &npstate.NPState{
		Name:     linkName,
		Action:   "port",
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
	if ns.Provider != "qwe" {
		t.Logf("Provider mismatch. Actual:'%s'", ns.Provider)
		t.Fail()
	}
	if ns.Interfaces["eth0"].Mtu != 2048 {
		t.Logf("eth0 MTU mismatch")
		t.Fail()
	}
	if ns.Endpoints["eth0"].Gateway != "10.1.3.1" {
		t.Logf("eth0 gateway mismatch")
		t.Fail()
	}
	if ns.Endpoints["eth0"].IP[1] != "10.20.30.40/24" {
		t.Logf("eth0 2nd IPaddr mismatch")
		t.Fail()
	}
	if ns.Endpoints["eth1"].IP[0] != "none" {
		t.Logf("eth1 IPaddr is not none")
		t.Fail()
	}
}

// -----------------------------------------------------------------------------

func TestNS__GenerateTopologyState(t *testing.T) {
	ns := new(NetworkScheme)
	ns_yaml := strings.NewReader(NetworkScheme_1())
	if err := ns.Load(ns_yaml); err != nil {
		t.FailNow()
	}
	nps := ns.TopologyState()
	ifOrder := []string{"eth0", "eth1", "eth2"}
	if !reflect.DeepEqual(nps.Order, ifOrder) {
		t.Logf("Wrong ordering: %v, instead %v", nps.Order, ifOrder)
		t.Fail()
	}
	wantedNps := TopologyState_1()
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
	nps := ns.TopologyState()
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
	nps := ns.TopologyState()
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
	nps := ns.TopologyState()
	for _, m := range [][]string{
		{"eth0", "aaa"}, // provider defined into interfaces section
		{"eth1", "bbb"}, // provider defined into interfaces section and re-defined into transformations
		{"eth2", "ccc"}, // provider defined into transformation
		{"eth3", "ovs"}, // todo(sv): should be 'ddd' provider defined into interfaces section and does not touched into transformations
		{"eth4", "ovs"}, // Default Provider used
	} {
		if nps.NP[m[0]].Provider != m[1] {
			t.Logf("Wrong provider for %s: %v, instead %v", m[0], nps.NP[m[0]].Provider, m[1])
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
	nps := ns.TopologyState()
	wantedNps := TopologyState_2()
	diff := nps.Compare(wantedNps)
	if !diff.IsEqual() {
		t.Logf("Incoming NetworkScheme: %v", ns_yaml)
		t.Logf("NS => NPS:\n%s", nps)
		t.Logf("Wanted NPS:\n%s", wantedNps)
		t.Logf("Diff:\n%s", diff)
		t.Fail()
	}
}

func TestNS__TransformationsOrder_1(t *testing.T) {
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
	nps := ns.TopologyState()
	wantedOrder := []string{"eth0", "eth0.101", "eth1"}
	if !reflect.DeepEqual(nps.Order, wantedOrder) {
		t.Logf("Wrong ordering: %v, instead %v", nps.Order, wantedOrder)
		t.Fail()
	}
}

func TestNS__TransformationsOrder_2(t *testing.T) {
	ns := new(NetworkScheme)
	ns_data := strings.NewReader(`
version: 1.1
provider: lnx
interfaces:
  eth0: {}
  eth1: {}
transformations:
  - name: br0
    action: bridge
  - name: eth0
    action: port
    bridge: br0
  - name: eth0.101
    action: port
`)
	if err := ns.Load(ns_data); err != nil {
		t.FailNow()
	}
	nps := ns.TopologyState()
	// eth1 before eth0 , because eth1 not listed into transformations
	wantedOrder := []string{"eth1", "br0", "eth0", "eth0.101"}
	if !reflect.DeepEqual(nps.Order, wantedOrder) {
		t.Logf("Wrong ordering: %v, instead %v", nps.Order, wantedOrder)
		t.Fail()
	}
}

func TestNS__Transformations__L2fields(t *testing.T) {
	ns := new(NetworkScheme)
	ns_data := strings.NewReader(`
version: 1.1
provider: lnx
interfaces:
    eth0: {}
transformations:
  - name: xxx
    action: fuck
    mtu: 9000
    bridge: aaa
    parent: eth0
    slaves:
      - x1
      - x2
    vlan_id: 101
`)
	if err := ns.Load(ns_data); err != nil {
		t.FailNow()
	}
	nps := ns.TopologyState()
	//todo(sv): True order should be

	if nps.NP["xxx"].Name != "xxx" {
		t.Logf("Field 'Name' not present or invalid")
		t.Fail()
	}
	if nps.NP["xxx"].Action != "fuck" {
		t.Logf("Field 'Action' not present or invalid")
		t.Fail()
	}
	if nps.NP["xxx"].L2.Mtu != 9000 {
		t.Logf("Field 'Mtu' not present or invalid")
		t.Fail()
	}
	if nps.NP["xxx"].L2.Bridge != "aaa" {
		t.Logf("Field 'Bridge' not present or invalid")
		t.Fail()
	}
	if nps.NP["xxx"].L2.Parent != "eth0" {
		t.Logf("Field 'Parent' not present or invalid")
		t.Fail()
	}
	if !reflect.DeepEqual(nps.NP["xxx"].L2.Slaves, []string{"x1", "x2"}) {
		t.Logf("Field 'Slaves' not present or invalid")
		t.Fail()
	}
	if t.Failed() {
		txt, _ := yaml.Marshal(nps.NP["xxx"])
		t.Logf(string(txt))
	}
}

// -----------------------------------------------------------------------------
func FlxNetworkScheme_1() string {
	return `
version: 1.2
provider: lnx
config-provider: ubuntu16
processing: imperative
interfaces:
  eth0: {}
transformations:
- name: br1
  mtu:  1496
  action: bridge
  vendor_specific:
    stp: false
- name: br2
  action: bridge
  vendor_specific:
    aaa: bbb
    ccc: 123
    ddd: eee
`
}

func GetPluginCustomProperties() PluginCustomProperties {
	rv := make(PluginCustomProperties)
	rv["interface"] = make(CustomProperties)
	rv["interface"]["provider"] = CustomProperty{
		// 'interface' -- is a reserved name for RAW interfaces
		Type:         "string",
		DefaultValue: "lnx",
	}
	rv["bridge"] = make(CustomProperties)
	rv["bridge"]["stp"] = CustomProperty{
		Type:         "bool",
		DefaultValue: "false",
	}
	rv["bridge"]["aaa"] = CustomProperty{
		Type:         "string",
		DefaultValue: "aaa",
	}
	rv["bridge"]["ccc"] = CustomProperty{
		Type:         "int",
		DefaultValue: "1",
	}
	return rv
}

func Test_FlxNS__CustomValues(t *testing.T) {
	ns := new(NetworkScheme)
	ns_yaml := strings.NewReader(FlxNetworkScheme_1())
	if err := ns.Load(ns_yaml); err != nil {
		t.FailNow()
	}
	if err := ns.ProcessVS(GetPluginCustomProperties()); err != nil {
		t.FailNow()
	}

	// if ns.Interfaces["eth0"].Mtu != 2048 {
	// 	t.Fail()
	// }
	// if ns.Endpoints["eth0"].Gateway != "10.1.3.1" {
	// 	t.Fail()
	// }
	// if ns.Endpoints["eth0"].IP[1] != "10.20.30.40/24" {
	// 	t.Fail()
	// }
	// if ns.Endpoints["eth1"].IP[0] != "none" {
	// 	t.Fail()
	// }
}

// -----------------------------------------------------------------------------
