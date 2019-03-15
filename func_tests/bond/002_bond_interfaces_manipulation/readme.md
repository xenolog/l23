# Test case

* initial state:
  * Interfaces eth1, eth2, eth3 in the UP state and has no IP address
  * bond 'bond0' are not exists

* test #01:
  * create bond 'bond0'
  * add eth1 as port to bond 'bond0'
  * add eth3 as port to bond 'bond0'
  * add IP address to bond0

* test #02:
  * remove eth3 from bond0
  * add eth2 as port to bond 'bond0'

* test #03:
  * add bond to the bridge 'br1'
  * move IP address from 'bond0' to bridge 'br1'
