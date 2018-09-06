package main

import (
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"log"
)

type NsIface struct {
	Mtu uint `yaml:"mtu"`
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
