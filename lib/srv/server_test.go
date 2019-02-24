// Copyright (c) 2019 by Matthew James Briggs, https://github.com/webern

package srv

import (
	"sync"
	"testing"

	"github.com/webern/flog"
	"github.com/webern/tftp/lib/stor"
)

func startServer(port int, store stor.Store) (server *Server, wg *sync.WaitGroup) {
	s := NewServer(store)
	server = &s
	server.Port = port
	server.Verbose = true
	wg := &sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()
		err := server.Serve()
		flog.Error(err.Error())
	}()

	return server, wg
}

func TestServer(t *testing.T) {

	srvr := NewServer(stor.NewMemStore())
	srvr.Port = 9909
	go srvr.Serve()
	defer srvr.Stop()

	// TODO - run actual server tests
}
