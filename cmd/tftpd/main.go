// Copyright (c) 2019 by Matthew James Briggs, https://github.com/webern

package main

import (
	"os"

	"github.com/webern/flog"
)

func main() {
	err := run(make(chan os.Signal, 1))

	if err != nil {
		flog.Error(err.Error())
		os.Exit(1)
	}

	flog.InfoAlways("successful exit")
}
