package main

import (
	"reflect"
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
			IPv4: []string{"127.0.0.1/8"},
		},
	}

	linkName = "eth1"
	rv.Link[linkName] = &NpLinkStatus{
		Name:   &linkName,
		Online: true,
		L3: L3Status{
			IPv4: []string{"10.20.30.40/24", "20.30.40.50/25"},
		},
	}
	return rv
}

func TestIfStatus__EqualNpStatuses(t *testing.T) {
	runtimeNps := RuntimeNpStatuses()
	wantedNps := RuntimeNpStatuses()
	diff := runtimeNps.Compare(wantedNps)

	// t.Logf("Runtime NPS: %s", runtimeNps)
	// t.Logf("Wanted NPS: %s", wantedNps)

	if !diff.IsEqual() {
		t.Fail()
	}
}

func TestIfStatus__ReducedIface(t *testing.T) {
	linkName := "eth1"
	runtimeNps := RuntimeNpStatuses()
	wantedNps := RuntimeNpStatuses()
	delete(wantedNps.Link, linkName)

	diff := runtimeNps.Compare(wantedNps)

	// t.Logf("Runtime NPS:\n%s", runtimeNps)
	// t.Logf("Wanted NPS:\n%s", wantedNps)
	// t.Logf("Diff:\n%s", diff)

	if len(diff.New) != 0 || len(diff.Different) != 0 || !reflect.DeepEqual(diff.Waste, []string{linkName}) {
		t.Fail()
	}

}

func TestIfStatus__AddedIface(t *testing.T) {
	runtimeNps := RuntimeNpStatuses()
	wantedNps := RuntimeNpStatuses()
	linkName := "eth2"
	wantedNps.Link[linkName] = &NpLinkStatus{
		Name:   &linkName,
		Online: true,
		L3: L3Status{
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
func TestIfStatus__DifferentIface(t *testing.T) {
	runtimeNps := RuntimeNpStatuses()
	wantedNps := RuntimeNpStatuses()
	linkName := "eth1"
	wantedNps.Link[linkName].L3.IPv4 = []string{"10.20.30.40/24", "20.30.40.55/25"}

	diff := runtimeNps.Compare(wantedNps)

	// t.Logf("Runtime NPS:\n%s", runtimeNps)
	// t.Logf("Wanted NPS:\n%s", wantedNps)
	// t.Logf("Diff:\n%s", diff)

	if len(diff.Waste) != 0 || len(diff.New) != 0 || !reflect.DeepEqual(diff.Different, []string{linkName}) {
		t.Fail()
	}

}
