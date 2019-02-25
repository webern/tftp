// Copyright (c) 2019 by Matthew James Briggs, https://github.com/webern

package cor

import (
	"bytes"
	"encoding/binary"

	"github.com/webern/flog"
)

// MaxPacketSize represents the maximum expected UDP packet size. Larger than a typical mtu (1500), and largest DATA
// packet (516). may limit the length of filenames in RRQ/WRQs -- RFC1350 doesn't offer a bound for these.
const MaxPacketSize = 2048

// Packet is the interface met by all packet structs
type Packet interface {
	// Op returns the OpType code for this packet
	Op() OpType

	// Parse parses a packet from its wire representation
	Parse([]byte) error

	// Serialize serializes a packet to its wire representation
	Serialize() []byte

	// IsRRQ is for convenience, returns true if the packet is a read request packet
	IsRRQ() bool

	// IsWRQ is for convenience, returns true if the packet is a write request packet
	IsWRQ() bool

	// IsData is for convenience, returns true if the packet is a data packet
	IsData() bool

	// IsAck is for convenience, returns true if the packet is a ack packet
	IsAck() bool

	// IsError is for convenience, returns true if the packet is an error packet
	IsError() bool
}

// PacketRequest represents a request to read or rite a file.
type PacketRequest struct {
	OpCode   OpType // OpRRQ or OpWRQ
	Filename string
	Mode     string
}

// Op returns the OpType code for this packet
func (p *PacketRequest) Op() OpType {
	return p.OpCode
}

// Parse parses a packet
func (p *PacketRequest) Parse(buf []byte) (err error) {
	var opUntyped uint16
	var buf2 []byte
	var buf3 []byte

	if opUntyped, buf2, err = parseUint16(buf); err != nil {
		return err
	}

	p.OpCode = OpType(opUntyped)

	if p.Filename, buf3, err = parseString(buf2); err != nil {
		return err
	}

	if p.Mode, _, err = parseString(buf3); err != nil {
		return err
	}

	return nil
}

// Serialize serializes a packet to its wire representation
func (p *PacketRequest) Serialize() []byte {
	buf := make([]byte, 2+len(p.Filename)+1+len(p.Mode)+1)
	binary.BigEndian.PutUint16(buf, uint16(p.OpCode))
	copy(buf[2:], p.Filename)
	copy(buf[2+len(p.Filename)+1:], p.Mode)
	return buf
}

// IsRRQ is for convenience, returns true if the packet is a read request packet
func (p *PacketRequest) IsRRQ() bool {
	return p.OpCode == OpRRQ
}

// IsWRQ is for convenience, returns true if the packet is a write request packet
func (p *PacketRequest) IsWRQ() bool {
	return p.OpCode == OpWRQ
}

// IsData is for convenience, returns true if the packet is a data packet
func (p *PacketRequest) IsData() bool {
	return false
}

// IsAck is for convenience, returns true if the packet is a ack packet
func (p *PacketRequest) IsAck() bool {
	return false
}

// IsError is for convenience, returns true if the packet is an error packet
func (p *PacketRequest) IsError() bool {
	return false
}

// PacketData carries a block of data in a file transmission.
type PacketData struct {
	BlockNum uint16
	Data     []byte
}

// Op returns the OpType code for this packet
func (p *PacketData) Op() OpType {
	return OpData
}

// Parse parses a packet
func (p *PacketData) Parse(buf []byte) (err error) {
	buf = buf[2:] // skip over op
	if p.BlockNum, buf, err = parseUint16(buf); err != nil {
		return err
	}
	p.Data = buf
	return nil
}

// Serialize serializes a packet to its wire representation
func (p *PacketData) Serialize() []byte {
	buf := make([]byte, 4+len(p.Data))
	binary.BigEndian.PutUint16(buf, OpData)
	binary.BigEndian.PutUint16(buf[2:], p.BlockNum)
	copy(buf[4:], p.Data)
	return buf
}

// IsRRQ is for convenience, returns true if the packet is a read request packet
func (p *PacketData) IsRRQ() bool {
	return false
}

// IsWRQ is for convenience, returns true if the packet is a write request packet
func (p *PacketData) IsWRQ() bool {
	return false
}

// IsData is for convenience, returns true if the packet is a data packet
func (p *PacketData) IsData() bool {
	return true
}

// IsAck is for convenience, returns true if the packet is a ack packet
func (p *PacketData) IsAck() bool {
	return false
}

// IsError is for convenience, returns true if the packet is an error packet
func (p *PacketData) IsError() bool {
	return false
}

// PacketAck acknowledges receipt of a data packet
type PacketAck struct {
	BlockNum uint16
}

// Op returns the OpType code for this packet
func (p *PacketAck) Op() OpType {
	return OpAck
}

// Parse parses a packet
func (p *PacketAck) Parse(buf []byte) (err error) {
	buf = buf[2:] // skip over op
	if p.BlockNum, buf, err = parseUint16(buf); err != nil {
		return err
	}
	return nil
}

// Serialize serializes a packet to its wire representation
func (p *PacketAck) Serialize() []byte {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint16(buf, OpAck)
	binary.BigEndian.PutUint16(buf[2:], p.BlockNum)
	return buf
}

// IsRRQ is for convenience, returns true if the packet is a read request packet
func (p *PacketAck) IsRRQ() bool {
	return false
}

// IsWRQ is for convenience, returns true if the packet is a write request packet
func (p *PacketAck) IsWRQ() bool {
	return false
}

// IsData is for convenience, returns true if the packet is a data packet
func (p *PacketAck) IsData() bool {
	return false
}

// IsAck is for convenience, returns true if the packet is a ack packet
func (p *PacketAck) IsAck() bool {
	return true
}

// IsError is for convenience, returns true if the packet is an error packet
func (p *PacketAck) IsError() bool {
	return false
}

// PacketError is sent by a peer who has encountered an error condition
type PacketError struct {
	Code ErrCode
	Msg  string
}

// Op returns the OpType code for this packet
func (p *PacketError) Op() OpType {
	return OpError
}

// Parse parses a packet
func (p *PacketError) Parse(buf []byte) (err error) {
	buf = buf[2:] // skip over op
	code := uint16(0)
	if code, buf, err = parseUint16(buf); err != nil {
		return err
	}
	p.Code = ErrCode(code)
	if p.Msg, buf, err = parseString(buf); err != nil {
		return err
	}
	return nil
}

// Serialize serializes a packet to its wire representation
func (p *PacketError) Serialize() []byte {
	buf := make([]byte, 4+len(p.Msg)+1)
	binary.BigEndian.PutUint16(buf, OpError)
	binary.BigEndian.PutUint16(buf[2:], uint16(p.Code))
	copy(buf[4:], p.Msg)
	return buf
}

// IsRRQ is for convenience, returns true if the packet is a read request packet
func (p *PacketError) IsRRQ() bool {
	return false
}

// IsWRQ is for convenience, returns true if the packet is a write request packet
func (p *PacketError) IsWRQ() bool {
	return false
}

// IsData is for convenience, returns true if the packet is a data packet
func (p *PacketError) IsData() bool {
	return false
}

// IsAck is for convenience, returns true if the packet is a ack packet
func (p *PacketError) IsAck() bool {
	return false
}

// IsError is for convenience, returns true if the packet is an error packet
func (p *PacketError) IsError() bool {
	return true
}

// parseUint16 reads a big-endian uint16 from the beginning of buf,
// returning it along with a slice pointing at the next position in the buffer.
func parseUint16(buf []byte) (uint16, []byte, error) {
	if len(buf) < 2 {
		return 0, nil, flog.Raise("packet truncated")
	}
	return binary.BigEndian.Uint16(buf), buf[2:], nil
}

// parseString reads a null-terminated ASCII string from buf,
// returning it along with a slice pointing at the next position in the buffer.
func parseString(buf []byte) (string, []byte, error) {
	i := bytes.IndexByte(buf, 0)
	if i < 0 {
		return "", nil, flog.Raise("packet truncated")
	}
	return string(buf[:i]), buf[i+1:], nil
}

// ParsePacket parses a packet from its wire representation.
func ParsePacket(buf []byte) (p Packet, err error) {
	var opcode uint16
	if opcode, _, err = parseUint16(buf); err != nil {
		return
	}
	switch OpType(opcode) {
	case OpRRQ, OpWRQ:
		p = &PacketRequest{}
	case OpData:
		p = &PacketData{}
	case OpAck:
		p = &PacketAck{}
	case OpError:
		p = &PacketError{}
	default:
		err = flog.Raisef("unexpected opcode %d", opcode)
		return
	}
	err = p.Parse(buf)
	return
}
