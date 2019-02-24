// Copyright (c) 2019 by Matthew James Briggs, https://github.com/webern

package srv

import (
	"testing"

	"github.com/webern/tftp/lib/stor"
)

func TestServer(t *testing.T) {

	srvr := NewServer(stor.NewMemStore())
	srvr.Port = 9909
	go srvr.Serve()
	defer srvr.Stop()

	// TODO - run actual server tests
}
