package storage

import (
	"sync/atomic"
)

// Returns the closest timestamp to the query key

func binarySearch(tsmemlog *TimestampMemLog, low, high uint64, key Timestamp) (uint64, error) {
	tslog := (*tsmemlog).tslog
	if high == low {
		return high, nil
	}

	mid := low + (high-low)/2

	tslog[row(mid)].lk.RLock() // Read lock segment
	mts := tslog[row(mid)].ts[col(mid)]
	tslog[row(mid)].lk.RUnlock()

	if key.Equal(mts) {
		return mid, nil
	}

	if key.After(mts) {
		return binarySearch(tsmemlog, mid+1, high, key)
	} else {
		return binarySearch(tsmemlog, low, mid-1, key)
	}

}

// Search for index of earliest timestamp in log
func findMinPos(tsmemlog *TimestampMemLog, low, high uint64) uint64 {
	tslog := (*tsmemlog).tslog

	// base cases
	if high == low {
		return high
	}

	var hts Timestamp
	mid := low + (high-low)/2
	tslog[row(mid)].lk.RLock() // Read lock segment
	mts := tslog[row(mid)].ts[col(mid)]
	tslog[row(mid)].lk.RUnlock()

	tslog[row(high)].lk.RLock() // Read lock segment
	hts = tslog[row(high)].ts[col(high)]
	tslog[row(high)].lk.RUnlock()

	if mts.Before(hts) {
		return findMinPos(tsmemlog, low, mid)
	} else {
		return findMinPos(tsmemlog, mid+1, high)
	}
}

// Searches in a circular log

func cyclicBinarySearch(tsmemlog *TimestampMemLog, low, high uint64, key Timestamp) (uint64, error) {
	if (low < 0) || (high > SEGSIZE*LOGSIZE-1) || (low > high) {
		return SEGSIZE * LOGSIZE, ErrSearchRange
	}

	tslog := (*tsmemlog).tslog
	var currpos, minpos uint64

	currpos = atomic.LoadUint64(&tsmemlog.currpos)

	// Find index of earliest timestamp in log
	tslog[row(high)].lk.RLock() // Read lock segment
	hts := tslog[row(high)].ts[col(high)]
	tslog[row(high)].lk.RUnlock()

	if hts.IsZero() { // Partial log
		minpos = findMinPos(tsmemlog, low, currpos)
	} else {
		minpos = findMinPos(tsmemlog, low, high)
	}

	// Check if timestamp to be searched is before earliest timestamp
	tslog[row(minpos)].lk.RLock() // Read lock segment
	mints := tslog[row(minpos)].ts[col(minpos)]
	tslog[row(minpos)].lk.RUnlock()

	// If it is  return earliest timestamp
	if key.Before(mints) {
		return minpos, nil
	}

	// If log is not cyclic, normal binary search
	if minpos == 0 {
		if currpos != high { // Partial log - search until currpos
			return binarySearch(tsmemlog, low, currpos, key)
		} else { // Complete log - search full log
			return binarySearch(tsmemlog, low, high, key)
		}
	}

	// Cyclic log - search two parts separately
	tslog[row(high)].lk.RLock() // Read lock segment
	hts = tslog[row(high)].ts[col(high)]
	tslog[row(high)].lk.RUnlock()

	if key.Before(hts) {
		return binarySearch(tsmemlog, minpos, high, key)
	} else {
		return binarySearch(tsmemlog, low, minpos-1, key)
	}
}
