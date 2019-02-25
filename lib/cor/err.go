// Copyright (c) 2019 by Matthew James Briggs, https://github.com/webern

package cor

import (
	"fmt"
	"net"

	"github.com/webern/flog"
)

type Err struct {
	packet   PacketError
	location string
}

func (e *Err) Error() string {
	return fmt.Sprintf("%s: %s, location: %s", e.packet.Code.String(), e.packet.Msg, e.location)
}

func (e *Err) Send(conn *net.UDPConn) error {
	_, err := conn.Write(e.packet.Serialize())

	if err != nil {
		return err
	}

	return nil
}

func (e *Err) Code() ErrCode {
	return e.packet.Code
}

func NewErr(code ErrCode, message string) *Err {
	e := Err{}
	e.packet.Msg = message
	e.packet.Code = code
	e.location = flog.Caller(2)
	return &e
}

func NewErrf(code ErrCode, messageFmt string, args ...interface{}) *Err {
	e := Err{}
	e.packet.Msg = fmt.Sprintf(messageFmt, args...)
	e.packet.Code = code
	e.location = flog.Caller(2)
	return &e
}

func NewErrWrap(err error) *Err {
	e := Err{}
	e.packet.Msg = err.Error()
	e.packet.Code = ErrUnknown
	e.location = flog.Caller(2)
	return &e
}
