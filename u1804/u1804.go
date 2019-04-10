package u1804

import (
	"errors"

	"gopkg.in/yaml.v2"

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
	Ethernets   SCEthernets `yaml:",omitempty"` //, inline, flow
	Bonds       SCBonds     `yaml:",omitempty"`
	Vlans       SCVlans     `yaml:",omitempty"`
	Bridges     SCBridges   `yaml:",omitempty"`
}

// -----------------------------------------------------------------------------

func (s *SavedConfig) Generate() error {
	if s.wantedState == nil {
		errMsg := "Wanted states of Network primitives are not set."
		s.log.Error("%s: %s", MsgPrefix, errMsg)
		return errors.New(errMsg)
	}
	for _, np := range *s.wantedState {
		switch np.Action {
		case "port":
			if np.L2.Vlan_id != 0 {
				// vlan
				if _, ok := s.Ethernets[np.L2.Parent]; !ok {
					s.Ethernets[np.L2.Parent] = &SCEthernet{}
				}
				s.Vlans[np.Name] = &SCVlan{
					Id:   np.L2.Vlan_id,
					Link: np.L2.Parent,
				}
				s.Vlans[np.Name].AddAddresses(np.L3.IPv4)
			} else {
				// just ethernet
				s.Ethernets[np.Name] = &SCEthernet{}
				s.Ethernets[np.Name].AddAddresses(np.L3.IPv4)
			}
			// default:
			// 	errMsg := fmt.Sprintf("Unsupported 'action' for '%s'.", np.Name)
			// 	s.log.Error("%s: %s", MsgPrefix, errMsg)
			// 	return errors.New(errMsg)
		}

	}
	return nil
}

func (s *SavedConfig) String() string {
	rv, _ := yaml.Marshal(s)
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
		// Bonds:     make(SCBonds),
		Vlans:   make(SCVlans),
		Bridges: make(SCBridges),
	}
	rv.SetLogger(log)

	return rv
}
