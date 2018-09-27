package plugin

import (
	"github.com/vishvananda/netlink"
	logger "github.com/xenolog/go-tiny-logger"
	npstate "github.com/xenolog/l23/npstate"
)

// -----------------------------------------------------------------------------

type NpOperator interface {
	Init(*npstate.NPState) error
	Create(bool) error
	Remove(bool) error
	Modify(bool) error
	Name() string
	IPv4addrList() []string
	//todo(sv): State() *NPState // move status generation here
	// Link() netlink.NP	// This two methods are Provider-specific
	// IfIndex() int		// IMHO it is a big cons to intlude to interface
}

type NpOperators map[string]interface{}

type RtPlugin interface {
	Init(*logger.Logger, *netlink.Handle) error
	Version() string
	Operators() NpOperators
	Observe() error // Observe runtime and build NPState
	NetworkState() *npstate.TopologyState
	GetLogger() *logger.Logger
	// GetNp(string) *npstate.NPState
	// GetHandle() *netlink.Handle
}

// -----------------------------------------------------------------------------
