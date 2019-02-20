package tftpsrv

import (
	"log"
	"net"
)

// TftpMTftpMaxPacketSize is the practical limit of the size of a UDP
// packet, which is the size of an Ethernet MTU minus the headers of
// TFTP (4 bytes), UDP (8 bytes) and IP (20 bytes). (source: google).
const TftpMaxPacketSize = 1468

type Server struct {
}

func NewServer() Server {
	return Server{}
}

func (s *Server) Serve(addr string) error {
	uaddr, err := net.ResolveUDPAddr("udp", addr)

	if err != nil {
		return err
	}

	uconn, err := net.ListenUDP("udp", uaddr)

	if err != nil {
		return err
	}

	err = nil

listeningLoop:
	for {
		buf := make([]byte, 0, TftpMaxPacketSize) // TODO: sync.Pool
		n, ua, err := uconn.ReadFromUDP(buf)
		if err != nil {
			return err
		} else if ua == nil {
			continue listeningLoop
		}

		log.Printf("New Connection from %s!", ua)
		log.Print(n)
		log.Print(ua)

		if ua.Port == 1 {
			break listeningLoop
		}
	}

	return err
}

func (s *Server) Stop() error {
	return nil
}
