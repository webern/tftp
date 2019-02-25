package cor

import (
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
