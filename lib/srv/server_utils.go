// Copyright (c) 2019 by Matthew James Briggs, https://github.com/webern

package srv

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/webern/tftp/lib/stor"

	"github.com/webern/flog"
	"github.com/webern/tftp/lib/cor"
)

func makeListener(port uint16) (*net.UDPConn, error) {
	strAddr := fmt.Sprintf(":%d", port)
	uaddr, err := net.ResolveUDPAddr("udp", strAddr)

	if err != nil {
		return &net.UDPConn{}, err
	}

	uconn, err := net.ListenUDP("udp", uaddr)

	if err != nil {
		return &net.UDPConn{}, err
	}

	return uconn, err
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
func waitForHandshake(conn *net.UDPConn) (handshake, error) {
	buf := handshakePool.Get().([]byte)
	defer handshakePool.Put(buf)
	memset(buf)
	numBytes, ua, err := conn.ReadFromUDP(buf)

	if err != nil {
		return handshake{}, flog.Wrap(err)
	}

	if ua == nil || numBytes <= 0 {
		return handshake{}, flog.Raise("unable to receive the udp packet")
	}

	tftpInfo, err := parsePacket(buf)

	if err != nil {
		return handshake{}, err
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

func parsePacket(buf []byte) (*cor.PacketRequest, error) {
	pkt, err := cor.ParsePacket(buf)
	if err != nil {
		return nil, flog.Wrap(err)
	}
	if pkt == nil {
		return nil, flog.Raise("nil packet received from ParsePacket")
	}
	if !pkt.IsRRQ() && !pkt.IsWRQ() {
		return nil, flog.Raisef("bad op value: %d", pkt.Op())
	}
	tftpInfo, ok := pkt.(*cor.PacketRequest)

	if !ok || tftpInfo == nil {
		return nil, flog.Raise("unable to downcast the packaet to the correct type")
	}

	return tftpInfo, nil
}

// memset sets all bytes to zero
func memset(b []byte) {
	for i := 0; i < len(b); i++ {
		b[i] = 0
	}
}

// transferFunction is a type alias for get and put, which both share the logic in doAsyncTransfer
type transferFunction = func(hndshk handshake, store stor.Store) (conn *net.UDPConn, numBytes int, err error)

// doAsyncTransfer wraps both the get and put functions with error handling and logging stuff
func doAsyncTransfer(hndshk handshake, store stor.Store, l LogEntry, lch chan<- LogEntry, f transferFunction) {
	conn, n, err := f(hndshk, store)

	if err != nil {
		switch e := err.(type) {
		case *cor.Err:
			{
				if conn != nil {
					_ = e.Send(conn)
				}

				l.Error = e
			}
		default:
			{
				wr := cor.NewErrWrap(e)
				if conn != nil {
					_ = wr.Send(conn)
				}

				l.Error = wr
			}
		}
	} else {
		l.Bytes = n
	}

	l.Duration = time.Since(l.Start)
	l.Client = hndshk.client
	l.File = hndshk.tftpInfo.Filename
	l.Op = hndshk.tftpInfo.Op()
	lch <- l
}
