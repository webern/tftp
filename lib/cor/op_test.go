package cor

import (
	"fmt"
	"testing"

	"github.com/webern/tcore"
)

func TestOpString(t *testing.T) {
	for i := -1; i < 10; i++ {
		op := OpType(i)
		str := op.String()
		got := len(str)
		want := 4
		if msg, ok := tcore.TAssertInt(fmt.Sprintf("len(\"%s\")", str), got, want); !ok {
			t.Error(msg)
		}
	}
}
