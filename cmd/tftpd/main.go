// Copyright (c) 2019 by Matthew James Briggs, https://github.com/webern

package main

import (
	"github.com/webern/flog"
	"github.com/webern/tftp/lib/srv"
	"github.com/webern/tftp/lib/stor"
	"os"
)

func main() {
	flog.SetTruncationPath("tftp/")
	flog.SetLevel(flog.TraceLevel)
	srvr := srv.NewServer(stor.NewMemStore())
	err := srvr.Serve(69)

	if err != nil {
		flog.Error(err.Error())
		os.Exit(1)
	}

	os.Exit(0)
}
