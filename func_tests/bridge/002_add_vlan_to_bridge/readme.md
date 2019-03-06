# Test case

* initial state:
  * Interface eth1 (master) in the UP state and has no IP address
  * Vlan interface eth1.101 not found
  * bridge br1 not found

* test #01:
  * create bridge br1
  * create eth1.101
  * add eth1.101 as port to bridge br1
  * add IP address to br1

* test #02:
  * add second IP address (with netmask /32) to br1

* test #03:
  * remove first IP address on br1
