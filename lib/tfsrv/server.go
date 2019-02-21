// Copyright (c) 2019 by Matthew James Briggs, https://github.com/webern

package tfsrv

import (
	"fmt"
	"github.com/webern/flog"
	"github.com/webern/tftp/lib/tfcore"
	"net"
	"time"
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
	parsedPayload tfcore.Packet
	numBytes      int
}

func newUDPPacket(data []byte, bytes_recv int) (packet udpPacket, err error) {
	packet.numBytes = bytes_recv
	return packet, nil
}

type Server struct {
}

func NewServer() Server {
	return Server{}
}

func sendError(conn *net.UDPConn, theError error) error {
	pktErr := tfcore.PacketError{}
	pktErr.Code = tfcore.OpError
	pktErr.Msg = theError.Error()
	_, err := conn.Write(pktErr.Serialize())
	return err
}

func (s *Server) Serve(port uint16) error {

	mainListener, err := makeListener(port)

	if err != nil {
		return err
	}

	err = nil
	theFile := make([]byte, 0, 1024)

listeningLoop:
	for {
		//buf := make([]byte, TftpMaxPacketSize) // TODO: sync.Pool
		//numBytes, ua, err := mainListener.ReadFromUDP(buf)
		handshake, err := waitForHandshake(mainListener)

		if err != nil {
			// TODO - log to the connection log
			flog.Error(err.Error())
			continue
		}

		//if ok {
		flog.Tracef("filename: %s", handshake.tftpInfo.Filename)

		wrq_conn, err := net.DialUDP("udp", &handshake.server, &handshake.client)
		if err != nil {
			flog.Errorf("Error dialing UDP to client: ", err.Error())
			panic(err)
		}

		// send first ack for wrq with block #0
		ak := tfcore.PacketAck{}
		ak.BlockNum = 0
		_, err = wrq_conn.Write(ak.Serialize())

		if err != nil {
			panic(err)
		}

		ak.BlockNum++
		var pkt_buf []byte = make([]byte, tfcore.MaxPacketSize)
		retry_cnt := 0
	dataLoop:
		for {
			// set timeout of 1 sec for receiving packet from client
			wrq_conn.SetReadDeadline(time.Now().Add(1 * time.Second))
			n, raddr, err := wrq_conn.ReadFromUDP(pkt_buf)
			if err != nil {
				e_tout, status := err.(net.Error)
				if status && e_tout.Timeout() {
					if retry_cnt >= 3 {
						fmt.Println("Timeout while reading data from connection.")
						return flog.Raise("poo")
					}
					// resend the previous response
					wrq_conn.Write(nil)
					retry_cnt = retry_cnt + 1
					continue
				}
				fmt.Println("Error reading data from connection: ", err)
				err := flog.Raise("poo")
				_ = sendError(wrq_conn, err)
				return err
			}
			packet, err := tfcore.ParsePacket(pkt_buf[:n])
			if err != nil {
				if packet.Op() == tfcore.OpError {
					if pErr, ok := packet.(*tfcore.PacketError); ok {
						return flog.Raisef("Error received from client - error: ", pErr.Msg)
					} else {
						return flog.Raise("Error received from client")
					}
				}
				flog.Errorf("Encountered error while parsing the tftp packet: %s", err.Error())
				continue
			}

			// check if appropriate sequence data packet is received and store it in linked-list
			if packet.Op() == tfcore.OpData &&
				tfcore.Block(packet) == ak.BlockNum &&
				handshake.client.Port == raddr.Port {
				dataPacket, ok := packet.(*tfcore.PacketData)
				if !ok {
					return flog.Raise("poo")
				}
				raw_data := make([]byte, len(dataPacket.Data)+4)
				copy(raw_data, dataPacket.Data)
				theFile = append(theFile, raw_data...)

				_, err = wrq_conn.Write(ak.Serialize())
				retry_cnt = 0 // reset retry count
				ak.BlockNum++

				// check if this is the last received data packet
				if len(dataPacket.Data) < 512 {
					flog.Trace("File transfer completed! Received file")
					break dataLoop
				}
			}
		}

		//}
		if len(theFile) > 0 {
			break listeningLoop
		}
	}

	fmt.Print("\n\n")
	fmt.Print(string(theFile))
	return err
}

func (s *Server) Stop() error {
	return nil
}
