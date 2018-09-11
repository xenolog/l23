package main

import (
	"github.com/vishvananda/netlink"
	"gopkg.in/xenolog/go-tiny-logger.v1"
)

type IpAddr4 struct {
	IpAddr string
	Mask   string
}

// todo(sv): implement IPv6 support
// type IpAddr6 struct {
// 	IpAddr string
// 	Mask   string
// }

type L2Status struct {
}

type L3Status struct {
	IPv4 []netlink.Addr
	// IPv6 []IpAddr6
}

// Np -- is a acronym for Network Primitive
type NpStatus struct {
	Name   sting
	IfNum  int
	Online bool
	L2     L2Status
	L3     L3Status
}

func (s *NpStatus) Load(r io.Reader) (err error) {

	return
}
