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
}

func run() error {
	flog.SetTruncationPath("tftp/")
	flog.SetLevel(flog.TraceLevel)
	srvr := srv.NewServer(stor.NewMemStore(), "tftp_conn.log")
	srvWait := sync.WaitGroup{}
	srvWait.Add(1)
	var srvErr error

	go func() {
		flog.InfoAlways("tftp server is starting on port %d", 69)
		srvErr = srvr.Serve(69)
		srvWait.Done()
	}()

	//err := srvr.Serve(69)
	//
	//if err != nil {
	//	flog.Error(err.Error())
	//	os.Exit(1)
	//}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	sig := <-sigChan
	if sig == syscall.SIGINT {
		fmt.Print("\n")
		flog.InfoAlways("SIGINT")
		err := srvr.Stop()
		if err != nil {
			return err
		}
	}

	srvWait.Wait()
	return srvErr
}
