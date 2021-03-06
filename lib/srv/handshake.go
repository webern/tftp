// Copyright (c) 2019 by Matthew James Briggs, https://github.com/webern

package srv

import (
	"net"

	"github.com/webern/tftp/lib/cor"
)

// handshake represents a handshake between the client and the server. it contains the client's port number, the
// server's port number, and the operation type
type handshake struct {
	tftpInfo cor.PacketRequest
	client   net.UDPAddr // the client's declared port for the transfer
	server   net.UDPAddr // the server's declared port for the transfer
}
