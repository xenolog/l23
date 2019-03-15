# Test case

* initial state:
  * Interfaces eth1, eth2 in the UP state and has no IP address
  * bond 'bond0' are not exists

* test #01:
  * create bond 'bond0'
  * add eth1 as port to bond 'bond0'
  * add eth2 as port to bond 'bond0'
  * add IP address to bond0

* test #02:
  * add secondary IP address to bond 'bond0'

* test #03:
  *
