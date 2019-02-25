// Copyright (c) 2019 by Matthew James Briggs, https://github.com/webern

package main

import (
	"flag"
)

// ProgramArgs represents the command line arguments after they have been parsed
type ProgramArgs struct {
	LogFilePath string // LogFilePath tells the server where to write the connection log
	Port        int    // The listening port, defaults to 69 per TFTP standard
	Verbose     bool   // Sets the stdout logging to 'trace'. Does not affect the connection log
	Quiet       bool   // Sets the stdout logging to 'error'. Does not affect the connection log
}

func parseArgs() ProgramArgs {
	a := ProgramArgs{}
	flag.StringVar(&a.LogFilePath, "logfile", "", "where to write the connection log. If you do not want the connections logged to a file then leave it blank and connection logs will be written to stdout instead.")
	flag.IntVar(&a.Port, "port", 69, "the port the tftp server should listen on")
	flag.BoolVar(&a.Verbose, "verbose", false, "increase the verbosity of logging to stdout. does not affect the connection logfile")
	flag.BoolVar(&a.Quiet, "quiet", false, "decrease the verbosity of logging to stdout. does not affect the connection logfile")
	flag.Parse()
	return a
}
