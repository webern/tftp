package cor

import (
	"testing"
)

func TestErrCodeString(t *testing.T) {
	for i := -1; i < 10; i++ {
		ec := ErrCode(i)
		str := ec.String()

		if len(str) == 0 {
			t.Error("empty string")
		}
	}
}
