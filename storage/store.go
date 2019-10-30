package storage

import (
	"time"
)

type Timestamp = time.Time
type Image = []byte

type Store interface {
	//Init()
	//Recover()
	Append(im Image, t Timestamp) error
	Read(imts chan<- ImageTimestamp, tstart, tstop Timestamp, err chan<- error)
	AppendStats() (int, uint64, Timestamp)
}
