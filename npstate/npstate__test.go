package npstate

import (
	"reflect"
	"testing"
	// logger "github.com/xenolog/go-tiny-logger"
)

func RuntimeNpStatuses() *TopologyState {
	var linkName string
	rv := &TopologyState{
		NP: make(map[string]*NPState),
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

	linkName = "eth1"
	rv.NP[linkName] = &NPState{
		Name:   linkName,
		Action: "port",
		Online: true,
		L3: L3State{
			IPv4: []string{"10.20.30.40/24", "20.30.40.50/25"},
		},
	}
	return rv
}

func RuntimeNpStatusesForRemove() *TopologyState {
	var linkName string
	rv := &TopologyState{
		NP: make(map[string]*NPState),
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
		Action: "remove",
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
	return rv
}

func Test_Npstate__EqualNpStatuses(t *testing.T) {
	runtimeNps := RuntimeNpStatuses()
	wantedNps := RuntimeNpStatuses()
	diff := runtimeNps.Compare(wantedNps)
	if !diff.IsEqual() {
		t.Fail()
	}
	if t.Failed() {
		t.Logf("Runtime NPS: %s", runtimeNps)
		t.Logf("Wanted NPS: %s", wantedNps)
	}
}

func Test_Npstate__ReducedIface(t *testing.T) {
	linkName := "eth1"
	runtimeNps := RuntimeNpStatuses()
	wantedNps := RuntimeNpStatuses()
	delete(wantedNps.NP, linkName)

	diff := runtimeNps.Compare(wantedNps)

	// t.Logf("Runtime NPS:\n%s", runtimeNps)
	// t.Logf("Wanted NPS:\n%s", wantedNps)
	// t.Logf("Diff:\n%s", diff)

	if len(diff.New) != 0 || len(diff.Different) != 0 || !reflect.DeepEqual(diff.Waste, []string{linkName}) {
		t.Fail()
	}

}

func Test_Npstate__ReducedIface__by_remove_action(t *testing.T) {
	linkName := "eth0"
	runtimeNps := RuntimeNpStatusesForRemove()
	wantedNps := RuntimeNpStatusesForRemove()

	diff := runtimeNps.Compare(wantedNps)

	// t.Logf("Runtime NPS:\n%s", runtimeNps)
	// t.Logf("Wanted NPS:\n%s", wantedNps)
	// t.Logf("Diff:\n%s", diff)

	if len(diff.New) != 0 || len(diff.Different) != 0 || !reflect.DeepEqual(diff.Waste, []string{linkName}) {
		t.Fail()
	}

	if t.Failed() {
		t.Logf("DIFF: \n%v", diff)
		t.Logf("Runtime NPS: %s", runtimeNps)
		t.Logf("Wanted NPS: %s", wantedNps)
	}

}

func Test_Npstate__AddedIface(t *testing.T) {
	runtimeNps := RuntimeNpStatuses()
	wantedNps := RuntimeNpStatuses()
	linkName := "eth2"
	wantedNps.NP[linkName] = &NPState{
		Name:   linkName,
		Action: "port",
		Online: true,
		L3: L3State{
			IPv4: []string{"192.168.0.1/24"},
		},
	}

	diff := runtimeNps.Compare(wantedNps)

	// t.Logf("Runtime NPS:\n%s", runtimeNps)
	// t.Logf("Wanted NPS:\n%s", wantedNps)
	// t.Logf("Diff:\n%s", diff)

	if len(diff.Waste) != 0 || len(diff.Different) != 0 || !reflect.DeepEqual(diff.New, []string{linkName}) {
		t.Fail()
	}

}
func Test_Npstate__DifferentIface(t *testing.T) {
	runtimeNps := RuntimeNpStatuses()
	wantedNps := RuntimeNpStatuses()
	linkName := "eth1"
	wantedNps.NP[linkName].L3.IPv4 = []string{"10.20.30.40/24", "20.30.40.55/25"}

	diff := runtimeNps.Compare(wantedNps)

	// t.Logf("Runtime NPS:\n%s", runtimeNps)
	// t.Logf("Wanted NPS:\n%s", wantedNps)
	// t.Logf("Diff:\n%s", diff)

	if len(diff.Waste) != 0 || len(diff.New) != 0 || !reflect.DeepEqual(diff.Different, []string{linkName}) {
		t.Fail()
	}

}
