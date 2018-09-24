package lnx

import (
	"github.com/vishvananda/netlink"
	logger "github.com/xenolog/go-tiny-logger"
	ifstatus "github.com/xenolog/l23/ifstatus"
)

const (
	MsgPrefix = "LNX plugin"
)

var LnxRtPluginEntryPoint *LnxRtPlugin

// -----------------------------------------------------------------------------

type NpOperator interface {
	Init(*ifstatus.NpLinkStatus) error
	Create(bool) error
	Remove(bool) error
	Modify(bool) error
	Name() string
	//todo(sv): Status() *NpLinkStatus // move status generation here
	//todo(sv): NS2Status() *NpLinkStatus // move wanted status generation from ns
}

type NpOperators map[string]interface{}

type RtPlugin interface {
	Init(*logger.Logger, *netlink.Handle) error
	Version() string
	Operators() NpOperators
	Observe() error // Observe runtime and build NPState
	NetworkState() *ifstatus.NpsStatus
	GetNp(string) *ifstatus.NpLinkStatus
	GetLogger() *logger.Logger
	GetHandle() *netlink.Handle
}

type LnxRtPlugin struct {
	log    *logger.Logger
	handle *netlink.Handle
	nps    *ifstatus.NpsStatus
}

// -----------------------------------------------------------------------------

type OpBase struct {
	plugin      *LnxRtPlugin
	log         *logger.Logger
	handle      *netlink.Handle
	wantedState *ifstatus.NpLinkStatus
	rtState     *ifstatus.NpLinkStatus
}

func (s *OpBase) Init(wantedState *ifstatus.NpLinkStatus) error {
	s.wantedState = wantedState
	s.rtState = nil
	return nil
}
func (s *OpBase) setupGlobals() {
	s.plugin = LnxRtPluginEntryPoint
	s.handle = s.plugin.handle
	s.log = s.plugin.log
}

func (s *OpBase) Name() string {
	return s.wantedState.Name
}

// -----------------------------------------------------------------------------

type L2Port struct {
	OpBase
}

func (s *L2Port) Create(dryrun bool) (err error) {
	if dryrun {
		s.log.Info("%s dryrun: Port '%s' created.", MsgPrefix, s.Name())
		return nil
	}

	return nil
}

func (s *L2Port) Remove(dryrun bool) error {
	if dryrun {
		s.log.Info("%s dryrun: Port '%s' removed.", MsgPrefix, s.Name())
		return nil
	}
	return nil
}

func (s *L2Port) Modify(dryrun bool) error {
	if dryrun {
		s.log.Info("%s dryrun: Port '%s' modifyed.", MsgPrefix, s.Name())
		return nil
	}
	return nil
}

func NewPort() NpOperator {
	rv := new(L2Port)
	rv.setupGlobals()
	return rv
}

// -----------------------------------------------------------------------------

type L2Bridge struct {
	OpBase
}

func (s *L2Bridge) Create(dryrun bool) (err error) {
	if dryrun {
		s.log.Info("%s dryrun: Bridge '%s' created.", MsgPrefix, s.Name())
		return nil
	}

	s.log.Info("%s Creating bridge '%s'", MsgPrefix, s.Name())
	br := netlink.Bridge{}
	br.Name = s.Name()
	if err = s.handle.LinkAdd(&br); err != nil {
		s.log.Error("%s: error while bridge creating: %v", MsgPrefix, err)
		return err
	} else {
		s.log.Info("%s: bridge created.", MsgPrefix)
	}

	err = s.Modify(dryrun)

	return err
}

func (s *L2Bridge) Remove(dryrun bool) (err error) {
	if dryrun {
		s.log.Info("%s: dryrun: Bridge '%s' removed.", MsgPrefix, s.Name())
		return nil
	}
	s.log.Info("%s: Removing bridge '%s'", MsgPrefix, s.Name())
	link, _ := netlink.LinkByName(s.Name())
	if err = s.handle.LinkSetDown(link); err != nil {
		s.log.Error("%s: error while bridge removing: %v", MsgPrefix, err)
		return err
	}
	if err = s.handle.LinkAdd(link); err != nil {
		s.log.Error("%s: error while bridge removing: %v", MsgPrefix, err)
	} else {
		s.log.Info("%s: bridge removed.", MsgPrefix)
	}
	return err
}

func (s *L2Bridge) Modify(dryrun bool) (err error) {
	if dryrun {
		s.log.Info("%s dryrun: Bridge '%s' modifyed.", MsgPrefix, s.Name())
		return nil
	}

	s.log.Info("%s: Modifying bridge '%s'", MsgPrefix, s.Name())
	link, _ := netlink.LinkByName(s.Name())
	attrs := link.Attrs()

	if err = s.handle.LinkSetDown(link); err != nil {
		s.log.Error("%s: error while bridge set to DOWN state: %v", MsgPrefix, err)
	}

	if s.wantedState.L2.Mtu > 0 && s.wantedState.L2.Mtu != attrs.MTU {
		s.log.Debug("%s: setting MTU to: %v", MsgPrefix, s.wantedState.L2.Mtu)
		if err = s.handle.LinkSetMTU(link, s.wantedState.L2.Mtu); err != nil {
			s.log.Error("%s: error while bridge set MTU: %v", MsgPrefix, err)
		}
	}

	if s.wantedState.Online {
		s.log.Debug("%s: setting to UP state", MsgPrefix)
		if err = s.handle.LinkSetUp(link); err != nil {
			s.log.Error("%s: error while bridge set to UP state: %v", MsgPrefix, err)
		}
	}

	return err
}

func NewBridge() NpOperator {
	rv := new(L2Bridge)
	rv.setupGlobals()
	return rv
}

// -----------------------------------------------------------------------------

// func NewIPv4() NpOperator {
// 	rv := new(NewIPv4)
// 	return rv
// }

// -----------------------------------------------------------------------------

func (s *LnxRtPlugin) Init(log *logger.Logger, hh *netlink.Handle) (err error) {
	if s.handle == nil && hh == nil {
		// generate new handle if need
		if s.handle, err = netlink.NewHandle(); err != nil {
			s.log.Error("%v", err)
			return err
		}
	} else if hh != nil {
		// setup handle
		s.handle = hh
	}
	s.log = log
	LnxRtPluginEntryPoint = s
	return nil
}

func (s *LnxRtPlugin) Operators() NpOperators {
	return NpOperators{
		"port":   NewPort,
		"bridge": NewBridge,
		// "endpoint":   NewIPv4,
	}
}

func (s *LnxRtPlugin) Version() string {
	return "LNX RUNTIME PLUGIN: v0.0.1"
}

func (s *LnxRtPlugin) Observe() error {
	s.nps = ifstatus.NewNpsStatus()
	return s.nps.ObserveRuntime()
}

func (s *LnxRtPlugin) NetworkState() *ifstatus.NpsStatus {
	return s.nps
}

func (s *LnxRtPlugin) GetNp(name string) *ifstatus.NpLinkStatus {
	rv, ok := s.nps.Link[name]
	if !ok {
		s.log.Error("Network primitive '%s' not found in the stored base", name)
		return nil
	}
	return rv
}

func (s *LnxRtPlugin) GetLogger() *logger.Logger {
	return s.log
}

func (s *LnxRtPlugin) GetHandle() *netlink.Handle {
	return s.handle
}

func NewLnxRtPlugin() *LnxRtPlugin {
	rv := new(LnxRtPlugin)
	return rv
}
