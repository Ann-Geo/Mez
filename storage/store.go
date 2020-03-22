package storage

import (
	"os"
	"time"
)

type Timestamp = time.Time
type Image = []byte

type Store interface {
	Recover(recoveryFile *os.File, camid string)
	Append(im Image, t Timestamp) error
	Read(imts chan<- ImageTimestamp, tstart, tstop Timestamp, err chan<- error)
	AppendStats() (int, uint64, Timestamp)
	Backup(fPath string)
}
