package storage

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestAppendRead(t *testing.T) {
	var (
		imBuf          []byte
		imageFilesPath = "../test_images"
		err            error
	)

	memlog := NewMemLog(2, 8)
	start := time.Now()
	time.Sleep(100 * time.Millisecond)

	// Read image file names from directory
	var files []string
	absImageFilePath, _ := filepath.Abs(imageFilesPath)
	err = filepath.Walk(absImageFilePath, func(path string, info os.FileInfo, err error) error {
		files = append(files, path)
		return nil
	})
	if err != nil {
		panic(err)
	}

	// Load a single image file to buffer
	fd, err := os.Open(files[1])
	if err != nil {
		panic(err)
	}
	defer fd.Close()
	fi, err := fd.Stat()
	imBuf = make([]byte, fi.Size())
	_, err = fd.Read(imBuf)
	if err != nil {
		t.Errorf("errored while copying from file to buf %v", err)
	}

	// Append to log
	var i uint64
	for i = 0; i < LOGSIZE*SEGSIZE; i++ {
		ts := time.Now()
		memlog.Append(Image(imBuf), ts)
		time.Sleep(100 * time.Millisecond)
		sz, pos, tins := memlog.AppendStats()
		t.Logf("Image of size %d inserted in log at position %d at time %v\n", sz, pos, tins)

	}
	// Read from log with input timestamp range outside the bounds of the log
	time.Sleep(100 * time.Millisecond)
	stop := time.Now()
	t.Log("Read Start End", start, stop)

	imts := make(chan ImageTimestamp, 200*1024)
	errch := make(chan error)
	defer close(imts)
	defer close(errch)

	go memlog.Read(imts, start, stop, errch)

	errc := <-errch
	if errc != nil {
		t.Errorf("Read error %s", errc)
	} else {
		ok := true
		for ok {
			select {
			case it := <-imts:
				t.Logf("Size of image is %d and timestamp is %v \n", len(it.Im), it.Ts)
				ok = true

			default:
				ok = false
			}
		}
	}
}

/*
// Test concurrent append and reads
func TestConcurrentAppendRead(t *testing.T) {
	// Note: Issues with printing from go threads in test. Currently printing only from main thread
	var (
		imBuf          []byte
		imageFilesPath = "../test_images"
		err            error
	)

	memlog := NewMemLog(2, 8)
	start := time.Now()

	// Read image file names from directory
	var files []string
	absImageFilePath, _ := filepath.Abs(imageFilesPath)
	err = filepath.Walk(absImageFilePath, func(path string, info os.FileInfo, err error) error {
		files = append(files, path)
		return nil
	})
	if err != nil {
		panic(err)
	}

	// Load a single image file to buffer
	fd, err := os.Open(files[1])
	if err != nil {
		panic(err)
	}
	defer fd.Close()
	fi, err := fd.Stat()
	imBuf = make([]byte, fi.Size())
	_, err = fd.Read(imBuf)
	if err != nil {
		t.Errorf("errored while copying from file to buf %v", err)
	}

	// Append to log in an infinite loop
	go func() {
		for {
			ts := time.Now()
			memlog.Append(Image(imBuf), ts)
			time.Sleep(100 * time.Millisecond)
			//	sz, pos, tins := memlog.AppendStats()

		}
	}()

	// Read from log
	time.Sleep(1 * time.Second)
	stop := time.Now()
	imts := make(chan ImageTimestamp, 200*1024)
	errc := make(chan error)

	go memlog.Read(imts, start, stop, errc)

	err = <-errc
	if err != nil {
		t.Log(err)
	} else {
		for imts := range imts {
			t.Logf("Size of image is %d and timestamp is %v \n", len(imts.Im), imts.Ts)
		}
	}

}
*/
