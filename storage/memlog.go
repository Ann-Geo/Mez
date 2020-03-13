/* Implementation of in-memory timestamp and image log
Supports a single writer, and multiple readers
Log is organized as an array of segments
Each segment is protected by a read-write lock
*/

package storage

import (
	"sync"
	"sync/atomic"
	"time"
)

var SEGSIZE uint64 = 20
var LOGSIZE uint64 = 50

type ImageTimestamp struct {
	Im Image
	Ts Timestamp
}

// Timestamp log
type TimestampLogSeg struct {
	ts []Timestamp
	lk *sync.RWMutex
}

type TimestampMemLog struct {
	tslog   []TimestampLogSeg
	currpos uint64 // To be read/written atomically
}

// Image log
// Not explicitly protected by a lock, but should be used with timstamp lock
type ImageLogSeg struct {
	im []Image
}

type ImageMemLog struct {
	imlog []ImageLogSeg
}

// Mem log is composed of timstamp and image logs
type MemLog struct {
	tsmemlog *TimestampMemLog
	immemlog *ImageMemLog
}

// Constructors
func NewTimestampMemLog() *TimestampMemLog {
	var tsmemlog TimestampMemLog
	tsmemlog.tslog = make([]TimestampLogSeg, LOGSIZE)
	var i uint64
	for i = 0; i < LOGSIZE; i++ {
		tsmemlog.tslog[i].ts = make([]Timestamp, SEGSIZE)
		tsmemlog.tslog[i].lk = &sync.RWMutex{}
	}
	tsmemlog.currpos = 0 //SEGSIZE * LOGSIZE //Initially outside range
	return &tsmemlog
}

func NewImageMemLog() *ImageMemLog {
	var immemlog ImageMemLog
	immemlog.imlog = make([]ImageLogSeg, LOGSIZE)
	var i uint64
	for i = 0; i < LOGSIZE; i++ {
		immemlog.imlog[i].im = make([]Image, SEGSIZE)
	}
	return &immemlog
}

func NewMemLog(segsize, logsize uint64) *MemLog {
	SEGSIZE = segsize
	LOGSIZE = logsize
	return &MemLog{tsmemlog: NewTimestampMemLog(),
		immemlog: NewImageMemLog()}
}

// Store methods
// Apppend: Input - image and associated timestamp
func (memlog *MemLog) Append(im Image, t Timestamp) error {
	var pos = atomic.LoadUint64(&memlog.tsmemlog.currpos)
	if pos != SEGSIZE*LOGSIZE {
		memlog.tsmemlog.tslog[row(pos)].lk.RLock()
		tprev := memlog.tsmemlog.tslog[row(pos)].ts[col(pos)]
		memlog.tsmemlog.tslog[row(pos)].lk.RUnlock()
		if t.Before(tprev) {
			return ErrTimestampOrder
		}
	}
	if pos == SEGSIZE*LOGSIZE || pos == SEGSIZE*LOGSIZE-1 {
		pos = 0
	} else {
		pos++
	}

	// Insert timestamp and image into log
	memlog.tsmemlog.tslog[row(pos)].lk.Lock()
	memlog.tsmemlog.tslog[row(pos)].ts[col(pos)] = t
	memlog.immemlog.imlog[row(pos)].im[col(pos)] = append([]byte(nil), im...)
	memlog.tsmemlog.tslog[row(pos)].lk.Unlock()

	// Update currpos
	atomic.StoreUint64(&memlog.tsmemlog.currpos, pos)

	return nil
}

// Read: Input - start and stop timestamps. Output - slice of images, timestamps of first and last images
func (memlog *MemLog) Read(imts chan<- ImageTimestamp, tstart, tstop Timestamp, err chan<- error) {

	if tstart.After(tstop) {
		err <- ErrTimestampOrder
		return
	}

	currpos := atomic.LoadUint64(&memlog.tsmemlog.currpos)
	//fmt.Println("currpos", currpos)
	memlog.tsmemlog.tslog[row(currpos)].lk.RLock()
	tcurr := memlog.tsmemlog.tslog[row(currpos)].ts[col(currpos)]
	memlog.tsmemlog.tslog[row(currpos)].lk.RUnlock()

	//fmt.Println("tstart ---", tstart)
	//fmt.Println("tcurr ----", tcurr)

	if tstart.After(tcurr) { // No entries yet
		err <- ErrTimestampMissing
		return
	}

	start, _ := cyclicBinarySearch(memlog.tsmemlog, 0, LOGSIZE*SEGSIZE-1, tstart)
	stop, _ := cyclicBinarySearch(memlog.tsmemlog, 0, LOGSIZE*SEGSIZE-1, tstop)

	if start <= stop {
		memlog.readHelper(imts, start, stop)
	} else { // Wrap around
		memlog.readHelper(imts, start, LOGSIZE*SEGSIZE-1)
		memlog.readHelper(imts, 0, stop)
	}

	err <- nil
}

// Helper to read the timestamp and image logs
func (memlog *MemLog) readHelper(imts chan<- ImageTimestamp, start, stop uint64) {

	for pos := start; pos <= stop; pos++ {
		memlog.tsmemlog.tslog[row(pos)].lk.RLock()
		image := make(Image, len(memlog.immemlog.imlog[row(pos)].im[col(pos)]))
		copy(image, memlog.immemlog.imlog[row(pos)].im[col(pos)])
		it := ImageTimestamp{Im: image, Ts: memlog.tsmemlog.tslog[row(pos)].ts[col(pos)]}
		imts <- it
		memlog.tsmemlog.tslog[row(pos)].lk.RUnlock() // Release lock after each image is read
	}
}

func (memlog *MemLog) AppendStats() (int, uint64, Timestamp) {
	pos := atomic.LoadUint64(&memlog.tsmemlog.currpos)
	if pos != SEGSIZE*LOGSIZE {
		memlog.tsmemlog.tslog[row(pos)].lk.RLock()
		defer memlog.tsmemlog.tslog[row(pos)].lk.RUnlock()
		t := memlog.tsmemlog.tslog[row(pos)].ts[col(pos)]
		return len(memlog.immemlog.imlog[row(pos)].im[col(pos)]), pos, t
	}
	return 0, SEGSIZE * LOGSIZE, time.Time{}
}

func row(pos uint64) uint64 {
	return pos / SEGSIZE
}

func col(pos uint64) uint64 {
	return pos % SEGSIZE
}
