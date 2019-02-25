package main

import (
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/webern/tcore"
)

func TestRun(t *testing.T) {

	initialArgs := os.Args

	os.Args = []string{"program-name", "--port=47381", "--quiet"}

	sigChan := make(chan os.Signal, 1)
	go func() {
		time.Sleep(500 * time.Millisecond)
		sigChan <- syscall.SIGINT
	}()

	err := run(sigChan)

	if msg, ok := tcore.TErr("err := run(sigChan)", err); !ok {
		t.Error(msg)
	}

	os.Args = initialArgs
}
