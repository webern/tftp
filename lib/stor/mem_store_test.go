// Copyright (c) 2019 by Matthew James Briggs, https://github.com/webern

package stor

import (
	"fmt"
	"github.com/webern/tcore"
	"github.com/webern/tftp/lib/cor"
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

func makeTestFile(name string, size int) cor.File {
	f := cor.File{}
	f.Name = name
	f.Data = makeTestData(size)
	return f
}

func TestMemStore(t *testing.T) {
	mstore := NewMemStore()
	defer mstore.Terminate()

	fname := "anyfile.txt"

	f := makeTestFile(fname, 100)
	err := mstore.Put(f)

	if msg, ok := tcore.TErr("mstore.Put(f)", err); !ok {
		t.Error(msg)
	}

	got, err := mstore.Get(fname)
	if msg, ok := tcore.TErr("got, err := mstore.Get(fname)", err); !ok {
		t.Error(msg)
	}

	stm := "len(got.Data)"
	gotI := len(got.Data)
	wantI := len(f.Data)
	if msg, ok := tcore.TAssertInt(stm, gotI, wantI); !ok {
		t.Error(msg)
		return // avoid a panic in the next loop
	}

	for i := 0; i < len(got.Data); i++ {
		stm = fmt.Sprintf("got.Data[%d]", i)
		gotI = int(got.Data[i])
		wantI = int(f.Data[i])
		if msg, ok := tcore.TAssertInt(stm, gotI, wantI); !ok {
			t.Error(msg)
		}
	}

	// prove that these are not pointing to the same byte array by changing one of them

	// first prove that both have '0' at index 0
	stm = "f.Data[0]"
	gotI = int(f.Data[0])
	wantI = 0
	if msg, ok := tcore.TAssertInt(stm, gotI, wantI); !ok {
		t.Error(msg)
	}

	stm = "got.Data[0]"
	gotI = int(got.Data[0])
	wantI = 0
	if msg, ok := tcore.TAssertInt(stm, gotI, wantI); !ok {
		t.Error(msg)
	}

	// change one of these
	got.Data[0] = 12

	// assert that only 'got' was changed, e.i. that 'f' did not change
	stm = "f.Data[0]"
	gotI = int(f.Data[0])
	wantI = 0
	if msg, ok := tcore.TAssertInt(stm, gotI, wantI); !ok {
		t.Error(msg)
	}

	stm = "got.Data[0]"
	gotI = int(got.Data[0])
	wantI = 12
	if msg, ok := tcore.TAssertInt(stm, gotI, wantI); !ok {
		t.Error(msg)
	}

	// prove that requesting a non-existent file returns an error
	_, err = mstore.Get("nope")

	if err == nil {
		t.Errorf("'%s' was expected to throw an error, but did not", "_, err = mstore.Get(\"nope\")")
	}
}
