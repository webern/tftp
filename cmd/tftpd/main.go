package main

import (
	"fmt"
	"github.com/webern/tftp/tftplib/srv"
	"os"
)

func main() {
	fmt.Println("Hello world!")
	srvr := tftpsrv.NewServer()
	err := srvr.Serve(":9909")

	if err != nil {
		fmt.Print(err.Error())
		os.Exit(1)
	}

	os.Exit(0)
}
