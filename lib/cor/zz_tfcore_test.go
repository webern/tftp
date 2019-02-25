// Copyright (c) 2019 by Matthew James Briggs, https://github.com/webern

package cor

import (
	"os"
	"testing"

	"github.com/webern/flog"
)

// TestMain runs once and it calls all tests with m.Run()
// This is here to set the logger.
func TestMain(m *testing.M) {
	flog.SetTruncationPath("/tftp")
	flog.SetLevel(flog.ErrorLevel)
	code := m.Run()
	os.Exit(code)
}
