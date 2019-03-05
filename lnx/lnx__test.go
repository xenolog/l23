package lnx

import (
	"reflect"
	"sort"
	"testing"

	logger "github.com/xenolog/go-tiny-logger"
	. "github.com/xenolog/l23/npstate"
	. "github.com/xenolog/l23/plugin"
	. "github.com/xenolog/l23/utils"
)

func TestLNX__OperatorList(t *testing.T) {
	lnxRtPlugin := NewLnxRtPlugin()
	keys := []string{}
	operators := lnxRtPlugin.Operators()
	for _, key := range reflect.ValueOf(operators).MapKeys() {
		keys = append(keys, key.String())
	}
	sort.Strings(keys)
	wantedKeys := []string{"bridge", "port"}

	if !reflect.DeepEqual(keys, wantedKeys) {
		t.Logf("Operator list from LnxRtPlugin broken, given %v, instead %v", keys, wantedKeys)
		t.Fail()
	}
}

// -----------------------------------------------------------------------------

func RuntimeNpStatuses__1__exists() *TopologyState {
	var linkName string
	rv := &TopologyState{
		NP:    make(map[string]*NPState),
		Order: []string{},
	}

	linkName = "lo"
	rv.NP[linkName] = &NPState{
		Name:   linkName,
		Action: "port",
		Online: true,
		L3: L3State{
			IPv4: []string{"127.0.0.1/8"},
		},
	}

	linkName = "eth0"
	rv.NP[linkName] = &NPState{
		Name:   linkName,
		Action: "port",
		Online: true,
		L3: L3State{
			IPv4: []string{"10.10.10.222/24"},
		},
	}

	linkName = "eth1"
	rv.NP[linkName] = &NPState{
		Name:   linkName,
		Action: "port",
		Online: true,
		L3: L3State{
			IPv4: []string{"10.20.30.40/24"},
		},
	}

	linkName = "eth1.222"
	rv.NP[linkName] = &NPState{
		Name:   linkName,
		Action: "port",
		Online: true,
	}

	for _, key := range reflect.ValueOf(rv.NP).MapKeys() {
		rv.Order = append(rv.Order, key.String())
	}
	sort.Strings(rv.Order)
	return rv
}

func RuntimeNpStatuses__1__wanted() *TopologyState {
	var linkName string
	rv := &TopologyState{
		NP:    make(map[string]*NPState),
		Order: []string{},
	}

	linkName = "lo"
	rv.NP[linkName] = &NPState{
		Name:   linkName,
		Action: "port",
		Online: true,
		L3: L3State{
			IPv4: []string{"127.0.0.1/8"},
		},
	}

	linkName = "eth0"
	rv.NP[linkName] = &NPState{
		Name:   linkName,
		Action: "port",
		Online: true,
		L3: L3State{
			IPv4: []string{"10.10.10.1/24"},
		},
	}

	linkName = "eth1"
	rv.NP[linkName] = &NPState{
		Name:   linkName,
		Action: "port",
		Online: true,
		L3: L3State{
			IPv4: []string{"10.20.30.40/24", "20.30.40.50/25"},
		},
	}

	linkName = "eth1.101"
	rv.NP[linkName] = &NPState{
		Name:   linkName,
		Action: "port",
		Online: true,
		L2: L2State{
			Bridge:  "br4",
			Parent:  "eth1",
			Vlan_id: 101,
		},
	}

	linkName = "br4"
	rv.NP[linkName] = &NPState{
		Name:   linkName,
		Action: "bridge",
		Online: true,
		L3: L3State{
			IPv4: []string{"10.40.40.1/24"},
		},
	}
	for _, key := range reflect.ValueOf(rv.NP).MapKeys() {
		rv.Order = append(rv.Order, key.String())
	}
	sort.Strings(rv.Order)
	return rv
}

func TestLNX__1__MainRun(t *testing.T) {
	log := logger.New()
	runtimeNps := RuntimeNpStatuses__1__exists()
	wantedNps := RuntimeNpStatuses__1__wanted()
	diff := runtimeNps.Compare(wantedNps)

	lnxRtPlugin := NewLnxRtPlugin()
	lnxRtPlugin.Init(log, nil)
	operators := lnxRtPlugin.Operators()
	t.Logf("Diff: %s", diff)

	// report
	npCreated := []string{}
	npRemoved := []string{}
	npModifyed := []string{}
	// walk ordr and implement diffs
	for _, npName := range wantedNps.Order {
		action, ok := operators[wantedNps.NP[npName].Action]
		if !ok {
			t.Logf("Unsupported actiom '%s' for '%s', skipped", action, npName)
			t.Fail()
			continue
		}
		t.Logf("action: %s", action)
		oper := action.(func() NpOperator)()
		oper.Init(wantedNps.NP[npName])

		t.Logf(npName)
		if IndexString(diff.Waste, npName) >= 0 {
			// this NP should be removed
			oper.Remove(true)
			npRemoved = append(npRemoved, npName)
		} else if IndexString(diff.New, npName) >= 0 {
			// this NP shoujld be created
			oper.Create(true)
			npCreated = append(npCreated, npName)
		} else if IndexString(diff.Different, npName) >= 0 {
			oper.Modify(true)
			npModifyed = append(npModifyed, npName)
		}
	}

	// evaluate report
	if !reflect.DeepEqual(npCreated, []string{"br4", "eth1.101"}) {
		t.Logf("Problen while creating resources: %v", npCreated)
		t.Fail()
	}
	if !reflect.DeepEqual(npModifyed, []string{"eth0", "eth1"}) {
		t.Logf("Problen while modifying resources: %v", npModifyed)
		t.Fail()
	}
	//todo(sv): there are no Removing, because "wanted" network_scheme has no
	// description of absent resources. TBD !!!
	//
	// if !reflect.DeepEqual(npRemoved, []string{"eth1.222"}) {
	// 	t.Logf("Problen while removing resources: %v", npRemoved)
	// 	t.Fail()
	// }
}
