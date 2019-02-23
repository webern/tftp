// Copyright (c) 2019 by Matthew James Briggs, https://github.com/webern

package srv

import (
	"fmt"
	"net"
	"sync"

	"github.com/webern/flog"
	"github.com/webern/tftp/lib/cor"
)

func makeListener(port uint16) (net.UDPConn, error) {
	strAddr := fmt.Sprintf(":%d", port)
	uaddr, err := net.ResolveUDPAddr("udp", strAddr)

	if err != nil {
		return net.UDPConn{}, err
	}

	uconn, err := net.ListenUDP("udp", uaddr)

	if err != nil {
		return net.UDPConn{}, err
	}

	return *uconn, err
}

var handshakePool = sync.Pool{
	// New creates an object when the pool has nothing available to return.
	// New must return an interface{} to make it flexible. You have to cast
	// your type after getting it.
	New: func() interface{} {
		// Pools often contain things like *bytes.Buffer, which are
		// temporary and re-usable.
		return make([]byte, TftpMaxPacketSize)
	},
}

// waitForHandshake parses a UDP packet from the conn. if no acceptable packet is received, returns ok == false
// if an acceptable packet is received, a handshake object is returned representing the client's declared address (i.e.
// port) for the transfer and the requested operation. the server's declared address (i.e. port) for the transfer is
// not established by this function.
func waitForHandshake(conn net.UDPConn) (handshake, error) {
	buf := handshakePool.Get().([]byte)
	defer handshakePool.Put(buf)
	memset(buf)
	numBytes, ua, err := conn.ReadFromUDP(buf)

	if err != nil {

		return handshake{}, flog.Wrap(err)
	} else if ua == nil || numBytes <= 0 {
		return handshake{}, flog.Raise("unable to receive the udp packet")
	}

	pkt, err := cor.ParsePacket(buf)

	if err != nil {
		return handshake{}, flog.Wrap(err)
	} else if pkt == nil {
		return handshake{}, flog.Raise("nil packet received from wire.ParsePacket")
	}

	if !pkt.IsRRQ() && !pkt.IsWRQ() {
		return handshake{}, flog.Raisef("bad op value: %d", pkt.Op())
	}

	tftpInfo, ok := pkt.(*cor.PacketRequest)

	if !ok || tftpInfo == nil {
		return handshake{}, flog.Raise("unable to downcast the packaet to the correct type")
	}

	serverAddress, err := net.ResolveUDPAddr("udp", ":0")

	if err != nil || serverAddress == nil {
		return handshake{}, flog.Raise("unable to resolve my server address")
	}

	handshk := handshake{}
	handshk.tftpInfo = *tftpInfo
	handshk.client = *ua
	handshk.server = *serverAddress
	return handshk, nil
}

func memset(b []byte) {
	for i := 0; i < len(b); i++ {
		b[i] = 0
	}
}
