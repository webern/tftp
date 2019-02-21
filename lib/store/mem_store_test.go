// Copyright (c) 2019 by Matthew James Briggs, https://github.com/webern

package store

import (
	"github.com/webern/flog"
	"github.com/webern/tftp/lib/tfcore"
	"math"
	"testing"
)

func makeTestData(size int) []byte {
	b := make([]byte, size, size)
	for i := 0; i < size; i++ {
		b[i] = uint8(i % math.MaxUint8)
	}
	return b
}

func makeTestFile(name string, size int) tfcore.File {
	f := tfcore.File{}
	f.Name = name
	f.Data = makeTestData(size)
	return f
}

func TestMemStore(t *testing.T) {
	mstore := NewMemStore()
	defer mstore.Terminate()
	reqCh := mstore.RequestChan()
	respCh := make(chan Response)

	testFile1 := makeTestFile("testFile1", 1437)
	go func() { reqCh <- Request{Type: ReqPut, File: &testFile1, ResponseChan: respCh} }()
	resp := <-respCh

	testFile2 := makeTestFile("testFile1", 2)
	go func() { reqCh <- Request{Type: ReqPut, File: &testFile2, ResponseChan: respCh} }()
	resp = <-respCh

	flog.Info(resp)
}
