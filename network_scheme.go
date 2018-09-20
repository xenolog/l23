package main

import (
	"io"
	"io/ioutil"
	"log"

	yaml "gopkg.in/yaml.v2"
)

type NsIface struct {
	Mtu int `yaml:"mtu"`
}

type NsEp struct {
	Gateway       string   `yaml:"gateway"`
	GatewayMetric int      `yamk:"gateway_metric"`
	IP            []string `yaml:"IP"`
}

type NsIfaces map[string]NsIface
type NsEps map[string]NsEp

type NetworkScheme struct {
	Version         string   `yaml:"version"`
	Interfaces      NsIfaces `yaml:"interfaces"`
	Transformations string   `yaml:"transformations"`
	Endpoints       NsEps    `yaml:"endpoints"`
	Provider        string   `yaml:"provider"`
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
		Link: make(map[string]*NpLinkStatus),
	}

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

	for key, iface := range s.Interfaces {
		if _, ok := rv.Link[key]; !ok {
			rv.Link[key] = new(NpLinkStatus)
			rv.Link[key].Name = key
			rv.Link[key].Online = true
		}
		rv.Link[key].L2.MTU = iface.Mtu
	}

	return rv
}
