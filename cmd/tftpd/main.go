// Copyright (c) 2019 by Matthew James Briggs, https://github.com/webern

package main

import (
	"github.com/webern/flog"
	"github.com/webern/tftp/lib/srv"
	"os"
)

func main() {
	flog.SetTruncationPath("tftp/")
	flog.SetLevel(flog.TraceLevel)
	srvr := srv.NewServer()
	err := srvr.Serve(9909)

	if err != nil {
		flog.Error(err.Error())
		os.Exit(1)
	}

	os.Exit(0)
}
