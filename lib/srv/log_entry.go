// Copyright (c) 2019 by Matthew James Briggs, https://github.com/webern

package srv

import (
	"fmt"
	"net"
	"time"

	"github.com/webern/tftp/lib/cor"
)

// LogEntry represents an item that will be written to the connection log.
// Each client connection is represented by one LogEntry
type LogEntry struct {
	Start    time.Time
	Duration time.Duration
	Op       cor.OpType
	Client   net.UDPAddr
	Error    *cor.Err
	File     string
	Bytes    int
}

func (l *LogEntry) String() string {
	opName := "UNK"

	if l.Op == cor.OpRRQ {
		opName = "GET"
	} else if l.Op == cor.OpWRQ {
		opName = "PUT"
	}

	baseInfoFormat := "%s, %s, %s"
	baseInfo := fmt.Sprintf(baseInfoFormat, l.Start.Format("2006-01-02 15:04:05.000"), opName, l.Duration.String())

	if l.Error != nil {
		errInfo := fmt.Sprintf("ERROR: %s", l.Error.Error())
		return fmt.Sprintf("%s, %s", baseInfo, errInfo)
	}

	successInfo := fmt.Sprintf("SUCCESS: '%s', %d bytes", l.File, l.Bytes)
	return fmt.Sprintf("%s, %s", baseInfo, successInfo)
}
