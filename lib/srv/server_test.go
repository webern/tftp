// Copyright (c) 2019 by Matthew James Briggs, https://github.com/webern

package srv

import (
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/webern/flog"
	"github.com/webern/tcore"
	"github.com/webern/tftp/lib/cor"
	"github.com/webern/tftp/lib/stor"
)

func TestServer(t *testing.T) {
	listeningPort := 11111
	store := stor.NewMemStore()
	server := NewServer(store)
	server.Port = listeningPort
	server.Verbose = true
	srvErrChan := make(chan error, 1)

	go func() {
		err := server.Serve()
		if err != nil {
			flog.Error(err.Error())
			srvErrChan <- err
		}
		close(srvErrChan)
	}()

	time.Sleep(50 * time.Millisecond)

	// TODO - I can't figure out how to actually test transfers to and from the server here
	// doServerTransfers(listeningPort, 12345, 1000)

	err := server.Stop()

	if msg, ok := tcore.TErr("err := server.Stop()", err); !ok {
		t.Error(msg)
	}

	for {
		err, ok := <-srvErrChan

		if !ok {
			break
		}

		if msg, ok := tcore.TErr("err := server.Serve()", err); !ok {
			t.Error(msg)
		}
	}
}

// TODO - I can't figure out how to write to a conn then read from it for the acknowledgement
func doServerTransfers(handShakePort int, returnPort int, fileSize int) {

	mainAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", handShakePort))

	if err != nil {
		panic(err)
	}

	myAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", returnPort))

	if err != nil {
		panic(err)
	}

	handshakeConn, err := net.DialUDP("udp", myAddr, mainAddr)

	if err != nil {
		flog.Error(err.Error())
		panic(err)
	}

	u, err := uuid.NewV4()
	if err != nil {
		panic(err)
	}
	filename := u.String()

	wpacket := cor.PacketRequest{}
	wpacket.OpCode = cor.OpWRQ
	wpacket.Filename = filename
	wpacket.Mode = "anything"

	buf := make([]byte, TftpMaxPacketSize)
	_, err = handshakeConn.Write(wpacket.Serialize())

	if err != nil {
		panic(err)
	}

	time.Sleep(1000 * time.Millisecond)
	n, sendAddr, err := handshakeConn.ReadFromUDP(buf)

	if err != nil {
		panic(err)
	} else if n <= 0 {
		panic(0)
	}

	err = handshakeConn.Close()

	if err != nil {
		panic(err)
	}
	handshakeConn = nil
	transferConn, err := net.DialUDP("udp", sendAddr, myAddr)
	data := makeTestData(fileSize)

	sendEmptyAtEnd := false
	blk := 1
	for pos := 0; pos < len(data); {
		packet := cor.PacketData{}
		packet.BlockNum = uint16(blk)
		end := pos + cor.BlockSize
		if end > len(data) {
			end = len(data)
			sendEmptyAtEnd = false
		} else {
			sendEmptyAtEnd = true
		}
		packet.Data = data[pos:end]
		_, err := transferConn.Write(packet.Serialize())

		if err != nil {
			panic(err)
		}

		blk++
		pos = end
	}

	if sendEmptyAtEnd {
		data := cor.PacketData{}
		data.BlockNum = uint16(blk)
		data.Data = make([]byte, 0)
		_, err = transferConn.Write(data.Serialize())

		if err != nil {
			flog.Error(err.Error())
		}
	}
}
