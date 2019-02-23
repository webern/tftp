// Copyright (c) 2019 by Matthew James Briggs, https://github.com/webern

package srv

import (
	"github.com/webern/flog"
	"github.com/webern/tftp/lib/stor"
	"net"

	"github.com/webern/tftp/lib/cor"
)

// TftpMTftpMaxPacketSize is the practical limit of the size of a UDP
// packet, which is the size of an Ethernet MTU minus the headers of
// TFTP (4 bytes), UDP (8 bytes) and IP (20 bytes). (source: google).
const TftpMaxPacketSize = 1468

// TID represents a 'transfer id' which is actually just two ports, the requester's port and the responder's port.
type TID struct {
	RequesterPort uint16
	ResponderPort uint16
}

type udpPacket struct {
	clientAddress *net.UDPAddr
	rawPayload    []byte
	parsedPayload cor.Packet
	numBytes      int
}

func newUDPPacket(data []byte, bytes_recv int) (packet udpPacket, err error) {
	packet.numBytes = bytes_recv
	return packet, nil
}

type Server struct {
	store stor.Store
}

func NewServer(store stor.Store) Server {
	return Server{
		store: store,
	}
	return Server{}
}

func sendError(conn *net.UDPConn, theError error) error {
	pktErr := cor.PacketError{}
	pktErr.Code = cor.OpError
	pktErr.Msg = theError.Error()
	_, err := conn.Write(pktErr.Serialize())
	return err
}

func (s *Server) Serve(port uint16) error {

	mainListener, err := makeListener(port)

	if err != nil {
		return err
	}

	for {
		handshake, err := waitForHandshake(mainListener)

		if handshake.tftpInfo.IsWRQ() {
			go doAsyncTransfer(handshake, s.store, put)
		} else if handshake.tftpInfo.IsRRQ() {
			go doAsyncTransfer(handshake, s.store, get)
		} else {
			go func() {
				conn, err := net.DialUDP("udp", &handshake.server, &handshake.client)
				if err != nil {
					flog.Error(err.Error())
					return
				}

				err = sendErr(conn, cor.ErrBadOp, "")

				if err != nil {
					flog.Error(err.Error())
					return
				}

			}()
		}

		if err != nil {
			panic(err)
		}

	}

	return nil
}

func (s *Server) Stop() error {
	return nil
}
