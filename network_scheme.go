package main

import (
	"io"
	"io/ioutil"
	"log"
	"reflect"
	"sort"

	yaml "gopkg.in/yaml.v2"
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

func (s *NetworkScheme) NpsStatus() *NpsStatus {

	rv := &NpsStatus{
		Link:            make(map[string]*NpLinkStatus),
		Order:           []string{},
		DefaultProvider: "lnx",
	}

	if s.Provider != "" {
		rv.DefaultProvider = s.Provider
	}

	// sort by interface name is required by design !!!
	iflist := []string{}
	for _, key := range reflect.ValueOf(s.Interfaces).MapKeys() {
		iflist = append(iflist, key.String())
	}
	sort.Strings(iflist)
	for _, key := range iflist {
		if _, ok := rv.Link[key]; !ok {
			rv.Link[key] = new(NpLinkStatus)
			rv.Link[key].Name = key
			rv.Link[key].Online = true
			rv.Order = append(rv.Order, key)
		}
		//todo(sv): call corresponded interface for resource
		rv.Link[key].L2.MTU = s.Interfaces[key].Mtu
		if s.Interfaces[key].Provider != "" {
			rv.Link[key].Provider = s.Interfaces[key].Provider
		} else {
			rv.Link[key].Provider = rv.DefaultProvider
		}
	}

	// transformations should be processed here
	for _, tr := range s.Transformations {
		if IndexString(rv.Order, tr.Name) < 0 {
			rv.Order = append(rv.Order, tr.Name)
		}
		if _, ok := rv.Link[tr.Name]; !ok {
			rv.Link[tr.Name] = new(NpLinkStatus)
			rv.Link[tr.Name].Name = tr.Name
			rv.Link[tr.Name].Online = true
		}
		//todo(sv): call corresponded interface for resource
		rv.Link[tr.Name].L2.MTU = tr.Mtu
		if tr.Provider != "" {
			rv.Link[tr.Name].Provider = tr.Provider
		} else if rv.Link[tr.Name].Provider == "" {
			rv.Link[tr.Name].Provider = rv.DefaultProvider
		}
	}

	// endpoints should be processed last
	for key, endpoint := range s.Endpoints {
		if _, ok := rv.Link[key]; !ok {
			rv.Link[key] = new(NpLinkStatus)
			rv.Link[key].Name = key
			rv.Link[key].Online = true
		}
		if len(endpoint.IP) > 0 && endpoint.IP[0] != "none" {
			rv.Link[key].L3.IPv4 = make([]string, len(endpoint.IP))
			copy(rv.Link[key].L3.IPv4, endpoint.IP)
		}
	}

	return rv
}
