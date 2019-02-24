// Copyright (c) 2019 by Matthew James Briggs, https://github.com/webern

package main

import (
	"flag"

	"github.com/webern/flog"
)

// ProgramArgs represents the command line arguments after they have been parsed
type ProgramArgs struct {
	LogFilePath string // LogFilePath tells the server where to write the connection log
	Port        int    // The listening port, defaults to 69 per TFTP standard
	Verbose     bool   // Sets the stdout logging to 'trace'. Does not affect the connection log
}

func parseArgs() ProgramArgs {
	a := ProgramArgs{}
	flag.StringVar(&a.LogFilePath, "logfile", "", "where to write the connection log. If you do not want the connections logged to a file then leave it blank and connection logs will be written to stdout instead.")
	flag.IntVar(&a.Port, "port", 69, "the port the tftp server should listen on")
	flag.BoolVar(&a.Verbose, "verbose", false, "increase the verbosity of logging to stdout. does not affect the ")
	flag.Parse()
	flog.Infof("args %v", a)
	return a
}

// https://stackoverflow.com/a/54747682/2779792
func isFlagPassed(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}
