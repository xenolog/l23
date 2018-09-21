package lnx

import (
	"testing"

	logger "github.com/xenolog/go-tiny-logger"
	. "github.com/xenolog/l23/ifstatus"
	. "github.com/xenolog/l23/utils"
)

func RuntimeNpStatuses() *NpsStatus {
	var linkName string
	rv := &NpsStatus{
		Link:  make(map[string]*NpLinkStatus),
		Order: []string{"lo", "eth1", "br4"},
	}

	linkName = "lo"
	rv.Link[linkName] = &NpLinkStatus{
		Name:   linkName,
		Action: "port",
		Online: true,
		L3: L3Status{
			IPv4: []string{"127.0.0.1/8"},
		},
	}

	linkName = "eth1"
	rv.Link[linkName] = &NpLinkStatus{
		Name:   linkName,
		Action: "port",
		Online: true,
		L3: L3Status{
			IPv4: []string{"10.20.30.40/24", "20.30.40.50/25"},
		},
	}

	linkName = "br4"
	rv.Link[linkName] = &NpLinkStatus{
		Name:   linkName,
		Action: "bridge",
		Online: true,
	}
	return rv
}

// func TestLNX__Callable_hash(t *testing.T) {
// 	runtimeNps := &NpsStatus{
// 		Link: make(map[string]*NpLinkStatus),
// 		Log:  logger.New(),
// 	}
// 	// we need create ALL
// 	wantedNps := RuntimeNpStatuses()
// 	diff := runtimeNps.Compare(wantedNps)

// 	// t.Logf("Runtime NPS: %s", runtimeNps)
// 	// t.Logf("Wanted NPS: %s", wantedNps)
// 	t.Logf("Diff: %s", diff)

// 	if !diff.IsEqual() {
// 		t.Fail()
// 	}
// }

func TestLNX__CallableHash(t *testing.T) {
	log := logger.New()
	runtimeNps := &NpsStatus{
		Link: make(map[string]*NpLinkStatus),
	}
	// we need create ALL
	wantedNps := RuntimeNpStatuses()
	diff := runtimeNps.Compare(wantedNps)

	lnxRtPlugin := NewLnxRtPlugin()
	lnxRtPlugin.Init(log, nil)
	operators := lnxRtPlugin.Operators()

	// walk ordr and implement diffs
	for _, npName := range wantedNps.Order {
		action, ok := operators[wantedNps.Link[npName].Action]
		if !ok {
			log.Error("Unsupported actiom '%s' for '%s', skipped", action, npName)
			continue
		}
		oper := action.(func() L2Operator)()
		oper.Init(log, lnxRtPlugin.GetHandle(), lnxRtPlugin.GetNp(npName))

		if IndexString(diff.Waste, npName) >= 0 {
			// this NP should be removed
			oper.Remove(true)
		} else if IndexString(diff.New, npName) >= 0 {
			// this NP shoujld be created
			oper.Create(true)
		} else if IndexString(diff.New, npName) >= 0 {
			oper.Modify(true)
		}
	}

	// t.Logf("Runtime NPS: %s", runtimeNps)
	// t.Logf("Wanted NPS: %s", wantedNps)
	t.Logf("Diff: %s", diff)
	t.Logf("Diff: %v", operators)

	if !diff.IsEqual() {
		t.Fail()
	}
}
