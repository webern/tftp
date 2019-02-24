// Copyright (c) 2019 by Matthew James Briggs, https://github.com/webern

package main

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/webern/flog"
	"github.com/webern/tftp/lib/srv"
	"github.com/webern/tftp/lib/stor"
)

func main() {
	err := run()

	if err != nil {
		flog.Error(err.Error())
		os.Exit(1)
	}

	flog.InfoAlways("successful exit")
}

// run is the main logic of the tftpd program
func run() error {
	programArgs := parseArgs()
	flog.SetTruncationPath("tftp/")
	server := srv.NewServer(stor.NewMemStore())
	server.LogFilePath = programArgs.LogFilePath
	server.Port = programArgs.Port
	server.Verbose = programArgs.Verbose

	if server.Verbose {
		flog.SetLevel(flog.TraceLevel)
	} else {
		flog.SetLevel(flog.InfoLevel)
	}

	srvWait := sync.WaitGroup{}
	srvWait.Add(1)
	var srvErr error

	// Serve blocks until Stop is called, so we run it on its own goroutine
	// This function will exit once server.Stop is called
	go func() {
		flog.InfofAlways("tftp server is starting on port %d", server.Port)
		srvErr = server.Serve()
		srvWait.Done()
	}()

	// listen for sigint (i.e. control-c) to stop the server
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	sig := <-sigChan
	if sig == syscall.SIGINT {
		fmt.Print("\n")
		flog.InfoAlways("SIGINT received - stopping tftp server")
		err := server.Stop()
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
