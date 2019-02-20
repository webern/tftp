package tftpsrv

import (
	"fmt"
	"github.com/webern/tftp/tftplib/wire"
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
		buf := make([]byte, TftpMaxPacketSize) // TODO: sync.Pool
		numBytes, ua, err := uconn.ReadFromUDP(buf)

		if err != nil {
			return err
		} else if ua == nil {
			continue listeningLoop
		} else if numBytes <= 0 {
			fmt.Printf("error no bytes addr %v\n", *ua)
			continue
		}

		pkt, err := wire.ParsePacket(buf)

		if err != nil {
			panic("x")
		} else if pkt == nil {
			panic("y")
		}

		log.Printf("%v", pkt)
		log.Printf("New Connection from %s!", ua)
		log.Print(numBytes)
		log.Print(ua)

		if ua.Port == 1 {
			break listeningLoop
		}

		if req, ok := pkt.(*wire.PacketRequest); ok && req != nil {
			fmt.Printf("filename: %s", req.Filename)

			wrq_addr, err := net.ResolveUDPAddr("udp", ":0")
			if err != nil {
				fmt.Println("Error resolving UDP address: ", err)
				panic(err)
			}
			wrq_conn, err := net.DialUDP("udp", wrq_addr, ua)
			if err != nil {
				fmt.Println("Error dialing UDP to client: ", err)
				panic(err)
			}

			// send first ack for wrq with block #0
			ak := wire.PacketAck{}
			ak.BlockNum = 0
			_, err = wrq_conn.Write(ak.Serialize())

			if err != nil {
				panic(err)
			}

		}

	}

	return err
}

func (s *Server) Stop() error {
	return nil
}
