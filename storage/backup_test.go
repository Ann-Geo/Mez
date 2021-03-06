package storage

import (
	"testing"
	"time"

	"gocv.io/x/gocv"
)

func TestBackupSameSizeImage(t *testing.T) {

	var (
		imBuf []byte
	)
	var tests = []struct {
		imageFilesPath  string
		numImagesInsert uint64
		logSize         uint64
		segSize         uint64
		frameRate       int16
	}{
		{"../../test_images/500K/", 100, 5, 10, 200},
	}

	for _, test := range tests {
		//new memlog
		memlog := NewMemLog(test.segSize, test.logSize)

		go memlog.Backup("/home/research/goworkspace/src/github.com/Ann-Geo/store_enb/")

		//Obtain filenames in the directory path given
		errMsg, fileList := walkAllFilesInDir(test.imageFilesPath)
		if errMsg != "file read success" {
			t.Fatalf("File read failed - %v\n", errMsg)
		}

		//read first file to buffer (conversion to bytes)
		buffer := gocv.IMRead(fileList[0], gocv.IMReadColor)
		imBuf = buffer.ToBytes()

		//slices to store timestamps and image sizes appended
		tsAppended := make([]time.Time, 0)
		imSizeAppended := make([]int, 0)

		//Append image sizes to the slice
		var i uint64
		for i = 0; i < test.numImagesInsert; i++ {
			imSizeAppended = append(imSizeAppended, len(imBuf))
		}

		// Append to log
		for i = 0; i < test.numImagesInsert; i++ {
			ts := time.Now()
			tsAppended = append(tsAppended, ts)
			memlog.Append(Image(imBuf), ts)
			time.Sleep(time.Duration(test.frameRate) * time.Millisecond)

		}

		//verifying the timestamps and image sizes directly from log
		//count variable to keep track of number of elements in log
		var count uint64
		//slices to read timestamps and image sizes
		tsRead := make([]time.Time, 0)
		imSizeRead := make([]int, 0)
		var j uint64
		for i = 0; i < test.logSize; i++ {
			for j = 0; j < test.segSize; j++ {
				if count == test.numImagesInsert {
					break
				}

				tsRead = append(tsRead, memlog.tsmemlog.tslog[i].ts[j])
				imSizeRead = append(imSizeRead, len(memlog.immemlog.imlog[i].im[j]))
				count = count + 1
			}

		}

		errMsg, status := sliceEquality(tsAppended, tsRead, imSizeAppended, imSizeRead)
		if !(status) {
			t.Errorf("Memlog append failed - %v\n", errMsg)
		}

	}

}
