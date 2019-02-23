// Copyright (c) 2019 by Matthew James Briggs, https://github.com/webern

package main

import (
	"fmt"
	"github.com/webern/flog"
	"github.com/webern/tftp/lib/srv"
	"github.com/webern/tftp/lib/stor"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func main() {
	err := run()

	if err != nil {
		flog.Error(err.Error())
		os.Exit(1)
	}

	flog.InfoAlways("successful exit")
}

func run() error {
	flog.SetTruncationPath("tftp/")
	flog.SetLevel(flog.TraceLevel)
	srvr := srv.NewServer(stor.NewMemStore(), "tftp_conn.log")
	srvWait := sync.WaitGroup{}
	srvWait.Add(1)
	var srvErr error

	// Serve blocks until Stop is called, so we run it on its own goroutine
	// This function will exit once srvr.Stop is called
	go func() {
		flog.InfoAlways("tftp server is starting on port %d", 69)
		srvErr = srvr.Serve(69)
		srvWait.Done()
	}()

	// listen for sigint (i.e. control-c) to stop the server
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	sig := <-sigChan
	if sig == syscall.SIGINT {
		fmt.Print("\n")
		flog.InfoAlways("SIGINT received - stopping tftp server")
		err := srvr.Stop()
		if err != nil {
			// TODO - we lose the Serve function's error, if any, in this case
			return err
		}
	}

	// wait for the Serve goroutine to stop
	srvWait.Wait()

	// return error if one was received from the Serve function
	return srvErr
}
