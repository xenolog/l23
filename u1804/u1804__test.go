package u1804

import (
	"fmt"
	"testing"

	npstate "github.com/xenolog/l23/npstate"
	"gopkg.in/yaml.v2"

	// logger "github.com/xenolog/go-tiny-logger"
	// . "github.com/xenolog/l23/utils"
	td "github.com/maxatome/go-testdeep"
)

func Test__Just_Ethernet(t *testing.T) {
	wantedState := make(npstate.NPStates)
	linkName := "eth1"
	wantedState[linkName] = &npstate.NPState{
		Name:   linkName,
		Action: "port",
		Online: true,
		L3: npstate.L3State{
			IPv4: []string{"10.10.10.131/25"},
		},
	}

	savedConfig := NewSavedConfig(nil)
	savedConfig.SetWantedState(&wantedState)
	savedConfig.Generate()
	actualYaml := savedConfig.String()
	wantedYaml := `
    version: "2"
    renderer: networkd
    ethernets:
      eth1:
        addresses:
          - 10.10.10.131/25
        dhcp4: false
        dhcp6: false
`
	actualSC := new(SavedConfig)
	if err := yaml.Unmarshal([]byte(actualYaml), actualSC); err != nil {
		t.Logf("Can't unmarshall the actual YAML: %s\n%s", err, actualYaml)
		t.FailNow()
	}
	wantedSC := new(SavedConfig)
	if err := yaml.Unmarshal([]byte(wantedYaml), wantedSC); err != nil {
		t.Logf("Can't unmarshall the wanted YAML: %s\n%s", err, wantedYaml)
		t.FailNow()
	}

	td.CmpDeeply(t, actualSC, wantedSC, "ETH properties are not equal")
}

func Test__Just_Vlan(t *testing.T) {
	wantedState := make(npstate.NPStates)
	linkName := "eth1.101"
	wantedState[linkName] = &npstate.NPState{
		Name:   linkName,
		Action: "port",
		Online: true,
		L2: npstate.L2State{
			Parent:  "eth1",
			Vlan_id: 101,
		},
		L3: npstate.L3State{
			IPv4: []string{"10.10.10.131/25"},
		},
	}

	savedConfig := NewSavedConfig(nil)
	savedConfig.SetWantedState(&wantedState)
	savedConfig.Generate()
	actualYaml := savedConfig.String()
	actualSC := new(SavedConfig)
	if err := yaml.Unmarshal([]byte(actualYaml), actualSC); err != nil {
		t.Logf("Can't unmarshall the actual YAML: %s\n%s", err, actualYaml)
		t.FailNow()
	}
	wantedYaml := `
    version: 2
    renderer: networkd
    ethernets:
      eth1:
        dhcp4: false
        dhcp6: false
    vlans:
      eth1.101:
        id: 101
        link: eth1
        addresses:
          - 10.10.10.131/25
        dhcp4: false
        dhcp6: false
`
	wantedSC := new(SavedConfig)
	if err := yaml.Unmarshal([]byte(wantedYaml), wantedSC); err != nil {
		t.Logf("Can't unmarshall the wanted YAML: %s\n%s", err, wantedYaml)
		t.FailNow()
	}

	td.CmpDeeply(t, actualSC, wantedSC, "ETH and VLAN properties are not equal")
}

func Test__Just_Bridge(t *testing.T) {
	wantedState := make(npstate.NPStates)
	brName := "br1"
	wantedState[brName] = &npstate.NPState{
		Name:   brName,
		Action: "bridge",
		Online: true,
		L3: npstate.L3State{
			IPv4: []string{"10.10.10.131/25"},
		},
	}
	for _, linkName := range []string{"eth1", "eth2", "eth3", "eth4", "eth5"} {
		wantedState[linkName] = &npstate.NPState{
			Name:   linkName,
			Action: "port",
			Online: true,
		}
		if linkName != "eth2" && linkName != "eth4" {
			wantedState[linkName].L2 = npstate.L2State{
				Bridge: brName,
			}
		}
	}

	savedConfig := NewSavedConfig(nil)
	savedConfig.SetWantedState(&wantedState)
	savedConfig.Generate()
	actualYaml := savedConfig.String()
	actualSC := new(SavedConfig)
	if err := yaml.Unmarshal([]byte(actualYaml), actualSC); err != nil {
		t.Logf("Can't unmarshall the actual YAML: %s\n%s", err, actualYaml)
		t.FailNow()
	}
	wantedYaml := `
    version: 2
    renderer: networkd
    ethernets:
      eth1:
        dhcp4: false
        dhcp6: false
      eth2:
        dhcp4: false
        dhcp6: false
      eth3:
        dhcp4: false
        dhcp6: false
      eth4:
        dhcp4: false
        dhcp6: false
      eth5:
        dhcp4: false
        dhcp6: false
    bridges:
      br1:
        interfaces: ["eth1","eth3","eth5"]
        addresses:
          - 10.10.10.131/25
        dhcp4: false
        dhcp6: false
`
	wantedSC := new(SavedConfig)
	if err := yaml.Unmarshal([]byte(wantedYaml), wantedSC); err != nil {
		t.Logf("Can't unmarshall the wanted YAML: %s\n%s", err, wantedYaml)
		t.FailNow()
	}
	// fmt.Printf("%v", actualSC)
	td.CmpDeeply(t, actualSC, wantedSC, "ETH and Bridge properties are not equal")
}
func Test__Bridge_with_vlans(t *testing.T) {
	wantedState := make(npstate.NPStates)
	brName := "br1"
	wantedState[brName] = &npstate.NPState{
		Name:   brName,
		Action: "bridge",
		Online: true,
		L3: npstate.L3State{
			IPv4: []string{"10.10.10.131/25"},
		},
	}
	for _, linkName := range []string{"eth1", "eth2", "eth3", "eth4", "eth5", "eth3.103"} {
		wantedState[linkName] = &npstate.NPState{
			Name:   linkName,
			Action: "port",
			Online: true,
		}
		if linkName == "eth3.103" {
			wantedState[linkName].L2 = npstate.L2State{
				Bridge:  brName,
				Parent:  "eth3",
				Vlan_id: 103,
			}
		} else if linkName != "eth2" && linkName != "eth4" {
			wantedState[linkName].L2 = npstate.L2State{
				Bridge: brName,
			}
		}

	}

	savedConfig := NewSavedConfig(nil)
	savedConfig.SetWantedState(&wantedState)
	savedConfig.Generate()
	actualYaml := savedConfig.String()
	actualSC := new(SavedConfig)
	if err := yaml.Unmarshal([]byte(actualYaml), actualSC); err != nil {
		t.Logf("Can't unmarshall the actual YAML: %s\n%s", err, actualYaml)
		t.FailNow()
	}
	wantedYaml := `
    version: 2
    renderer: networkd
    ethernets:
      eth1:
        dhcp4: false
        dhcp6: false
      eth2:
        dhcp4: false
        dhcp6: false
      eth3:
        dhcp4: false
        dhcp6: false
      eth4:
        dhcp4: false
        dhcp6: false
      eth5:
        dhcp4: false
        dhcp6: false
    vlans:
      eth3.103:
        id: 103
        link: eth3
        dhcp4: false
        dhcp6: false
    bridges:
      br1:
        interfaces: ["eth1","eth3","eth3.103","eth5"]
        addresses:
          - 10.10.10.131/25
        dhcp4: false
        dhcp6: false
`
	wantedSC := new(SavedConfig)
	if err := yaml.Unmarshal([]byte(wantedYaml), wantedSC); err != nil {
		t.Logf("Can't unmarshall the wanted YAML: %s\n%s", err, wantedYaml)
		t.FailNow()
	}
	// fmt.Printf("%v", actualSC)
	td.CmpDeeply(t, actualSC, wantedSC, "ETH and Bridge properties are not equal")
}

func Test__Just_Bond(t *testing.T) {
	wantedState := make(npstate.NPStates)
	bondName := "bond1"
	wantedState[bondName] = &npstate.NPState{
		Name:   bondName,
		Action: "bond",
		Online: true,
		L2: npstate.L2State{
			Slaves: []string{"eth2", "eth3"},
		},
		L3: npstate.L3State{
			IPv4: []string{"10.10.10.131/25"},
		},
	}
	for _, linkName := range []string{"eth2", "eth3"} {
		wantedState[linkName] = &npstate.NPState{
			Name:   linkName,
			Action: "port",
			Online: true,
		}
	}

	savedConfig := NewSavedConfig(nil)
	savedConfig.SetWantedState(&wantedState)
	savedConfig.Generate()
	actualYaml := savedConfig.String()
	actualSC := new(SavedConfig)
	if err := yaml.Unmarshal([]byte(actualYaml), actualSC); err != nil {
		t.Logf("Can't unmarshall the actual YAML: %s\n%s", err, actualYaml)
		t.FailNow()
	}
	wantedYaml := `
    version: 2
    renderer: networkd
    ethernets:
      eth2:
        dhcp4: false
        dhcp6: false
      eth3:
        dhcp4: false
        dhcp6: false
    bonds:
      bond1:
        interfaces: ["eth2","eth3"]
        addresses:
          - 10.10.10.131/25
        dhcp4: false
        dhcp6: false
`
	wantedSC := new(SavedConfig)
	if err := yaml.Unmarshal([]byte(wantedYaml), wantedSC); err != nil {
		t.Logf("Can't unmarshall the wanted YAML: %s\n%s", err, wantedYaml)
		t.FailNow()
	}
	//fmt.Printf("%v", actualSC)
	td.CmpDeeply(t, actualSC, wantedSC, "ETH and Bond properties are not equal")
}

func Test__Bond_into_bridge(t *testing.T) {
	wantedState := make(npstate.NPStates)
	brName := "br1"
	wantedState[brName] = &npstate.NPState{
		Name:   brName,
		Action: "bridge",
		Online: true,
		L3: npstate.L3State{
			IPv4: []string{"10.10.10.131/25"},
		},
	}
	bondName := "bond1"
	wantedState[bondName] = &npstate.NPState{
		Name:   bondName,
		Action: "bond",
		Online: true,
		L2: npstate.L2State{
			Bridge: "br1",
			Slaves: []string{"eth2", "eth3"},
		},
	}
	for _, linkName := range []string{"eth2", "eth3"} {
		wantedState[linkName] = &npstate.NPState{
			Name:   linkName,
			Action: "port",
			Online: true,
		}
	}

	savedConfig := NewSavedConfig(nil)
	savedConfig.SetWantedState(&wantedState)
	savedConfig.Generate()
	actualYaml := savedConfig.String()
	actualSC := new(SavedConfig)
	if err := yaml.Unmarshal([]byte(actualYaml), actualSC); err != nil {
		t.Logf("Can't unmarshall the actual YAML: %s\n%s", err, actualYaml)
		t.FailNow()
	}
	wantedYaml := `
    version: 2
    renderer: networkd
    ethernets:
      eth2:
        dhcp4: false
        dhcp6: false
      eth3:
        dhcp4: false
        dhcp6: false
    bridges:
      br1:
        interfaces: ["bond1"]
        addresses:
          - 10.10.10.131/25
        dhcp4: false
        dhcp6: false
    bonds:
      bond1:
        interfaces: ["eth2","eth3"]
        dhcp4: false
        dhcp6: false
`
	wantedSC := new(SavedConfig)
	if err := yaml.Unmarshal([]byte(wantedYaml), wantedSC); err != nil {
		t.Logf("Can't unmarshall the wanted YAML: %s\n%s", err, wantedYaml)
		t.FailNow()
	}
	fmt.Printf("%v", actualSC)
	td.CmpDeeply(t, actualSC, wantedSC, "ETH and Bond properties are not equal")
}

// -----------------------------------------------------------------------------