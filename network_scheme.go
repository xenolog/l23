package main

import (
	"io"
	"io/ioutil"
	"log"
	"reflect"
	"sort"

	npstate "github.com/xenolog/l23/npstate"

	. "github.com/xenolog/l23/utils"
	yaml "gopkg.in/yaml.v3"
)

type NsPrimitive struct {
	Action       string   `yaml:"action"`
	Name         string   `yaml:"name"`
	Mtu          int      `yaml:"mtu,omitempty"`
	Bridge       string   `yaml:"bridge,omitempty"`
	Parent       string   `yaml:"parent,omitempty"`
	Slaves       []string `yaml:"slaves,omitempty"`
	Vlan_id      int      `yaml:"vlan_id,omitempty"`
	Stp          bool     `yaml:"stp,omitempty"`
	Bpdu_forward bool     `yaml:"bpdu_forward,omitempty"`
	Type         string   `yaml:"Type,omitempty"`
	Provider     string   `yaml:"provider"`
	// Vendor_specific string   `yaml:"vendor_specific,omitempty"`
	// Ethtool
	// External_ids
	// Bond_properties
	// Interface_properties
}

type NsEp struct {
	Gateway       string   `yaml:"gateway,omitempty"`
	GatewayMetric int      `yamk:"gateway_metric,omitempty"`
	IP            []string `yaml:"IP"`
}

type NsTransformations []NsPrimitive
type NsIfaces map[string]NsPrimitive
type NsEps map[string]NsEp

type NetworkScheme struct {
	Version         string            `yaml:"version"`
	Interfaces      NsIfaces          `yaml:"interfaces"`
	Transformations NsTransformations `yaml:"transformations"`
	Endpoints       NsEps             `yaml:"endpoints"`
	Provider        string            `yaml:"provider"`
}

// func (s *NetworkScheme) setLogger(log *logger.Logger) {
// 	if log != nil {
// 		s.log = log
// 	} else if s.log == nil && log == nil {
// 		s.log = new(logger.Logger)
// 	}
// }

func (s *NetworkScheme) Load(r io.Reader) (err error) {
	var (
		data []byte
	)
	// todo(sv): forget about ReadAll
	if data, err = ioutil.ReadAll(r); err != nil {
		log.Printf("NetworkScheme YAML reading error: %v", err)
		return
	}

	if err = yaml.Unmarshal(data, s); err != nil {
		log.Printf("NetworkScheme YAML parsing error: %v", err)
		return
	}
	return
}

func (s *NetworkScheme) TopologyState() *npstate.TopologyState {

	rv := &npstate.TopologyState{
		NP:              make(map[string]*npstate.NPState),
		Order:           []string{},
		DefaultProvider: "lnx",
	}

	if s.Provider != "" {
		rv.DefaultProvider = s.Provider
	}

	// transformations should be processed first. Unlisted interfaces will be
	// added later
	for _, tr := range s.Transformations {
		if IndexString(rv.Order, tr.Name) < 0 {
			rv.Order = append(rv.Order, tr.Name)
		}
		if _, ok := rv.NP[tr.Name]; !ok {
			rv.NP[tr.Name] = new(npstate.NPState)
			rv.NP[tr.Name].Name = tr.Name
			rv.NP[tr.Name].Online = true
		}
		if tr.Action != "" {
			rv.NP[tr.Name].Action = tr.Action
		}
		//todo(sv): call corresponded interface for resource
		rv.NP[tr.Name].L2.Mtu = tr.Mtu
		rv.NP[tr.Name].L2.Bridge = tr.Bridge
		rv.NP[tr.Name].L2.Parent = tr.Parent
		rv.NP[tr.Name].L2.Slaves = tr.Slaves
		rv.NP[tr.Name].L2.Vlan_id = tr.Vlan_id
		rv.NP[tr.Name].L2.Stp = tr.Stp                   // todo(sv): move to vendor_specific
		rv.NP[tr.Name].L2.Bpdu_forward = tr.Bpdu_forward // todo(sv): move to vendor_specific
		if tr.Provider != "" {
			rv.NP[tr.Name].Provider = tr.Provider
		} else if rv.NP[tr.Name].Provider == "" {
			rv.NP[tr.Name].Provider = rv.DefaultProvider
		}
	}

	// Add unlisted interface names into head of ordering
	// (required by design !!!)
	iflist := []string{}
	for _, key := range reflect.ValueOf(s.Interfaces).MapKeys() {
		keyName := key.String()
		if IndexString(rv.Order, keyName) < 0 {
			iflist = append(iflist, keyName)
		}
	}
	sort.Strings(iflist)
	iflist = ReverseString(iflist)
	for _, key := range iflist {
		if _, ok := rv.NP[key]; !ok {
			rv.NP[key] = new(npstate.NPState)
			rv.NP[key].Name = key
			rv.NP[key].Online = true
			rv.Order = PrependString(rv.Order, key)
		}
		rv.NP[key].Action = "port"
		//todo(sv): call corresponded interface for resource
		rv.NP[key].L2.Mtu = s.Interfaces[key].Mtu
		if s.Interfaces[key].Provider != "" {
			rv.NP[key].Provider = s.Interfaces[key].Provider
		} else {
			rv.NP[key].Provider = rv.DefaultProvider
		}
	}

	// endpoints should be processed last
	for key, endpoint := range s.Endpoints {
		if IndexString(rv.Order, key) < 0 {
			log.Printf("Endpoint '%s' is not an interface or network primitive created by transformation", key)
			continue
		}
		if _, ok := rv.NP[key]; !ok {
			rv.NP[key] = new(npstate.NPState)
			rv.NP[key].Name = key
			rv.NP[key].Online = true
		}
		if len(endpoint.IP) > 0 && endpoint.IP[0] != "none" {
			rv.NP[key].L3.IPv4 = make([]string, len(endpoint.IP))
			copy(rv.NP[key].L3.IPv4, endpoint.IP)
		}
	}

	return rv
}
