// Copyright (c) 2019 by Matthew James Briggs, https://github.com/webern

package tfsrv

import (
	"fmt"
	"github.com/webern/flog"
	"github.com/webern/tcore"
	"github.com/webern/tftp/lib/tfcore"
	"math"
	"net"
	"sync"
	"testing"
	"time"
)

var client, _ = net.ResolveUDPAddr("udp", ":12345")
var server, _ = net.ResolveUDPAddr("udp", ":54321")

func TestSendHandshakeAck(t *testing.T) {
	receiver, err := net.DialUDP("udp", client, server)
	sender, err := net.DialUDP("udp", server, client)

	if err != nil {
		t.Error(err.Error())
	}

	buf := make([]byte, tfcore.MaxPacketSize)
	readyForSend := sync.WaitGroup{}
	readyForSend.Add(1)
	doneReceiving := sync.WaitGroup{}
	doneReceiving.Add(1)

	go func() {
		go func() {
			defer doneReceiving.Done()
			numBytes, addr, err := receiver.ReadFromUDP(buf)

			if err != nil {
				t.Error(err.Error())
			}

			if addr == nil {
				t.Error("addr is nil")
			}

			if numBytes <= 0 {
				t.Error("num bytes should be greater than zero")
			}
		}()
		readyForSend.Done()
	}()

	readyForSend.Wait()
	err = sendHandshakeAck(sender)
	doneReceiving.Wait()
	packet, err := tfcore.ParsePacket(buf)

	if err != nil {
		t.Errorf("a malformed packet was sent: %s", err.Error())
	}

	if packet == nil {
		t.Errorf("a nil packet was received")
		return
	}

	if packet.Op() != tfcore.OpAck {
		t.Errorf("the wrong op type was received: want %d, got %d", tfcore.OpAck, packet.Op())
	}

	ack, ok := packet.(*tfcore.PacketAck)

	if !ok || ack == nil {
		t.Error("the ack packet could not be downcast to the correct type")
		return
	}

	if ack.BlockNum != 0 {
		t.Errorf("the ack packet has the wrong block num: want %d, got %d", 0, ack.BlockNum)
	}
}

func makeTestData(size int) []byte {
	b := make([]byte, size, size)
	for i := 0; i < size; i++ {
		b[i] = uint8(i % math.MaxUint8)
	}
	return b
}

func TestPut(t *testing.T) {
	filename := "testfile.zip"
	testFile := makeTestData(3671)
	clientConn, err := net.DialUDP("udp", client, server)

	if err != nil {
		t.Error(err.Error())
	}

	//buf := make([]byte, tfcore.MaxPacketSize)
	readyForSend := sync.WaitGroup{}
	readyForSend.Add(1)
	doneReceiving := sync.WaitGroup{}
	doneReceiving.Add(1)
	ch := make(chan tfcore.File, 2)
	ready := make(chan struct{})

	go func() {
		defer readyForSend.Done()
		h := handshake{}
		h.client = *client
		h.server = *server
		h.tftpInfo.OpCode = tfcore.OpWRQ
		h.tftpInfo.Filename = filename
		err := put(h, ch, ready)
		if err != nil {
			flog.Error(err.Error())
		}
	}()

	//readyForSend.Wait()

	<-ready
	//time.Sleep(500 * time.Millisecond)

	go func() {
		sendEmptyAtEnd := false
		blk := 1
		for pos := 0; pos < len(testFile); {
			data := tfcore.PacketData{}
			data.BlockNum = uint16(blk)
			end := pos + tfcore.BlockSize
			if end > len(testFile) {
				end = len(testFile)
				sendEmptyAtEnd = false
			} else {
				sendEmptyAtEnd = true
			}
			data.Data = testFile[pos:end]
			_, err := clientConn.Write(data.Serialize())
			time.Sleep(50 * time.Millisecond)

			if err != nil {
				flog.Error(err.Error())
				//os.Exit(1)
			}

			//n, y, err := clientConn.ReadFromUDP(make([]byte, 2048)) // ignore acklowledgement
			//
			//if err != nil {
			//	flog.Error(err.Error())
			//	os.Exit(1)
			//} else if n <= 0 {
			//	flog.Error("no bytes")
			//	os.Exit(1)
			//}
			//
			//flog.Info(y)
			blk++
			pos = end
		}

		if sendEmptyAtEnd {
			data := tfcore.PacketData{}
			data.BlockNum = uint16(blk)
			data.Data = make([]byte, 0)
			clientConn.Write(data.Serialize())
			//clientConn.ReadFromUDP(make([]byte, 100)) // ignore acklowledgement
		}
	}()

	file := <-ch

	stm := "file.Name"
	gotS := file.Name
	wantS := filename
	if msg, ok := tcore.TAssertString(stm, gotS, wantS); !ok {
		t.Error(msg)
	}

	stm = "len(file.Data)"
	gotI := len(file.Data)
	wantI := len(testFile)
	if msg, ok := tcore.TAssertInt(stm, gotI, wantI); !ok {
		t.Error(msg)
	}

	l := len(file.Data)
	if len(testFile) < l {
		l = len(testFile)
	}

	for i := 0; i < l; i++ {
		stm = fmt.Sprintf("int(file.Data[%d])", i)
		gotI = int(file.Data[i])
		wantI = int(testFile[i])
		if msg, ok := tcore.TAssertInt(stm, gotI, wantI); !ok {
			t.Error(msg)
		}
	}
}
