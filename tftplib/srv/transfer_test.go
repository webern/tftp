package tftpsrv

import (
	"github.com/webern/flog"
	"github.com/webern/tftp/tftplib/wire"
	"net"
	"sync"
	"testing"
)

var client, _ = net.ResolveUDPAddr("udp", ":12345")
var server, _ = net.ResolveUDPAddr("udp", ":54321")

func TestSendHandshakeAck(t *testing.T) {
	receiver, err := net.DialUDP("udp", client, server)
	sender, err := net.DialUDP("udp", server, client)

	if err != nil {
		t.Error(err.Error())
	}

	buf := make([]byte, wire.MaxPacketSize)
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

	flog.Info(string(buf))
	packet, err := wire.ParsePacket(buf)

	if err != nil {
		t.Errorf("a malformed packet was sent: %s", err.Error())
	}

	if packet == nil {
		t.Errorf("a nil packet was received")
		return
	}

	if packet.Op() != wire.OpAck {
		t.Errorf("the wrong op type was received: want %d, got %d", wire.OpAck, packet.Op())
	}

	ack, ok := packet.(*wire.PacketAck)

	if !ok {
		t.Error("the ack packet could not be downcast to the correct type")
	}

	if ack.BlockNum != 0 {
		t.Errorf("the ack packet has the wrong block num: want %d, got %d")
	}

	flog.Info(packet)
}
