// Copyright (c) 2019 by Matthew James Briggs, https://github.com/webern

package srv

import (
	"fmt"
	"net"

	"github.com/webern/flog"
	"github.com/webern/tftp/lib/cor"
	"github.com/webern/tftp/lib/stor"
)

// get transfers data from the store to a UDP TFTP Client
func get(hndshk handshake, store stor.Store) (conn *net.UDPConn, numBytes int, err error) {
	conn, err = net.DialUDP("udp", &hndshk.server, &hndshk.client)

	if err != nil {
		// attempt to send an error back to the client if the conn is not nil
		if conn != nil {
			// if this send fails we can drop the error because there's nothing more we can do with it
			_ = sendErr(conn, cor.ErrNotFound, err.Error())
		}

		// exit because we could not establish the connection
		return nil, 0, flog.Wrap(err)
	}

	// obtain the desired file from the store
	theFile := cor.File{}
	theFile.Name = hndshk.tftpInfo.Filename
	theFile, err = store.Get(theFile.Name)

	if err != nil {
		return conn, 0, cor.NewErr(cor.ErrNotFound, fmt.Sprintf("the file '%s' could not be found", hndshk.tftpInfo.Filename))
	}

	if err := sendHandshakeAck(conn); err != nil {
		return conn, 0, cor.NewErrf(cor.ErrUnknown, "acknowledgement packet could not be sent")
	}

	// set up for the sending loop
	numBytes = len(theFile.Data)
	blk := 1
	buf := packetPool.Get().([]byte)
	memset(buf)
	defer packetPool.Put(buf)
	sendEmptyAtEnd := len(theFile.Data)%cor.BlockSize == 0

	// loop through the file data with a stride length of BlockSize, send the data to the client
	for pos := 0; pos < len(theFile.Data); {
		end := pos + cor.BlockSize

		if end > len(theFile.Data) {
			end = len(theFile.Data)
		}

		err := sendDataPacket(hndshk, blk, conn, &theFile, pos, buf)

		if err != nil {
			return conn, 0, flog.Wrap(err)
		}

		blk++
		pos = end
	}

	// if the last chunk we sent above was of size BlockSize then the client needs an empty block to know we are done
	if sendEmptyAtEnd {
		err := sendEmptyEnding(hndshk, blk, conn, buf)

		if err != nil {
			return conn, 0, flog.Wrap(err)
		}
	}

	return conn, numBytes, nil
}

// sendEmptyEnding sends an empty block (to be used after the last chunk in the event that the last chunk aligned
// exactly with tftp's 512 byte chunk size).
func sendEmptyEnding(hndshk handshake, block int, conn *net.UDPConn, buf []byte) error {
	data := cor.PacketData{}
	data.BlockNum = uint16(block)
	data.Data = make([]byte, 0)
	_, err := conn.Write(data.Serialize())

	if err != nil {
		return err
	}

	// receive confirmation from client
	n, addr, err := conn.ReadFromUDP(buf)

	if err != nil {
		return err
	}

	// check that the client port matches our expected port
	if addr.Port != hndshk.client.Port {
		return flog.Raisef("wrong client port, got %d, want %d", addr.Port, hndshk.client.Port)
	}

	if n <= 0 {
		return flog.Raisef("bad acknowledgement packet")
	}

	packet, err := cor.ParsePacket(buf)

	if err != nil {
		return flog.Wrap(err)
	}

	if !packet.IsAck() {
		return flog.Raise("wrong packet type")
	}

	ack, ok := packet.(*cor.PacketAck)

	// if we were unable to downcast the packet then this is a fatal bug in the program
	if !ok {
		flog.Bug()
	}

	// check for the correct block number in the acknowledgement
	if ack.BlockNum != uint16(block) {
		return flog.Raisef("wrong block ack, got %d, want %d", ack.BlockNum, block)
	}

	return nil
}

func sendDataPacket(hndshk handshake, blk int, conn *net.UDPConn, theFile *cor.File, pos int, buf []byte) error {
	data := cor.PacketData{}
	data.BlockNum = uint16(blk)
	end := pos + cor.BlockSize

	if end > len(theFile.Data) {
		end = len(theFile.Data)
	}

	data.Data = theFile.Data[pos:end]
	_, err := conn.Write(data.Serialize())

	if err != nil {
		return err
	}

	n, addr, err := conn.ReadFromUDP(buf)

	if err != nil {
		return err
	}

	if addr.Port != hndshk.client.Port {
		return flog.Raisef("wrong client port, got %d, want %d", addr.Port, hndshk.client.Port)
	}

	if n <= 0 {
		return flog.Raisef("bad acknowledgement packet")
	}

	packet, err := cor.ParsePacket(buf)

	if err != nil {
		return flog.Wrap(err)
	}

	if !packet.IsAck() {
		return flog.Raise("wrong packet type")
	}

	ack, ok := packet.(*cor.PacketAck)

	if !ok {
		return flog.Raise("bug, could not downcast packet")
	}

	if ack.BlockNum != uint16(blk) {
		return flog.Raisef("wrong block ack, got %d, want %d", ack.BlockNum, blk)

	}

	return nil
}
