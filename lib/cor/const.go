// Copyright (c) 2019 by Matthew James Briggs, https://github.com/webern

package cor

// https://tools.ietf.org/html/rfc1350

const (
	BlockSize = 512
)

type ErrCode uint16

const (
	ErrUnknown  ErrCode = 0 // Not defined, see error message (if any).
	ErrNotFound         = 1 // File not found.
	ErrAccess           = 2 // Access violation.
	ErrDisk             = 3 // Disk full or allocation exceeded.
	ErrBadOp            = 4 // Illegal TFTP operation.
	ErrBadID            = 5 // Unknown transfer ID.
	ErrDupFile          = 6 // File already exists.
	ErrUnkUser          = 7 // No such user.
)

func (e ErrCode) String() string {
	switch e {
	case ErrUnknown:
		return "E_UNKNOWN"
	case ErrNotFound:
		return "E_NOT_FOUND"
	case ErrAccess:
		return "E_ACCESS"
	case ErrDisk:
		return "E_DISK"
	case ErrBadOp:
		return "E_BAD_OP"
	case ErrBadID:
		return "E_BAD_ID"
	case ErrDupFile:
		return "E_DUP_FILE"
	case ErrUnkUser:
		return "E_UNK_USER"
	default:
		break
	}

	return "E_UNKNOWN"
}
