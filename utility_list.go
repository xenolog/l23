package main

import (
	"fmt"
	"net"

	cli "github.com/urfave/cli"
	"github.com/vishvananda/netlink"
	ifstatus "github.com/xenolog/l23/ifstatus"
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
		// err      error
		linkName string
		ipaddrs  string
		online   string
	)

	nps := ifstatus.NewNpsStatus()
	nps.ObserveRuntime()

	for _, link := range nps.Link {
		linkAttrs := link.Attrs()
		if linkAttrs.Alias == "" {
			linkName = link.Name
		} else {
			linkName = fmt.Sprintf("%s(%s)", link.Name, linkAttrs.Alias)
		}

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

		Log.Debug("%v", linkAttrs)
		Log.Info("%02d: %2s%15s (%s) %s", linkAttrs.Index, online, linkName, link.Type(), ipaddrs)
	}

	return nil
}
