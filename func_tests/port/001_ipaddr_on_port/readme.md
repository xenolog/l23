# Test case

* initial state:
  * Interface eth1 in the DOWN state and has no IP address

* test #01:
  * add IP address to eth1
  * put eth1 to UP state

* test #02:
  * add second IP address to eth1
  * note! here and below IP addresses (for multi-address cases) should be from different subnets, or secondary addresses should have /32 netmask. This limitation due native linux kernel network stack features found.

* test #03:
  * remove first IP address on eth1

* test #04:
  * change network mask on the single IP address on eth1

* test #05:
  * change single IP address on eth1

* test #06:
  * remove single IP address on eth1

* test #07:
  * add three IP addresses to eth1

* test #08:
  * remove all existing IP address on eth1
