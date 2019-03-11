#!/bin/bash
set -x

IFNAMES=${IFNAME:-eth1 eth2}
BRNAME=${BRNAME:-br1}
BONDNAME=${BONDNAME:-bond1}

for i in $IFNAMES ; do
  ip addr flush $i
  ip link set down $i
done

ip link del $BONDNAME
ip link del $BRNAME