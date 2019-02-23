package srv

import (
	"github.com/webern/flog"
	"github.com/webern/tftp/lib/cor"
	"github.com/webern/tftp/lib/stor"
	"net"
)

func get(hndshk handshake, store stor.Store) (*net.UDPConn, error) {
	conn, err := net.DialUDP("udp", &hndshk.server, &hndshk.client)

	if err != nil {
		_ = sendErr(conn, cor.ErrNotFound, err.Error())
		return nil, flog.Wrap(err)
	}

	theFile := cor.File{}
	theFile.Name = hndshk.tftpInfo.Filename
	theFile, err = store.Get(theFile.Name)

	if err := sendHandshakeAck(conn); err != nil {
		return conn, cor.NewErrf(cor.ErrNotFound, "the file '%s' was not found", hndshk.tftpInfo.Filename)
	}

	// block 0 is the acknowledgement, block 1 is the first data block
	blk := 1

	buf := packetPool.Get().([]byte)
	memset(buf)
	defer packetPool.Put(buf)
	sendEmptyAtEnd := len(theFile.Data)%cor.BlockSize == 0

	for pos := 0; pos < len(theFile.Data); {
		data := cor.PacketData{}
		data.BlockNum = uint16(blk)
		end := pos + cor.BlockSize

		if end > len(theFile.Data) {
			end = len(theFile.Data)
		}

		data.Data = theFile.Data[pos:end]
		_, err := conn.Write(data.Serialize())

		if err != nil {
			flog.Error(err.Error())
		}

		n, addr, err := conn.ReadFromUDP(buf)
		if addr.Port != hndshk.client.Port {
			return conn, flog.Raisef("wrong client port, got %d, want %d", addr.Port, hndshk.client.Port)
		} else if n <= 0 {
			return conn, flog.Raisef("bad acknowledgement packet")
		}

		packet, err := cor.ParsePacket(buf)

		if err != nil {
			return conn, flog.Wrap(err)
		} else if !packet.IsAck() {
			return conn, flog.Raise("wrong packet type")
		}

		ack, ok := packet.(*cor.PacketAck)

		if !ok {
			flog.Bug()
		}

		if ack.BlockNum != uint16(blk) {
			return conn, flog.Raisef("wrong block ack, got %d, want %d", ack.BlockNum, blk)
		}

		blk++
		pos = end
	}

	if sendEmptyAtEnd {
		data := cor.PacketData{}
		data.BlockNum = uint16(blk)
		data.Data = make([]byte, 0)
		_, err = conn.Write(data.Serialize())

		if err != nil {
			flog.Error(err.Error())
		}

		n, addr, err := conn.ReadFromUDP(buf)
		if addr.Port != hndshk.client.Port {
			return conn, flog.Raisef("wrong client port, got %d, want %d", addr.Port, hndshk.client.Port)
		} else if n <= 0 {
			return conn, flog.Raisef("bad acknowledgement packet")
		}

		packet, err := cor.ParsePacket(buf)

		if err != nil {
			return conn, flog.Wrap(err)
		} else if !packet.IsAck() {
			return conn, flog.Raise("wrong packet type")
		}

		ack, ok := packet.(*cor.PacketAck)

		if !ok {
			flog.Bug()
		}

		if ack.BlockNum != uint16(blk) {
			return conn, flog.Raisef("wrong block ack, got %d, want %d", ack.BlockNum, blk)
		}
	}

	return conn, nil
}
