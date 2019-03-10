# Test case

* initial state:
  * Interface eth1in the UP state and has no IP address
  * bridges br1, br2 not found

* test #01:
  * create bridge br1
  * add eth1 as port to bridge br1

* test #02:
  * create bridge br2
  * remove eth1 from br1 (implicitly)
  * add eth1 as port to bridge br2

* test #03:
  * remove eth1 from all bridges
  * add ip address to eth1 directly
