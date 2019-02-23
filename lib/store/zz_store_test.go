package store

import (
	"github.com/webern/flog"
	"os"
	"testing"
)

// TestMain runs once and it calls all tests with m.Run()
// This is here to set the logger.
func TestMain(m *testing.M) {
	flog.SetTruncationPath("/tftp")
	flog.SetLevel(flog.TraceLevel)
	code := m.Run()
	os.Exit(code)
}
