#!/bin/bash
set -x

IFNAME=${IFNAME:-eth1}
BRNAME=${BRNAME:-br1}
IPADDR=${IPADDR:-192.168.0.10/24}

# for i in $IFNAME $BRNAME ; do
#   ip addr flush $i
#   ip l set down $i
# done

ip l set down $BRNAME && ip l del $BRNAME

ip addr flush $IFNAME
ip l set up $IFNAME
ip addr add $IPADDR dev $IFNAME
