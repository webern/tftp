// Copyright (c) 2019 by Matthew James Briggs, https://github.com/webern

package tfsrv

import (
	"fmt"
	"github.com/webern/flog"
	"github.com/webern/tftp/lib/tfcore"
	"net"
	"time"
)

func put(hndshk handshake, ch chan<- tfcore.File, ready chan<- struct{}) error {
	conn, err := net.DialUDP("udp", &hndshk.server, &hndshk.client)
	file := tfcore.File{}
	file.Name = hndshk.tftpInfo.Filename
	file.Data = make([]byte, 0, 1024)

	if err != nil {
		return flog.Wrap(err)
	}

	if err := sendHandshakeAck(conn); err != nil {
		return flog.Wrap(err)
	}

	// the first block is always 1, the 0 block is the acknowledgement
	blk := 1

	var pkt_buf []byte = make([]byte, tfcore.MaxPacketSize)
	retry_cnt := 0
	didSignalReady := false
dataLoop:
	for {
		// set timeout of 1 sec for receiving packet from client
		err := conn.SetReadDeadline(time.Now().Add(10000 * time.Second))

		if err != nil {
			return err
		}

		if ready != nil && !didSignalReady {
			// let our caller know we are ready to receive packets
			didSignalReady = true
			ready <- struct{}{}

		}

		n, raddr, err := conn.ReadFromUDP(pkt_buf)
		if err != nil {
			e_tout, status := err.(net.Error)
			if status && e_tout.Timeout() {
				if retry_cnt >= 3 {
					fmt.Println("Timeout while reading data from connection.")
					return flog.Raise("poo")
				}
				// resend the previous response
				conn.Write(nil)
				retry_cnt = retry_cnt + 1
				continue
			}
			fmt.Println("Error reading data from connection: ", err)
			err := flog.Raise("poo")
			_ = sendError(conn, err)
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

		if packet.Op() != tfcore.OpData {
			flog.Errorf("wrong op type %d", packet.Op())
			flog.Bug()
		} else if tfcore.Block(packet) != uint16(blk) {
			flog.Errorf("wrong block, want %d, got %d", uint16(blk), tfcore.Block(packet))
			flog.Bug()
		} else if hndshk.client.Port != raddr.Port {
			flog.Bug()
		}

		// check if appropriate sequence data packet is received and store it in linked-list
		if packet.Op() == tfcore.OpData &&
			tfcore.Block(packet) == uint16(blk) &&
			hndshk.client.Port == raddr.Port {
			dataPacket, ok := packet.(*tfcore.PacketData)
			if !ok {
				return flog.Raise("poo")
			}
			raw_data := make([]byte, len(dataPacket.Data)+4)
			copy(raw_data, dataPacket.Data)
			file.Data = append(file.Data, raw_data...)
			err := sendAck(conn, blk)
			if err != nil {
				flog.Error("what should i do with this error? %s", err.Error())
			}
			blk++

			// check if this is the last received data packet
			if len(dataPacket.Data) < 512 {
				flog.Trace("File transfer completed! Received file")
				break dataLoop
			}
		} else {
			flog.Bug()
		}
	}

	ch <- file
	return nil
}

func sendHandshakeAck(conn *net.UDPConn) error {
	return sendAck(conn, 0)
}

func sendAck(conn *net.UDPConn, block int) error {
	ack := tfcore.PacketAck{}
	ack.BlockNum = uint16(block)
	_, err := conn.Write(ack.Serialize())

	if err != nil {
		return err
	}

	return nil
}
