package u1804

import (
	"errors"
	"fmt"
	"sort"

	"gopkg.in/yaml.v3"

	logger "github.com/xenolog/go-tiny-logger"
	npstate "github.com/xenolog/l23/npstate"
	// . "github.com/xenolog/l23/plugin"
	// . "github.com/xenolog/l23/utils"
	// "golang.org/x/sys/unix"
)

const (
	MsgPrefix = "Netplan plugin"
)

// -----------------------------------------------------------------------------

type SCBase struct {
	Addresses []string `yaml:",omitempty"`
	Dhcp4     bool
	Dhcp6     bool
}

func (s *SCBase) AddAddresses(aa []string) {
	if len(aa) > 0 {
		s.Addresses = append(s.Addresses, aa...)
	}
}

// -----------------------------------------------------------------------------

type SCVlan struct {
	SCBase `yaml:",inline"`
	Id     int
	Link   string
}
type SCVlans map[string]*SCVlan

type SCBridge struct {
	SCBase     `yaml:",inline"`
	Interfaces []string `yaml:",omitempty"`
}
type SCBridges map[string]*SCBridge

type SCBond struct {
	SCBase     `yaml:",inline"`
	Interfaces []string `yaml:",omitempty"`
}
type SCBonds map[string]*SCBond

type SCEthernet struct {
	SCBase `yaml:",inline"`
	Match  map[string]string `yaml:",omitempty"`
}
type SCEthernets map[string]*SCEthernet

type SavedConfig struct {
	log         *logger.Logger
	wantedState *npstate.NPStates
	Version     string
	Renderer    string
	Ethernets   SCEthernets `yaml:",omitempty"`
	Bonds       SCBonds     `yaml:",omitempty"`
	Vlans       SCVlans     `yaml:",omitempty"`
	Bridges     SCBridges   `yaml:",omitempty"`
}

// -----------------------------------------------------------------------------

func (s *SavedConfig) addEthIfRequired(name string) {
	if _, ok := s.Ethernets[name]; !ok {
		s.Ethernets[name] = &SCEthernet{}
	}
}

func (s *SavedConfig) addBrIfRequired(name string) {
	if _, ok := s.Bridges[name]; !ok {
		s.Bridges[name] = &SCBridge{}
	}
}

func (s *SavedConfig) CheckWS() (rv error) {
	if s.wantedState == nil {
		errMsg := "Wanted states of Network primitives are not set."
		s.log.Error("%s: %s", MsgPrefix, errMsg)
		rv = errors.New(errMsg)
	}
	return rv
}

func (s *SavedConfig) Generate() error {
	if err := s.CheckWS(); err != nil {
		return err
	}
	for _, np := range *s.wantedState {
		switch np.Action {
		case "port":
			if np.L2.VlanID != 0 {
				// vlan
				//s.addEthIfRequired(np.L2.Parent)  // there are no such action here!!! Parent Interface always into NetworkState will be.
				s.Vlans[np.Name] = &SCVlan{
					Id:   np.L2.VlanID,
					Link: np.L2.Parent,
				}
				s.Vlans[np.Name].AddAddresses(np.L3.IPv4)
			} else {
				// just ethernet
				s.addEthIfRequired(np.Name)
				s.Ethernets[np.Name].AddAddresses(np.L3.IPv4)
			}
		case "bridge":
			var ports []string
			s.addBrIfRequired(np.Name)
			for _, member := range *s.wantedState {
				if member.L2.Bridge == np.Name {
					ports = append(ports, member.Name)
				}
			}
			if len(ports) > 0 {
				sort.Strings(ports)
				s.Bridges[np.Name].Interfaces = append(s.Bridges[np.Name].Interfaces, ports...)
			}
			s.Bridges[np.Name].AddAddresses(np.L3.IPv4)
		case "bond":
			if _, ok := s.Bonds[np.Name]; !ok {
				s.Bonds[np.Name] = &SCBond{}
			}
			s.Bonds[np.Name].Interfaces = np.L2.Slaves
			s.Bonds[np.Name].AddAddresses(np.L3.IPv4)
		default:
			errMsg := fmt.Sprintf("Unsupported 'action' for '%s'.", np.Name)
			s.log.Error("%s: %s", MsgPrefix, errMsg)
			return errors.New(errMsg)
		}

	}
	return nil
}

func (s *SavedConfig) String() string {
	type xxx struct {
		Network *SavedConfig
	}
	rv, _ := yaml.Marshal(&xxx{Network: s})
	return string(rv[:])
}

func (s *SavedConfig) SetWantedState(wantedState *npstate.NPStates) {
	s.wantedState = wantedState
}

func (s *SavedConfig) SetLogger(log *logger.Logger) {

	if log != nil {
		s.log = log
	} else {
		s.log = new(logger.Logger)
	}
}

// -----------------------------------------------------------------------------

func NewSavedConfig(log *logger.Logger) *SavedConfig {
	rv := &SavedConfig{
		Version:   "2",
		Renderer:  "networkd",
		Ethernets: make(SCEthernets),
		Bonds:     make(SCBonds),
		Vlans:     make(SCVlans),
		Bridges:   make(SCBridges),
	}
	rv.SetLogger(log)

	return rv
}
