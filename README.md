In-memory TFTP Server
=====================

master: [![CircleCI](https://circleci.com/gh/webern/tftp/tree/master.svg?style=svg)](https://circleci.com/gh/webern/tftp/tree/master)

develop: [![CircleCI](https://circleci.com/gh/webern/tftp/tree/develop.svg?style=svg)](https://circleci.com/gh/webern/tftp/tree/develop)

This is a simple in-memory TFTP server, implemented in Go.  It is
RFC1350-compliant, but doesn't implement the additions in later RFCs.  In
particular, options are not recognized.

TODO
----

Remember to commit flog, tcore, etc. to master so that go get will retreive working versions

Usage
-----
TODO

Testing
-------
TODO

TODO: other relevant documentation

Inspiration / Research
----------------------

  * [Detailed Description of UDP in Go](https://ops.tips/blog/udp-client-and-server-in-go/)
  * [TFTP Client Usage on macOS](http://www.i-helpdesk.com.au/index.php?/Knowledgebase/Article/View/721/0/how-to-upgrade-router-firmware-via-tftp-using-mac-os-devices)
  * [A Server Implementation](https://github.com/shriganeshs/golang-tftp-server/blob/master/main.go)


