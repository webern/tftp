// Copyright (c) 2019 by Matthew James Briggs, https://github.com/webern

package srv

import (
	"fmt"
	"github.com/webern/flog"
	"github.com/webern/tftp/lib/stor"
	"net"
	"os"
	"time"

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
	lch   chan LogEntry
	lfile string
}

func NewServer(store stor.Store, logFilename string) Server {
	return Server{
		store: store,
		lch:   make(chan LogEntry, 3),
		lfile: logFilename,
	}
}

func sendError(conn *net.UDPConn, theError error) error {
	pktErr := cor.PacketError{}
	pktErr.Code = cor.OpError
	pktErr.Msg = theError.Error()
	_, err := conn.Write(pktErr.Serialize())
	return err
}

func (s *Server) Serve(port uint16) error {
	go s.logAsync()
	mainListener, err := makeListener(port)

	if err != nil {
		return err
	}

	for {
		handshake, err := waitForHandshake(mainListener)
		l := LogEntry{
			Start: time.Now(),
		}

		if handshake.tftpInfo.IsWRQ() {
			go doAsyncTransfer(handshake, s.store, l, s.lch, put)
		} else if handshake.tftpInfo.IsRRQ() {
			go doAsyncTransfer(handshake, s.store, l, s.lch, get)
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
	close(s.lch)
	return nil
}

func (s *Server) logAsync() {
	defer flog.Trace("logging goroutine exit")
	lfile, err := os.Create(s.lfile)

	if err != nil {
		flog.Errorf("could not create log file: %s", err.Error())
		return
	}

	err = lfile.Close()
	lfile = nil

	if err != nil {
		flog.Errorf("could not close log file: %s", err.Error())
		return
	}

	for {
		le, ok := <-s.lch

		if !ok {
			return
		}

		lfile, err = os.OpenFile(s.lfile, os.O_APPEND|os.O_WRONLY, 0600)

		if err != nil {
			flog.Errorf("could not open log file: %s", err.Error())
			return
		}

		_, err = lfile.WriteString(fmt.Sprintf("%s\n", le.String()))

		if err != nil {
			flog.Errorf("could not close log file: %s", err.Error())
			return
		}
	}
}
