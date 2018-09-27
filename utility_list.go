package main

import (
	"fmt"
	"net"

	cli "github.com/urfave/cli"
	"github.com/vishvananda/netlink"
	"github.com/xenolog/l23/lnx"
)

func UtilityListNetworkPrimitivesOld(c *cli.Context) error {
	var (
		h        *netlink.Handle
		ll       []netlink.Link
		err      error
		linkName string
		online   string
	)

	if h, err = netlink.NewHandle(); err != nil {
		Log.Fail("%v", err)
	}

	if ll, err = h.LinkList(); err != nil {
		Log.Fail("%v", err)
	}

	for _, link := range ll {
		linkAttrs := link.Attrs()
		if linkAttrs.Alias == "" {
			linkName = linkAttrs.Name
		} else {
			linkName = fmt.Sprintf("%s(%s)", linkAttrs.Name, linkAttrs.Alias)
		}

		online = ""
		if linkAttrs.Flags&net.FlagUp != 0 {
			online = "UP"
		}

		Log.Debug("%v", linkAttrs)
		Log.Info("%02d: %2s%15s (%s) ", linkAttrs.Index, online, linkName, link.Type())
	}

	return nil
}

func UtilityListNetworkPrimitives(c *cli.Context) error {
	var (
		ipaddrs string
		online  string
	)

	// initialize and configure LnxRtPlugin
	lnxRtPlugin := lnx.NewLnxRtPlugin()
	lnxRtPlugin.Init(Log, nil)
	lnxRtPlugin.Observe()
	Log.Debug("LnxRtPlugin initialized")

	nps := lnxRtPlugin.Topology()

	for _, link := range nps.NP {

		ipaddrs = ""
		sep := ""
		for _, ip := range link.L3.IPv4 {
			ipaddrs = fmt.Sprintf("%s%s%s", ipaddrs, sep, ip)
			sep = ","
		}

		online = ""
		if link.Online {
			online = "UP"
		}

		Log.Info("%02d: %2s%15s (%s) %s", link.IfIndex, online, link.Name, link.Type(), ipaddrs)
	}

	return nil
}
