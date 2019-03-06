#!/bin/bash
set -x

IFNAME=${IFNAME:-eth1.101}
BRNAME=${BRNAME:-br1}
# IPADDR=${IPADDR:-192.168.0.10/24}

master_ifname=$(echo ${IFNAME} | awk -F'.' '{print $1}')

ip link set up $master_ifname
ip addr flush dev $master_ifname

ip addr flush $IFNAME && \
ip link set down $IFNAME && \
ip link del $IFNAME

ip link del $BRNAME