// Copyright (c) 2019 by Matthew James Briggs, https://github.com/webern

package tftpsrv

import (
	"testing"
)

func TestServer(t *testing.T) {

	srvr := NewServer()
	go srvr.Serve(9909)
	defer srvr.Stop()

	//if err != nil {
	//	t.Errorf("error received from server.Serve: %s", err.Error())
	//	return
	//}

}
