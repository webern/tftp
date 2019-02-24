// Copyright (c) 2019 by Matthew James Briggs, https://github.com/webern

package cor

import (
	"fmt"
	"net"

	"github.com/webern/flog"
)

// Err represents a PacketErr while also implementing the error interface
type Err struct {
	packet   PacketError
	location string
}

// Error implements the error interface
func (e *Err) Error() string {
	return fmt.Sprintf("%s: %s, location: %s", e.packet.Code.String(), e.packet.Msg, e.location)
}

// Send sends the Err as a PacketError to conn
func (e *Err) Send(conn *net.UDPConn) error {
	_, err := conn.Write(e.packet.Serialize())

	if err != nil {
		return err
	}

	return nil
}

// Code gets the error Code
func (e *Err) Code() ErrCode {
	return e.packet.Code
}

// NewErr creates a new Err
func NewErr(code ErrCode, message string) *Err {
	e := Err{}
	e.packet.Msg = message
	e.packet.Code = code
	e.location = flog.Caller(2)
	return &e
}

// NewErrf creates a new error using fmt.Printf semantics
func NewErrf(code ErrCode, messageFmt string, args ...interface{}) error {
	e := Err{}
	e.packet.Msg = fmt.Sprintf(messageFmt, args...)
	e.packet.Code = code
	e.location = flog.Caller(2)
	return &e
}

// NewErrWrap takes an error and creates an Err out of it
func NewErrWrap(err error) *Err {
	e := Err{}
	e.packet.Msg = err.Error()
	e.packet.Code = ErrUnknown
	e.location = flog.Caller(2)
	return &e
}
