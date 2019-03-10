#!/bin/bash
set -x

IFNAME=${IFNAME:-eth1}
BRNAMES=${BRNAMES:-br1 br2}

ip addr flush $IFNAME
ip link set down $IFNAME

for i in $BRNAMES ; do ip link del $i ; done