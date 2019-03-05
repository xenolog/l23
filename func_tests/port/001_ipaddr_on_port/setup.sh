#!/bin/bash
set -x

IFNAME=${IFNAME:-eth1}
IPADDR=${IPADDR:-10.1.251.1/24}

ip addr flush $IFNAME
ip l set down $IFNAME
