package storage

import (
	"errors"
)

var (
	ErrSearchRange      = errors.New("Storage: Illegal search range")
	ErrTimestampOrder   = errors.New("Memlog: Timestamp order error")
	ErrTimestampMissing = errors.New("Memlog: Start timestamp not yet in log")
)
