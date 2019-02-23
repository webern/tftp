// Copyright (c) 2019 by Matthew James Briggs, https://github.com/webern

package cor

import (
	"github.com/webern/tcore"
	"reflect"
	"testing"
)

func TestSerializationDeserialization(t *testing.T) {
	tests := []struct {
		bytes  []byte
		packet Packet
		op     uint16
	}{
		{
			[]byte("\x00\x01foo\x00bar\x00"),
			&PacketRequest{OpRRQ, "foo", "bar"},
			OpRRQ,
		},
		{
			[]byte("\x00\x02foo\x00bar\x00"),
			&PacketRequest{OpWRQ, "foo", "bar"},
			OpWRQ,
		},
		{
			[]byte("\x00\x03\x12\x34fnord"),
			&PacketData{0x1234, []byte("fnord")},
			OpData,
		},
		{
			[]byte("\x00\x03\x12\x34"),
			&PacketData{0x1234, []byte("")},
			OpData,
		},
		{
			[]byte("\x00\x04\xd0\x0f"),
			&PacketAck{0xd00f},
			OpAck,
		},
		{
			[]byte("\x00\x05\xab\xcdparachute failure\x00"),
			&PacketError{0xabcd, "parachute failure"},
			OpError,
		},
	}

	for _, test := range tests {
		actualBytes := test.packet.Serialize()
		if !reflect.DeepEqual(test.bytes, actualBytes) {
			t.Errorf("Serializing %#v: expected %q; got %q", test.packet, test.bytes, actualBytes)
		}

		actualPacket, err := ParsePacket(test.bytes)
		if err != nil {
			t.Errorf("Unable to parse packet %q: %s", test.bytes, err)
		} else if !reflect.DeepEqual(test.packet, actualPacket) {
			t.Errorf("Deserializing %q: expected %#v; got %#v", test.bytes, test.packet, actualPacket)
		}

		stm := "Packet.Op()"
		gotOP := actualPacket.Op()
		wantOP := test.op
		if msg, ok := tcore.TAssertInt(stm, int(gotOP), int(wantOP)); !ok {
			t.Error(msg)
		}

		stm = "Packet.IsRRQ()"
		gotB := actualPacket.IsRRQ()
		wantB := test.op == OpRRQ
		if msg, ok := tcore.TAssertBool(stm, gotB, wantB); !ok {
			t.Error(msg)
		}

		stm = "Packet.IsWRQ()"
		gotB = actualPacket.IsWRQ()
		wantB = test.op == OpWRQ
		if msg, ok := tcore.TAssertBool(stm, gotB, wantB); !ok {
			t.Error(msg)
		}

		stm = "Packet.IsData()"
		gotB = actualPacket.IsData()
		wantB = test.op == OpData
		if msg, ok := tcore.TAssertBool(stm, gotB, wantB); !ok {
			t.Error(msg)
		}

		stm = "Packet.IsAck()"
		gotB = actualPacket.IsAck()
		wantB = test.op == OpAck
		if msg, ok := tcore.TAssertBool(stm, gotB, wantB); !ok {
			t.Error(msg)
		}

		stm = "Packet.IsError()"
		gotB = actualPacket.IsError()
		wantB = test.op == OpError
		if msg, ok := tcore.TAssertBool(stm, gotB, wantB); !ok {
			t.Error(msg)
		}
	}
}

func TestDeserializationInvalid(t *testing.T) {
	tests := [][]byte{
		// no opcode
		[]byte(""),

		// invalid opcode
		[]byte("\x00\x00"),
		[]byte("\x00\x06"),
		[]byte("\xff\x01"),
		[]byte("\xff\xff"),

		// short RRQ
		[]byte("\x00\x01"),
		[]byte("\x00\x01foo"),
		[]byte("\x00\x01foo\x00"),
		[]byte("\x00\x01foo\x00bar"),

		// short WRQ
		[]byte("\x00\x02"),
		[]byte("\x00\x02foo"),
		[]byte("\x00\x02foo\x00"),
		[]byte("\x00\x02foo\x00bar"),

		// short data
		[]byte("\x00\x03"),
		[]byte("\x00\x03\x01"),

		// short ack
		[]byte("\x00\x04"),
		[]byte("\x00\x04\x01"),

		// short error
		[]byte("\x00\x05"),
		[]byte("\x00\x05\xab"),
		[]byte("\x00\x05\xab\xcd"),
		[]byte("\x00\x05\xab\xcdparachute failure"),

		// truncated error
		[]byte("\x05"),
	}

	for _, test := range tests {
		if p, err := ParsePacket(test); err == nil {
			t.Errorf("Parsing packet %q: expected error; got %#v", test, p)
		}
	}
}

func TestParseError(t *testing.T) {
	tests := [][]byte{
		// no opcode
		[]byte(""),

		// invalid opcode
		[]byte("\x00\x00"),
		[]byte("\x00\x06"),
		[]byte("\xff\x01"),
		[]byte("\xff\xff"),

		// no zero terminator for mode
		[]byte("\x00\x01foo\x00bar\x01"),
	}

	for _, test := range tests {
		p := PacketRequest{}
		if err := p.Parse(test); err == nil {
			t.Errorf("Parsing packet %q: expected error; got %#v", test, p)
		}
	}
}
