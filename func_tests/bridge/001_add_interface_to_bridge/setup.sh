#!/bin/bash
set -x

IFNAME=${IFNAME:-eth1}
BRNAME=${BRNAME:-br1}

ip addr flush $IFNAME
ip link set down $IFNAME

ip link show type vlan | grep "@${IFNAME}" | awk '{print $2}' | awk -F'@' '{print $1}' | xargs -n1 ip link del

ip link del $BRNAME