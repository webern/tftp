package cor

import (
	"net"
	"testing"

	"github.com/webern/flog"
	"github.com/webern/tcore"
)

func TestNewErr(t *testing.T) {
	err := NewErr(ErrNotFound, "hi")

	if msg, ok := tcore.TAssertInt("", int(err.Code()), int(ErrNotFound)); !ok {
		t.Error(msg)
	}
}

func TestNewErrf(t *testing.T) {
	err := NewErrf(ErrBadID, "hi-%d", 50)

	if msg, ok := tcore.TAssertInt("", int(err.Code()), int(ErrBadID)); !ok {
		t.Error(msg)
	}
}

func TestNewErrWrap(t *testing.T) {
	err := NewErrWrap(flog.Raise("hi"))

	if msg, ok := tcore.TAssertInt("", int(err.Code()), int(ErrUnknown)); !ok {
		t.Error(msg)
	}

	str := err.Error()

	if len(str) == 0 {
		t.Error("str := err.Error() produced an empty string")
	}
}

func TestErrSend(t *testing.T) {
	conn, err := setupConn()

	if err != nil {
		t.Error(err)
		return
	}

	theError := NewErr(ErrAccess, "--")
	err = theError.Send(conn)

	if err != nil {
		t.Error(err)
		return
	}
}

func setupConn() (*net.UDPConn, error) {
	addr1, err := net.ResolveUDPAddr("udp", ":3333")

	if err != nil {
		return nil, err
	}

	addr2, err := net.ResolveUDPAddr("udp", ":4444")

	if err != nil {
		return nil, err
	}

	conn, err := net.DialUDP("udp", addr1, addr2)

	if err != nil {
		return nil, err
	}

	return conn, nil
}
