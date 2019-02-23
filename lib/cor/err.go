package cor

import (
	"fmt"
	"github.com/webern/flog"
	"net"
	"strconv"
)

type Err struct {
	packet   PacketError
	location string
}

func (e *Err) Error() string {
	return e.location + " - error - " + e.packet.Msg
}

func (e *Err) Send(conn *net.UDPConn) error {
	flog.NotImplemented()
	return flog.Raise("not implemented")
}

func NewErr(code ErrCode, message string) *Err {
	e := Err{}
	e.packet.Msg = "Err " + strconv.Itoa(int(code)) + " " + message
	e.packet.Code = code
	e.location = flog.Caller(2)
	return &e
}

func NewErrf(code ErrCode, messageFmt string, args ...interface{}) error {
	e := Err{}
	e.packet.Msg = "Err " + strconv.Itoa(int(code)) + " " + fmt.Sprintf(messageFmt, args...)
	e.packet.Code = code
	e.location = flog.Caller(2)
	return &e
}

func NewErrWrap(err error) *Err {
	return NewErr(ErrUnknown, err.Error())
}
