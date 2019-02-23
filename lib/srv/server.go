// Copyright (c) 2019 by Matthew James Briggs, https://github.com/webern

package srv

import (
	"fmt"
	"github.com/webern/flog"
	"github.com/webern/tftp/lib/stor"
	"net"
	"os"
	"sync"
	"time"

	"github.com/webern/tftp/lib/cor"
)

// TftpMTftpMaxPacketSize is the practical limit of the size of a UDP
// packet, which is the size of an Ethernet MTU minus the headers of
// TFTP (4 bytes), UDP (8 bytes) and IP (20 bytes). (source: google).
const TftpMaxPacketSize = 1468

type Server struct {
	store  stor.Store
	lch    chan LogEntry
	lfile  string
	conn   *net.UDPConn
	stopMX sync.RWMutex // protects the stop boolean
	stop   bool         // tells the Serve function when it should bail out
}

func NewServer(store stor.Store, logFilename string) Server {
	s := Server{
		store: store,
		lch:   make(chan LogEntry, 3),
		lfile: logFilename,
	}
	return s
}

func sendError(conn *net.UDPConn, theError error) error {
	pktErr := cor.PacketError{}
	pktErr.Code = cor.OpError
	pktErr.Msg = theError.Error()
	_, err := conn.Write(pktErr.Serialize())
	return err
}

func (s *Server) Serve(port uint16) error {
	defer flog.Trace("stopped")
	go s.logAsync()
	mainListener, err := makeListener(port)

	if err != nil {
		return err
	} else if mainListener == nil {
		return flog.Raise("main listening connection could not be opened")
	}

	s.conn = mainListener

	for {

		s.stopMX.RLock()
		if s.stop {
			s.stopMX.RUnlock()
			return nil
		}
		s.stopMX.RUnlock()

		handshake, err := waitForHandshake(s.conn)

		s.stopMX.RLock()
		if s.stop {
			s.stopMX.RUnlock()
			return nil
		}
		s.stopMX.RUnlock()

		if err != nil {
			return err
		}

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
			return err
		}
	}

	return nil
}

func (s *Server) Stop() error {
	defer flog.Trace("stopped")
	var err error
	s.stopMX.Lock()
	defer s.stopMX.Unlock()
	s.stop = true

	if s.conn != nil {
		err = s.conn.Close()
		s.conn = nil
	}

	close(s.lch)

	if s.store != nil {
		s.store.Terminate()
	}

	return err
}

func (s *Server) logAsync() {
	defer flog.Trace("exit")
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
