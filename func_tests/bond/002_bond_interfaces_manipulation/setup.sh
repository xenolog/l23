#!/bin/bash
set -x

IFNAMES=${IFNAMES:-eth1 eth2 eth3}
BRNAME=${BRNAME:-br1}
BONDNAME=${BONDNAME:-bond0}

for i in $IFNAMES ; do
  ip addr flush $i
  ip link set down $i
done

ip link del $BONDNAME
ip link del $BRNAME