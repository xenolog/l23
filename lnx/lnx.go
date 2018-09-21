package lnx

import (
	"github.com/vishvananda/netlink"
	logger "github.com/xenolog/go-tiny-logger"
	ifstatus "github.com/xenolog/l23/ifstatus"
)

const (
	MsgPrefix = "LNX plugin"
)

// -----------------------------------------------------------------------------

type L2Operator interface {
	Init(*logger.Logger, *netlink.Handle, *ifstatus.NpLinkStatus) error
	Create(bool) error
	Remove(bool) error
	Modify(bool) error
	Name() string
	//todo(sv): Status() *NpLinkStatus // move status generation here
	//todo(sv): NS2Status() *NpLinkStatus // move wanted status generation from ns
}

//type L2Operators map[string]interface{}
type L2Operators map[string]interface{}

type RtPlugin interface {
	Init(*logger.Logger, *netlink.Handle) error
	Version() string
	Operators() L2Operators
	Observe() error // Observe runtime and build NPState
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

type L2Base struct {
	log    *logger.Logger
	handle *netlink.Handle
	state  *ifstatus.NpLinkStatus
}

func (s *L2Base) Init(log *logger.Logger, handle *netlink.Handle, st *ifstatus.NpLinkStatus) error {
	s.handle = handle
	s.log = log
	s.state = st
	return nil
}

func (s *L2Base) Name() string {
	return s.state.Name
}

// -----------------------------------------------------------------------------

type L2Port struct {
	L2Base
}

func (s *L2Port) Create(dryrun bool) error {
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

func NewPort() L2Operator {
	rv := new(L2Port)
	return rv
}

// -----------------------------------------------------------------------------

type L2Bridge struct {
	L2Base
}

func (s *L2Bridge) Create(dryrun bool) error {
	if dryrun {
		s.log.Info("%s dryrun: Bridge '%s' created.", MsgPrefix, s.Name())
		return nil
	}
	return nil
}

func (s *L2Bridge) Remove(dryrun bool) error {
	if dryrun {
		s.log.Info("%s dryrun: Bridge '%s' removed.", MsgPrefix, s.Name())
		return nil
	}
	return nil
}

func (s *L2Bridge) Modify(dryrun bool) error {
	if dryrun {
		s.log.Info("%s dryrun: Bridge '%s' modifyed.", MsgPrefix, s.Name())
		return nil
	}
	return nil
}

// func NewBridge() *L2Bridge {
func NewBridge() L2Operator {
	rv := new(L2Bridge)
	return rv
}

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
	return nil
}

func (s *LnxRtPlugin) Operators() L2Operators {
	return L2Operators{
		"port":   NewPort,
		"bridge": NewBridge,
	}
}

func (s *LnxRtPlugin) Version() string {
	return "LNX RUNTIME PLUGIN: v0.0.1"
}

func (s *LnxRtPlugin) Observe() error {
	s.nps = ifstatus.NewNpsStatus()
	return s.nps.ObserveRuntime()
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
