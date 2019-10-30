package storage

import (
	"testing"
	"time"
)

func TestBinarySearch(t *testing.T) {
	tsmemlog := NewTimestampMemLog()
	var query, approxquery, initial Timestamp
	initial = Timestamp(time.Now())
	time.Sleep(100 * time.Millisecond)
	var i, j uint64
	for i = 0; i < LOGSIZE; i++ {
		for j = 0; j < SEGSIZE; j++ {
			tsmemlog.tslog[i].ts[j] = Timestamp(time.Now())
			time.Sleep(100 * time.Millisecond)
			if i == 1 && j == 1 {
				time.Sleep(100 * time.Millisecond)
				approxquery = Timestamp(time.Now())
			}
		}
	}

	// Search for exact timestamp entry
	go func() {
		query = tsmemlog.tslog[1].ts[1]
		pos, _ := binarySearch(tsmemlog, 0, LOGSIZE*SEGSIZE-1, query)
		tres := tsmemlog.tslog[row(pos)].ts[col(pos)]
		if !tres.Equal(query) {
			t.Errorf("Error - exact entry not found - query %v, tres %v", query, tres)
		}
	}()

	// Search for non existent timestamp entry in the future
	go func() {
		query = Timestamp(time.Now())
		pos, _ := binarySearch(tsmemlog, 0, LOGSIZE*SEGSIZE-1, query)
		tres := tsmemlog.tslog[row(pos)].ts[col(pos)]
		if !tres.Equal(tsmemlog.tslog[LOGSIZE-1].ts[SEGSIZE-1]) {
			t.Errorf("Error - non existent future query failed to report highest timestamp - query %v, tres %v", query, tres)
		}
	}()

	// Search for non existent timestamp entry in the past
	go func() {
		query = initial
		pos, _ := binarySearch(tsmemlog, 0, LOGSIZE*SEGSIZE-1, query)
		tres := tsmemlog.tslog[row(pos)].ts[col(pos)]
		if !tres.Equal(tsmemlog.tslog[0].ts[0]) {
			t.Errorf("Error - non existent past query failed to report highest timestamp - query %v, tres %v", query, tres)
		}
	}()

	// Search for inexact time stamp
	go func() {
		query = approxquery
		pos, _ := binarySearch(tsmemlog, 0, LOGSIZE*SEGSIZE-1, query)
		tres := tsmemlog.tslog[row(pos)].ts[col(pos)]
		if !tres.Before(tres.Add(2*time.Second)) || !(tres.After(tres.Add(-2 * time.Second))) {
			//t.Errorf("Error: Incorrect bounds for inexact search -  query %v, tl %v, th %v", query, tres)
			t.Errorf("Error: Incorrect bounds for inexact search -  query %v, th %v", query, tres)
		}
	}()

}

func TestCylicSearch(t *testing.T) {
	tsmemlog := NewTimestampMemLog()
	var query, tres Timestamp
	var pos uint64

	var i uint64

	tbefore := time.Now()
	time.Sleep(100 * time.Millisecond)

	// Second half has earlier time stamps
	for i = LOGSIZE * SEGSIZE / 2; i < LOGSIZE*SEGSIZE; i++ {
		tsmemlog.tslog[row(i)].ts[col(i)] = Timestamp(time.Now())
		time.Sleep(100 * time.Millisecond)
	}

	for i = 0; i < LOGSIZE*SEGSIZE/2; i++ {
		tsmemlog.tslog[row(i)].ts[col(i)] = Timestamp(time.Now())
		time.Sleep(100 * time.Millisecond)
	}

	time.Sleep(100 * time.Millisecond)
	tafter := time.Now()

	// Search for exact timestamp entry from second half
	go func() {
		query = tsmemlog.tslog[row(LOGSIZE*SEGSIZE/2+4)].ts[col(LOGSIZE*SEGSIZE/2+4)]
		pos, _ = cyclicBinarySearch(tsmemlog, 0, LOGSIZE*SEGSIZE-1, query)
		tres = tsmemlog.tslog[row(pos)].ts[col(pos)]
		if !tres.Equal(query) {
			t.Errorf("Error - exact entry not found in second half - query %v, tres %v", query, tres)
		}
	}()

	// Search for exact timestamp entry from first half
	go func() {
		query = tsmemlog.tslog[row(LOGSIZE*SEGSIZE/2-1)].ts[col(LOGSIZE*SEGSIZE/2-1)]
		pos, _ = cyclicBinarySearch(tsmemlog, 0, LOGSIZE*SEGSIZE-1, query)
		tres = tsmemlog.tslog[row(pos)].ts[col(pos)]
		if !tres.Equal(query) {
			t.Errorf("Error - exact entry not found in first half - query %v, tres %v", query, tres)
		}
	}()

	// Search for a timestamp that's earlier than those in the log. Should return min timeestamp
	go func() {
		query = tbefore
		pos, _ = cyclicBinarySearch(tsmemlog, 0, LOGSIZE*SEGSIZE-1, query)
		tres = tsmemlog.tslog[row(LOGSIZE*SEGSIZE/2)].ts[col(LOGSIZE*SEGSIZE/2)]
		if tres.Before(query) {
			t.Errorf("Error - min timestamp not returned - query %v, tres %v", query, tres)
		}
	}()

	// Search for a timestamp that's later than those in the log. Should return max timeestamp
	go func() {
		query = tafter
		pos, _ = cyclicBinarySearch(tsmemlog, 0, LOGSIZE*SEGSIZE-1, query)
		tres = tsmemlog.tslog[row(LOGSIZE*SEGSIZE/2-1)].ts[col(LOGSIZE*SEGSIZE/2-1)]
		if tres.After(query) {
			t.Errorf("Error - min timestamp not returned - query %v, tres %v", query, tres)
		}
	}()

}
