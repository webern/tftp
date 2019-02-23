// Copyright (c) 2019 by Matthew James Briggs, https://github.com/webern

package srv

import (
	"github.com/webern/tftp/lib/stor"
	"testing"
)

func TestServer(t *testing.T) {

	srvr := NewServer(stor.NewMemStore())
	go srvr.Serve(9909)
	defer srvr.Stop()

	// TODO - run actual server tests
}
