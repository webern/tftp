In-memory TFTP Server
=====================

master: [![CircleCI](https://circleci.com/gh/webern/tftp/tree/master.svg?style=svg)](https://circleci.com/gh/webern/tftp/tree/master)

develop: [![CircleCI](https://circleci.com/gh/webern/tftp/tree/develop.svg?style=svg)](https://circleci.com/gh/webern/tftp/tree/develop)

This is a simple in-memory TFTP server, implemented in Go.  It is
RFC1350-compliant, but doesn't implement the additions in later RFCs.  In
particular, options are not recognized.

It always operates in `Octet` mode, and ignores any mode string.

Installation and Build
----------------------

The project is using GOPATH, not modules. For best results, get the project with:

`go get github.com/webern/tftp`

Then, with your GOPATH properly set, cd into the `tftp` repo:

`cd $GOPATH/src/github.com/webern/tftp`

Then, from the root of the `tftp` repo, get dependencies:

`go get ./...`

Then, to build the `tftpd` program from the root of the `tftp` repo:

`go build -o ./build/tftpd github.com/webern/tftp/cmd/tftpd`

Usage
-----

Having build the `tftpd` binary per the above instructions, from the root of the `tftp` repo, run the binary with:

`./build/tftpd --logfile="./build/connection.log" --port=69 --verbose"`

You may now send and receive files to/from the `tftpd` server.

To stop the server, use control-c to send a sigint.


TODO
----

Remember to commit flog, tcore, etc. to master so that go get will retreive working versions



Testing
-------
To run all tests, from the root of the repo:

`go test -v ./...`

To run all tests and get a code coverage report, from the root of the repo:

`go test -v ./... -coverprofile cover.out && go tool cover -func cover.out`

Additionally, tests are running on CircleCI, [here](https://circleci.com/gh/webern/tftp).

Inspiration / Research
----------------------

  * [Detailed Description of UDP in Go](https://ops.tips/blog/udp-client-and-server-in-go/)
  * [TFTP Client Usage on macOS](http://www.i-helpdesk.com.au/index.php?/Knowledgebase/Article/View/721/0/how-to-upgrade-router-firmware-via-tftp-using-mac-os-devices)
  * [A Server Implementation](https://github.com/shriganeshs/golang-tftp-server/blob/master/main.go)


