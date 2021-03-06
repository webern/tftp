// Copyright (c) 2019 by Matthew James Briggs, https://github.com/webern

package srv

import (
	"fmt"
	"math"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/webern/tftp/lib/stor"

	"github.com/webern/flog"
	"github.com/webern/tcore"
	"github.com/webern/tftp/lib/cor"
)

func TestSendHandshakeAck(t *testing.T) {
	receiver, sender, err := setupConn()

	if msg, ok := tcore.TErr("receiver, sender, err := setupConn()", err); !ok {
		t.Error(msg)
		return
	}

	buf := make([]byte, cor.MaxPacketSize)
	doneReceiving := sync.WaitGroup{}
	doneReceiving.Add(1)

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

	err = sendHandshakeAck(sender)

	if msg, ok := tcore.TErr("err = sendHandshakeAck(sender)", err); !ok {
		t.Error(msg)
	}

	doneReceiving.Wait()
	checkAckPacket(t, buf)
}

func checkAckPacket(t *testing.T, buf []byte) {
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

//func setupAddr( ) {
//
//}

func setupConn() (receiver, sender *net.UDPConn, err error) {
	client, err := net.ResolveUDPAddr("udp", ":54236")

	if err != nil {
		return nil, nil, err
	}

	server, err := net.ResolveUDPAddr("udp", ":32451")

	if err != nil {
		return nil, nil, err
	}

	receiver, err = net.DialUDP("udp", client, server)

	if err != nil {
		return nil, nil, err
	}

	sender, err = net.DialUDP("udp", server, client)

	if err != nil {
		return nil, nil, err
	}

	return receiver, sender, nil
}

func TestPut(t *testing.T) {
	var client, _ = net.ResolveUDPAddr("udp", ":12985")
	var server, _ = net.ResolveUDPAddr("udp", ":21985")
	filename := "testfile.zip"
	testFile := makeTestData(3671)
	clientConn, err := net.DialUDP("udp", client, server)

	if err != nil {
		t.Error(err.Error())
	}

	memStore := stor.NewMemStore()

	wg := callPutFunctionOnServerAsync(client, server, filename, memStore)

	time.Sleep(50 * time.Millisecond)

	err = sendData(testFile, clientConn, err)

	if msg, ok := tcore.TErr("err = sendData(testFile, clientConn, err)", err); !ok {
		t.Error(msg)
	}

	wg.Wait()
	time.Sleep(100 * time.Millisecond)
	doPutTestAssertions(t, err, memStore, filename, testFile)
}

func callPutFunctionOnServerAsync(client *net.UDPAddr, server *net.UDPAddr, filename string, memStore stor.Store) *sync.WaitGroup {
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go callPutFunctionOnServer(client, server, filename, memStore, wg)
	return wg
}

func callPutFunctionOnServer(client *net.UDPAddr, server *net.UDPAddr, filename string, memStore stor.Store, wg *sync.WaitGroup) {
	h := handshake{}
	h.client = *client
	h.server = *server
	h.tftpInfo.OpCode = cor.OpWRQ
	h.tftpInfo.Filename = filename

	_, _, err := put(h, memStore)

	if err != nil {
		flog.Error(err.Error())
	}

	wg.Done()
}

func sendData(testFile []byte, clientConn *net.UDPConn, err error) error {
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

		if err != nil {
			return err
		}

		blk++
		pos = end
	}

	if sendEmptyAtEnd {
		data := cor.PacketData{}
		data.BlockNum = uint16(blk)
		data.Data = make([]byte, 0)
		_, err = clientConn.Write(data.Serialize())

		if err != nil {
			return err
		}
	}

	return nil
}

func doPutTestAssertions(t *testing.T, err error, memStore stor.Store, filename string, testFile []byte) {
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
		//return // avoid panic in the comparison loop
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
			//break
		}
	}
}
