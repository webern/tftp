// Copyright (c) 2019 by Matthew James Briggs, https://github.com/webern

package srv

import (
	"io"
	"net"
	"sync"
	"time"

	"github.com/webern/tftp/lib/stor"

	"github.com/webern/flog"
	"github.com/webern/tftp/lib/cor"
)

var packetPool = sync.Pool{
	New: func() interface{} {
		return make([]byte, cor.MaxPacketSize)
	},
}

type TransferFunc = func(hndshk handshake, store stor.Store) (conn *net.UDPConn, numBytes int, err error)

func doAsyncTransfer(hndshk handshake, store stor.Store, l LogEntry, lch chan<- LogEntry, f TransferFunc) {
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

func put(hndshk handshake, store stor.Store) (conn *net.UDPConn, numBytes int, err error) {
	conn, err = net.DialUDP("udp", &hndshk.server, &hndshk.client)

	if err != nil {
		return nil, 0, flog.Wrap(err)
	}

	theFile := cor.File{}
	theFile.Name = hndshk.tftpInfo.Filename
	theFile.Data = make([]byte, 0)

	if err := sendHandshakeAck(conn); err != nil {
		return conn, 0, flog.Wrap(err)
	}

	// block 0 is the acknowledgement, block 1 is the first data block
	blk := 1

	buf := packetPool.Get().([]byte)
	memset(buf)
	defer packetPool.Put(buf)

dataLoop:
	for {
		n, raddr, err := readWithRetry(conn, 3, buf, blk)

		if err != nil {
			return conn, 0, err
		}

		packet, err := cor.ParsePacket(buf[:n])

		if err != nil {
			return conn, 0, err
		}

		// check a bunch of possible error conditions
		err = verifyDataPacket(packet, hndshk, raddr)

		if err != nil {
			return conn, 0, err
		}

		chunk, err := handleData(conn, packet, blk)
		theFile.Data = append(theFile.Data, chunk...)

		if err == io.EOF {
			break dataLoop
		}

		blk++
	}

	numBytes = len(theFile.Data)
	err = store.Put(theFile)

	if err != nil {
		return conn, 0, err
	}

	return conn, numBytes, nil
}

func sendHandshakeAck(conn *net.UDPConn) error {
	return sendAck(conn, 0)
}

func sendAck(conn *net.UDPConn, block int) error {
	ack := cor.PacketAck{}
	ack.BlockNum = uint16(block)
	_, err := conn.Write(ack.Serialize())

	if err != nil {
		return err
	}

	return nil
}

func sendErr(conn *net.UDPConn, code cor.ErrCode, message string) error {
	ePacket := cor.PacketError{}
	ePacket.Code = code
	ePacket.Msg = message
	_, err := conn.Write(ePacket.Serialize())

	if err != nil {
		return err
	}

	return nil
}

func readWithRetry(conn *net.UDPConn, retries int, ioBuf []byte, lastSuccessfulBlock int) (numBytes int, raddr *net.UDPAddr, err error) {
	for retryCount := 0; retryCount <= retries; retryCount++ {
		err := conn.SetReadDeadline(time.Now().Add(timeout))

		if err != nil {
			return 0, nil, err
		}

		numBytes, raddr, err = conn.ReadFromUDP(ioBuf)

		if err == nil {
			// no error - return the results
			return numBytes, raddr, err
		}

		// an error condition exists, check if we can downcast it to net.Error
		netErr, ok := err.(net.Error)

		if !ok {
			// this is not a net.Error - terminate and notify the client that things are bad
			_ = sendError(conn, err)
			return numBytes, raddr, flog.Wrap(err)
		}

		if !netErr.Timeout() && !netErr.Temporary() {
			// this is not a recoverable error - terminate and notify the client things are bad
			_ = sendError(conn, err)
			return numBytes, raddr, flog.Wrap(netErr)
		}

		// notify the client that we want to retry
		err = sendAck(conn, lastSuccessfulBlock)

		if err != nil {
			// unable to communicate with the client - bail out
			err = flog.Raisef("lost communication with client: %s", err.Error())
			_ = sendError(conn, err)
			return numBytes, raddr, err
		}
	}

	err = flog.Raisef("tried connecting %d time(s) without success: %s", err.Error())
	_ = sendError(conn, err)
	return numBytes, raddr, err
}

func handleData(conn *net.UDPConn, packet cor.Packet, expectedBlock int) ([]byte, error) {
	dataPacket, ok := packet.(*cor.PacketData)

	if !ok {
		return nil, flog.Raise("the packet is not a data packet")
	}

	if dataPacket.BlockNum != uint16(expectedBlock) {
		return nil, flog.Raisef("wrong block num, got %d, want %d", dataPacket.BlockNum, expectedBlock)
	}

	copied := make([]byte, len(dataPacket.Data))
	copy(copied, dataPacket.Data)
	err := sendAck(conn, expectedBlock)

	if err != nil {
		return nil, flog.Raisef("acknowledgement could not be sent %s", err.Error())
	}

	// check if this is the last received data packet
	if len(dataPacket.Data) < cor.BlockSize {
		return copied, io.EOF
	}

	return copied, nil
}

func verifyDataPacket(packet cor.Packet, hndshk handshake, currentAddr *net.UDPAddr) error {
	if packet.IsError() {
		return flog.Raise("error received from client")
	} else if !packet.IsData() {
		return flog.Raisef("wrong op type %d", packet.Op())
	} else if currentAddr == nil {
		return flog.Raise("address is nil")
	} else if hndshk.client.Port != currentAddr.Port {
		return flog.Raisef("wrong port, want %d, got %d", currentAddr.Port, hndshk.client.Port)
	}

	return nil
}
