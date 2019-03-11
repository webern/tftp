// Copyright (c) 2019 by Matthew James Briggs, https://github.com/webern

package srv

import (
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"github.com/webern/flog"
	"github.com/webern/tftp/lib/stor"

	"github.com/webern/tftp/lib/cor"
)

// TftpMaxPacketSize is the practical limit of the size of a UDP
// packet, which is the size of an Ethernet MTU minus the headers of
// TFTP (4 bytes), UDP (8 bytes) and IP (20 bytes). (source: google).
const TftpMaxPacketSize = 1468

const retries = 3
const timeout = 3 * time.Second
const logChanDepth = 3

// Server listens and responds to UDP TFTP Requests
type Server struct {
	// these data members should be set before calling Serve

	// LogFilePath tells the server where to write the connection log. If you do not want the connections
	// logged to a file then leave it blank and connection logs will be written to stdout instead.
	LogFilePath string

	Port    int            // The listening port, defaults to 69 per TFTP standard
	Verbose bool           // Sets the stdout logging to 'trace'. Does not affect the connection log
	store   stor.Store     // stores and retrieves files by name
	lch     chan LogEntry  // log entries will be sent to this channel for the connection log
	conn    *net.UDPConn   // is nil until Serve is called
	stopMX  *sync.RWMutex  // protects the stop boolean
	stop    bool           // tells the Serve function when it should bail out
	wg      sync.WaitGroup // tells us if we have running transfers
}

// NewServer creates a new TFTP server. The Store is injected.
// After NewServer, you should set Port and Verbose if you do not want the defaults.
func NewServer(store stor.Store) Server {
	s := Server{
		Port:    69,
		Verbose: false,
		store:   store,
		lch:     make(chan LogEntry, logChanDepth),
		conn:    nil,
		stopMX:  new(sync.RWMutex),
		stop:    false,
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

// Serve listens for incoming UDP TFTP connections and responds to them. Serve blocks until server.Stop is called
// by another goroutine. It is recommended to run Serve in its own goroutine due to its blocking nature.
func (s *Server) Serve() error {
	defer flog.Trace("stopped")
	go s.logAsync()
	mainListener, err := makeListener(uint16(s.Port))

	if err != nil {
		return err
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

			if s.conn != nil {
				e := cor.NewErr(cor.ErrNotFound, "nope")
				err = e.Send(s.conn)
				if err != nil {
					flog.Error(err)
				}
			}

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
			// increment the number of active connections
			flog.Info("wg add")
			flog.Info(&s.wg)
			s.wg.Add(1)
			go doAsyncTransfer(handshake, s.store, l, s.lch, put, s.wg)
		} else if handshake.tftpInfo.IsRRQ() {
			// increment the number of active connections
			flog.Info("wg add")
			flog.Info(&s.wg)
			s.wg.Add(1)
			go doAsyncTransfer(handshake, s.store, l, s.lch, get, s.wg)
		} else {
			// increment the number of active connections
			flog.Info("wg add")
			flog.Info(&s.wg)
			s.wg.Add(1)
			go s.sendBadOp(handshake, s.wg)
		}

		if err != nil {
			return err
		}
	}
	// unreachable
}

func (s *Server) sendBadOp(h handshake, wg sync.WaitGroup) {
	defer func() {
		flog.Info("wg.done")
		wg.Done()
	}()
	conn, err := net.DialUDP("udp", &h.server, &h.client)

	if err != nil {
		flog.Error(err.Error())
		return
	}

	err = sendErr(conn, cor.ErrBadOp, "")

	if err != nil {
		flog.Error(err.Error())
		return
	}
}

// Stop will stop the server and cause server.Serve() to exit
func (s *Server) Stop() error {
	defer flog.Trace("stopped")
	var err error
	s.stopMX.Lock()
	defer s.stopMX.Unlock()
	s.stop = true
	s.wg.Wait()

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

// logAsync runs on its own goroutine, receiving and writing connection logs
func (s *Server) logAsync() {
	defer flog.Trace("exit")

	// create the file, will be appended with each log entry
	if len(s.LogFilePath) > 0 {
		_ = s.createLogFile()
	}

	// receive log entries on channel, exit when channel is closed
	for {
		le, ok := <-s.lch

		if !ok {
			return
		}

		s.writeLog(le)
	}
}

func (s *Server) writeLog(le LogEntry) {
	if s.Verbose || len(s.LogFilePath) == 0 {
		flog.Trace(le.String())
	}

	if len(s.LogFilePath) == 0 {
		return
	}

	lfile, err := os.OpenFile(s.LogFilePath, os.O_APPEND|os.O_WRONLY, 0600)

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

func (s *Server) createLogFile() error {
	lfile, err := os.Create(s.LogFilePath)

	if err != nil {
		flog.Errorf("could not create log file: %s", err.Error())
		return nil
	}

	err = lfile.Close()

	if err != nil {
		flog.Errorf("could not close log file: %s", err.Error())
		return err
	}

	return nil
}
