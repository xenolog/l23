# Test case

* initial state:
  * Interface eth1 (master) in the UP state and has no IP address
  * Vlan interface eth1.101 not found

* test #01:
  * create eth1.101
  * add IP address to eth1.101
  * put eth1.101 to UP state

* test #02:
  * add second IP address (with netmask /32) to eth1.101

* test #03:
  * remove first IP address on eth1.101
