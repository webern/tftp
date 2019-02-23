// Copyright (c) 2019 by Matthew James Briggs, https://github.com/webern

package srv

import (
	"fmt"
	"github.com/webern/tftp/lib/stor"
	"math"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/webern/flog"
	"github.com/webern/tcore"
	"github.com/webern/tftp/lib/cor"
)

var client, _ = net.ResolveUDPAddr("udp", ":12345")
var server, _ = net.ResolveUDPAddr("udp", ":54321")

func TestSendHandshakeAck(t *testing.T) {
	receiver, err := net.DialUDP("udp", client, server)
	sender, err := net.DialUDP("udp", server, client)

	if err != nil {
		t.Error(err.Error())
	}

	buf := make([]byte, cor.MaxPacketSize)
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
	packet, err := cor.ParsePacket(buf)

	if err != nil {
		t.Errorf("a malformed packet was sent: %s", err.Error())
	}

	if packet == nil {
		t.Errorf("a nil packet was received")
		return
	}

	if packet.Op() != cor.OpAck {
		t.Errorf("the wrong op type was received: want %d, got %d", cor.OpAck, packet.Op())
	}

	ack, ok := packet.(*cor.PacketAck)

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
	ready := make(chan struct{})
	memStore := stor.NewMemStore()

	go func() {
		defer readyForSend.Done()
		h := handshake{}
		h.client = *client
		h.server = *server
		h.tftpInfo.OpCode = cor.OpWRQ
		h.tftpInfo.Filename = filename
		err := put(h, memStore, ready)
		if err != nil {
			flog.Error(err.Error())
		}
	}()

	//readyForSend.Wait()

	<-ready
	//time.Sleep(50 * time.Millisecond)

	go func() {
		sendEmptyAtEnd := false
		blk := 1
		for pos := 0; pos < len(testFile); {
			data := cor.PacketData{}
			data.BlockNum = uint16(blk)
			end := pos + cor.BlockSize
			if end > len(testFile) {
				end = len(testFile)
				sendEmptyAtEnd = false
			} else {
				sendEmptyAtEnd = true
			}
			data.Data = testFile[pos:end]
			_, err := clientConn.Write(data.Serialize())
			//time.Sleep(50 * time.Millisecond)

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
			data := cor.PacketData{}
			data.BlockNum = uint16(blk)
			data.Data = make([]byte, 0)
			_, _ = clientConn.Write(data.Serialize())

			//if msg, ok := tcore.TErr("_, err = clientConn.Write(data.Serialize())", err); !ok {
			//	t.Error(msg)
			//}
		}
	}()

	// TODO - this is lame, how can data race be properly avoided in this test?
	time.Sleep(500 * time.Millisecond)
	gotFile, err := memStore.Get(filename)

	if msg, ok := tcore.TErr("gotFile, err := memStore.Get(filename)", err); !ok {
		t.Error(msg)
	}

	stm := "gotFile.Name"
	gotS := gotFile.Name
	wantS := filename
	if msg, ok := tcore.TAssertString(stm, gotS, wantS); !ok {
		t.Error(msg)
	}

	stm = "len(gotFile.Data)"
	gotI := len(gotFile.Data)
	wantI := len(testFile)
	if msg, ok := tcore.TAssertInt(stm, gotI, wantI); !ok {
		t.Error(msg)
		return // avoid panic in the comparison loop
	}

	errCount := 0
	for i := 0; i < len(gotFile.Data); i++ {
		stm = fmt.Sprintf("int(gotFile.Data[%d])", i)
		gotI = int(gotFile.Data[i])
		wantI = int(testFile[i])
		if msg, ok := tcore.TAssertInt(stm, gotI, wantI); !ok {
			t.Error(msg)
			errCount++
		}

		if errCount > 10 {
			t.Error("more than 10 errors, stopping this loop")
			break
		}
	}
}
