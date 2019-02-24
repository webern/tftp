// Copyright (c) 2019 by Matthew James Briggs, https://github.com/webern

package cor

type OpType uint16

const (
	OpRRQ   OpType = 1 // Read Request
	OpWRQ          = 2 // Write Request
	OpData         = 3 // Data Packet
	OpAck          = 4 // Acknowledgement
	OpError        = 5 // Error Packet
)

func (o OpType) String() string {
	switch o {
	case OpRRQ:
		return "READ"
	case OpWRQ:
		return "WRIT"
	case OpData:
		return "DATA"
	case OpAck:
		return "ACKN"
	case OpError:
		return "ERRO"
	default:
		break
	}

	return "UNKN"
}
