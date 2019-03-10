#!/bin/bash
set -x

IFNAME=${IFNAME:-eth1.101}
BRNAME=${BRNAME:-br1}
# IPADDR=${IPADDR:-192.168.0.10/24}

master_ifname=$(echo ${IFNAME} | awk -F'.' '{print $1}')

ip link set up $master_ifname
ip addr flush dev $master_ifname

ip link show type vlan | grep "@${master_ifname}" | awk '{print $2}' | awk -F'@' '{print $1}' | xargs -n1 ip link del

ip link del $BRNAME