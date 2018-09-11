package main

import (
	// "github.com/vishvananda/netlink"
	// "github.com/xenolog/go-tiny-logger"
	"fmt"
	cli "github.com/urfave/cli"
)

func UtilityListNetworkPrimitives(c *cli.Context) error {
	fmt.Println("new task template: ", c.Args().First())
	return nil
}
