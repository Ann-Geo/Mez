package storage

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

//***************************Helper functions start*****************************************

/*
To return list of absolute paths of files in a directory
This helper function is used to read the test images from input directory
Input - directory path
Output- error message, list of filepaths
*/
func walkAllFilesInDir(dir string) (string, []string) {
	var errMsg string = "file read success"
	fileList := make([]string, 0)
	_ = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			errMsg = "Incorrect file path"
		}

		// check if it is a regular file (not dir)
		if info.Mode().IsRegular() {
			fileList = append(fileList, path)

		}
		return nil
	})

	return errMsg, fileList
}

/*To compare equality of timestamp slices (Appended and Read)
and to compare equality of image size slices (Appended and Read)
Input - Appended and Read slices
Output- true or false boolean
*/
func sliceEquality(tsAppended, tsRead []time.Time, imSizeAppended, imSizeRead []int) (string, bool) {

	var errMsg string
	var status bool = true
	if len(tsAppended) != len(tsRead) {
		errMsg = "timestamp logs length mismatch"
		status = false

	}

	if len(imSizeAppended) != len(imSizeRead) {
		errMsg = "image logs length mismatch"
		status = false

	}

	for i := 0; i < len(tsAppended); i++ {
		if tsAppended[i] != tsRead[i] {
			errMsg = "timestamp logs contents mismatch"
			status = false

		}
		if imSizeAppended[i] != imSizeRead[i] {
			errMsg = "image logs contents mismatch"
			status = false

		}
	}

	if status == false {
		return errMsg, status
	} else {
		return "slices are equal", true
	}

}

//To Read images sent by memlog.Read function from channel
//All readers have same Tstart and Tend
func readFromChannelSameStartEnd(readerId uint64, imts chan ImageTimestamp, errch chan error, t *testing.T, done chan bool,
	stop time.Time, fill string, noIm, ss, ls uint64, readErrMsg chan string) {
	errc := <-errch

	//To count total number of images read
	var readImageCount uint64
	//last read time stamp from the log
	var lastTsRead time.Time

	if errc != nil {
		t.Fatalf("Memlog Read internal error - %s", errc)
	} else {
		ok := true
		for ok {
			select {
			case it := <-imts:
				readImageCount++
				lastTsRead = it.Ts
				ok = true

			default:
				ok = false
			}
		}
	}

	errMsg := "memlog read success"
	if fill == "NO" {
		if !(stop.After(lastTsRead)) {
			errMsg = "did not read until tstop"
		} else if !(readImageCount == noIm) {
			errMsg = "#images read do not match #images inserted"
		}
	} else {
		if !(stop.After(lastTsRead)) {
			errMsg = "did not read until tstop"
		} else if !(readImageCount == ss*ls) {
			errMsg = "#images read do not match #images inserted"
		}

	}

	done <- true
	readErrMsg <- errMsg
}

type rIdImCount struct {
	//done bool
	readerId   uint64
	readImages uint64
}

//To Read images sent by memlog.Read function from channel
//For readers with dissimilar Tstart and Tend
func readFromChannelDissimilarStartEnd(readerId uint64, imts chan ImageTimestamp, errch chan error, t *testing.T,
	stop time.Time, fill string, noIm, ss, ls uint64, idcount chan rIdImCount) {

	errc := <-errch

	//To count total number of images read
	var readImageCount uint64

	if errc != nil {
		t.Fatalf("Memlog Read internal error - %s", errc)
	} else {
		ok := true
		for ok {
			select {
			case _ = <-imts:
				readImageCount++
				ok = true

			default:
				ok = false
			}
		}
	}

	idcount <- rIdImCount{readerId: readerId, readImages: readImageCount}
}

//To Read images sent by memlog.Read function from channel
//Readers read images concurrently with Append
//All readers have same Tstart and Tend
func readFromChannelConcurrent(readerId uint64, imts chan ImageTimestamp, errch chan error, t *testing.T, readIms chan uint64,
	stop time.Time, fill string, noIm, ss, ls uint64) {
	errc := <-errch

	//To count total number of images read
	var readImageCount uint64

	if errc != nil {
		t.Fatalf("Memlog Read internal error - %s", errc)
	} else {
		ok := true
		for ok {
			select {
			case _ = <-imts:
				readImageCount++
				ok = true

			default:
				ok = false
			}
		}
	}
	readIms <- readImageCount
}

//**********************************Helper functions end***********************************

//***********************************Storage tests start************************************

/*
Tests sstorage Append to store same sized images
- at 30, 24 and 15fps rate
- for partially written and fully written logs
*/
func TestAppendSameSizeImage(t *testing.T) {

	var (
		imBuf []byte
		err   error
	)
	var tests = []struct {
		imageFilesPath  string
		numImagesInsert uint64
		logSize         uint64
		segSize         uint64
		frameRate       int16
	}{
		{"../../test_images/2.1M/", 1000, 10, 100, 33},
		{"../../test_images/2.1M/", 500, 10, 100, 33},
		{"../../test_images/2.1M/", 1000, 10, 100, 42},
		{"../../test_images/2.1M/", 500, 10, 100, 42},
		{"../../test_images/2.1M/", 1000, 10, 100, 67},
		{"../../test_images/2.1M/", 500, 10, 100, 67},
		{"../../test_images/1.5M/", 1000, 10, 100, 33},
		{"../../test_images/1.5M/", 500, 10, 100, 33},
		{"../../test_images/1.5M/", 1000, 10, 100, 42},
		{"../../test_images/1.5M/", 500, 10, 100, 42},
		{"../../test_images/1.5M/", 1000, 10, 100, 67},
		{"../../test_images/1.5M/", 500, 10, 100, 67},
		{"../../test_images/1.2M/", 1000, 10, 100, 33},
		{"../../test_images/1.2M/", 500, 10, 100, 33},
		{"../../test_images/1.2M/", 1000, 10, 100, 42},
		{"../../test_images/1.2M/", 500, 10, 100, 42},
		{"../../test_images/1.2M/", 1000, 10, 100, 67},
		{"../../test_images/1.2M/", 500, 10, 100, 67},
		{"../../test_images/1.0M/", 1000, 10, 100, 33},
		{"../../test_images/1.0M/", 500, 10, 100, 33},
		{"../../test_images/1.0M/", 1000, 10, 100, 42},
		{"../../test_images/1.0M/", 500, 10, 100, 42},
		{"../../test_images/1.0M/", 1000, 10, 100, 67},
		{"../../test_images/1.0M/", 500, 10, 100, 67},
		{"../../test_images/800K/", 1000, 10, 100, 33},
		{"../../test_images/800K/", 500, 10, 100, 33},
		{"../../test_images/800K/", 1000, 10, 100, 42},
		{"../../test_images/800K/", 500, 10, 100, 42},
		{"../../test_images/800K/", 1000, 10, 100, 67},
		{"../../test_images/800K/", 500, 10, 100, 67},
		{"../../test_images/500K/", 1000, 10, 100, 33},
		{"../../test_images/500K/", 500, 10, 100, 33},
		{"../../test_images/500K/", 1000, 10, 100, 42},
		{"../../test_images/500K/", 500, 10, 100, 42},
		{"../../test_images/500K/", 1000, 10, 100, 67},
		{"../../test_images/500K/", 500, 10, 100, 67},
		{"../../test_images/200K/", 1000, 10, 100, 33},
		{"../../test_images/200K/", 500, 10, 100, 33},
		{"../../test_images/200K/", 1000, 10, 100, 42},
		{"../../test_images/200K/", 500, 10, 100, 42},
		{"../../test_images/200K/", 1000, 10, 100, 67},
		{"../../test_images/200K/", 500, 10, 100, 67},
		{"../../test_images/100K/", 1000, 10, 100, 33},
		{"../../test_images/100K/", 500, 10, 100, 33},
		{"../../test_images/100K/", 1000, 10, 100, 42},
		{"../../test_images/100K/", 500, 10, 100, 42},
		{"../../test_images/100K/", 1000, 10, 100, 67},
		{"../../test_images/100K/", 500, 10, 100, 67},
		{"../../test_images/50K/", 1000, 10, 100, 33},
		{"../../test_images/50K/", 500, 10, 100, 33},
		{"../../test_images/50K/", 1000, 10, 100, 42},
		{"../../test_images/50K/", 500, 10, 100, 42},
		{"../../test_images/50K/", 1000, 10, 100, 67},
		{"../../test_images/50K/", 500, 10, 100, 67},
		{"../../test_images/10K/", 1000, 10, 100, 33},
		{"../../test_images/10K/", 500, 10, 100, 33},
		{"../../test_images/10K/", 1000, 10, 100, 42},
		{"../../test_images/10K/", 500, 10, 100, 42},
		{"../../test_images/10K/", 1000, 10, 100, 67},
		{"../../test_images/10K/", 500, 10, 100, 67},
	}

	for _, test := range tests {
		//new memlog
		memlog := NewMemLog(test.segSize, test.logSize)

		//Obtain filenames in the directory path given
		errMsg, fileList := walkAllFilesInDir(test.imageFilesPath)
		if errMsg != "file read success" {
			t.Fatalf("File read failed - %v\n", errMsg)
		}

		//read first file to buffer (conversion to bytes)
		imBuf, err = ioutil.ReadFile(fileList[0])
		if err != nil {
			t.Fatalf("Cannot read image file %v\n", err)

		}

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
			//Specify frame rate here
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

/*
Tests sstorage Append to store same sized images
- at 30, 24 and 15fps rate
- for over written log
*/

func TestAppendSameImSizeOverWrittenLog(t *testing.T) {

	var (
		imBuf []byte
		err   error
	)
	var tests = []struct {
		imageFilesPath  string
		numImagesInsert uint64
		logSize         uint64
		segSize         uint64
		frameRate       int16
	}{
		{"../../test_images/2.1M/", 545, 10, 30, 33},
		{"../../test_images/1.5M/", 545, 10, 30, 33},
		{"../../test_images/1.2M/", 545, 10, 30, 33},
		{"../../test_images/1.0M/", 545, 10, 30, 33},
		{"../../test_images/800K/", 545, 10, 30, 33},
		{"../../test_images/500K/", 545, 10, 30, 33},
		{"../../test_images/200K/", 545, 10, 30, 33},
		{"../../test_images/100K/", 545, 10, 30, 33},
		{"../../test_images/50K/", 545, 10, 30, 33},
		{"../../test_images/10K/", 545, 10, 30, 33},
		{"../../test_images/2.1M/", 545, 10, 30, 42},
		{"../../test_images/1.5M/", 545, 10, 30, 42},
		{"../../test_images/1.2M/", 545, 10, 30, 42},
		{"../../test_images/1.0M/", 545, 10, 30, 42},
		{"../../test_images/800K/", 545, 10, 30, 42},
		{"../../test_images/500K/", 545, 10, 30, 42},
		{"../../test_images/200K/", 545, 10, 30, 42},
		{"../../test_images/100K/", 545, 10, 30, 42},
		{"../../test_images/50K/", 545, 10, 30, 42},
		{"../../test_images/10K/", 545, 10, 30, 42},
		{"../../test_images/2.1M/", 545, 10, 30, 67},
		{"../../test_images/1.5M/", 545, 10, 30, 67},
		{"../../test_images/1.2M/", 545, 10, 30, 67},
		{"../../test_images/1.0M/", 545, 10, 30, 67},
		{"../../test_images/800K/", 545, 10, 30, 67},
		{"../../test_images/500K/", 545, 10, 30, 67},
		{"../../test_images/200K/", 545, 10, 30, 67},
		{"../../test_images/100K/", 545, 10, 30, 67},
		{"../../test_images/50K/", 545, 10, 30, 67},
		{"../../test_images/10K/", 545, 10, 30, 67},
	}

	for _, test := range tests {

		//new memlog
		memlog := NewMemLog(test.segSize, test.logSize)

		//Obtain filenames in the directory path given
		errMsg, fileList := walkAllFilesInDir(test.imageFilesPath)
		if errMsg != "file read success" {
			t.Fatalf("File read failed - %v\n", errMsg)
		}

		//read first file to buffer (conversion to bytes)
		imBuf, err = ioutil.ReadFile(fileList[0])
		if err != nil {
			t.Fatalf("Cannot read image file %v\n", err)

		}

		//slices to store timestamps and image sizes appended
		tsAppended := make([]time.Time, 0)
		imSizeAppended := make([]int, 0)

		//Append image sizes to the slice
		var i uint64
		for i = test.segSize * test.logSize; i < test.numImagesInsert; i++ {
			imSizeAppended = append(imSizeAppended, len(imBuf))
		}

		// Append to log
		for i = 0; i < test.numImagesInsert; i++ {
			ts := time.Now()

			if i >= test.segSize*test.logSize {
				tsAppended = append(tsAppended, ts)
			}

			memlog.Append(Image(imBuf), ts)
			//Specify frame rate here
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
				if count == (test.numImagesInsert - (test.segSize * test.logSize)) {
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

/*
Tests storage Append to store variable sized images
- at 30, 24, 15fps rate
- for partially written and fully written logs
*/
func TestAppendVariableSizeImage(t *testing.T) {

	var (
		imBuf []byte
		err   error
	)
	var tests = []struct {
		imageFilesPath  string
		numImagesInsert uint64
		logSize         uint64
		segSize         uint64
		frameRate       int16
	}{
		{"../../test_images/1000_images/", 1000, 10, 100, 33},
		{"../../test_images/1000_images/", 500, 10, 100, 33},
		{"../../test_images/1000_images/", 1000, 10, 100, 42},
		{"../../test_images/1000_images/", 500, 10, 100, 42},
		{"../../test_images/1000_images/", 1000, 10, 100, 67},
		{"../../test_images/1000_images/", 500, 10, 100, 67},
	}

	for _, test := range tests {

		//new memlog
		memlog := NewMemLog(test.segSize, test.logSize)

		//Obtain filenames in the directory path given
		errMsg, fileList := walkAllFilesInDir(test.imageFilesPath)
		if errMsg != "file read success" {
			t.Fatalf("File read failed - %v\n", errMsg)
		}

		//slices to store timestamps and image sizes appended
		tsAppended := make([]time.Time, 0)
		imSizeAppended := make([]int, 0)

		//read all files to buffer (conversion to bytes)
		//Append image sizes to the slice
		//Append images to log
		//read images count
		var readImCount uint64
		var i uint64
		for _, file := range fileList {
			if readImCount == test.numImagesInsert {
				break
			}
			imBuf, err = ioutil.ReadFile(file)
			imSizeAppended = append(imSizeAppended, len(imBuf))
			if err != nil {
				t.Fatalf("Cannot read image file %v\n", err)

			}
			readImCount++

			ts := time.Now()
			tsAppended = append(tsAppended, ts)
			memlog.Append(Image(imBuf), ts)
			//Specify frame rate here
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

/*
Tests storage Append to store variable sized images
- at 30, 24, 15fps rate
- for over written log
*/

func TestAppendVarImSizeOverWrittenLog(t *testing.T) {

	var (
		imBuf []byte
		err   error
	)
	var tests = []struct {
		imageFilesPath  string
		numImagesInsert uint64
		logSize         uint64
		segSize         uint64
		frameRate       int16
	}{
		{"../../test_images/1000_images/", 545, 10, 30, 33},
		{"../../test_images/1000_images/", 545, 10, 30, 42},
		{"../../test_images/1000_images/", 545, 10, 30, 67},
	}

	for _, test := range tests {

		//new memlog
		memlog := NewMemLog(test.segSize, test.logSize)

		//Obtain filenames in the directory path given
		errMsg, fileList := walkAllFilesInDir(test.imageFilesPath)
		if errMsg != "file read success" {
			t.Fatalf("File read failed - %v\n", errMsg)
		}

		//slices to store timestamps and image sizes appended
		tsAppended := make([]time.Time, 0)
		imSizeAppended := make([]int, 0)

		//read all files to buffer (conversion to bytes)
		//Append image sizes to the slice
		//Append to log
		//read images count
		var readImCount uint64
		var i uint64
		for _, file := range fileList {
			ts := time.Now()
			if readImCount == test.numImagesInsert {
				break
			}
			imBuf, err = ioutil.ReadFile(file)
			if readImCount >= test.segSize*test.logSize {
				imSizeAppended = append(imSizeAppended, len(imBuf))
				tsAppended = append(tsAppended, ts)
			}
			if err != nil {
				t.Fatalf("Cannot read image file %v\n", err)

			}
			readImCount++

			memlog.Append(Image(imBuf), ts)
			//Specify frame rate here
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
				if count == (test.numImagesInsert - (test.segSize * test.logSize)) {
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

/*
Tests storage Read for already appended images
- for images with size 2.1M to 1K
- for partially, fully and over written logs
*/

func TestReadAppended(t *testing.T) {
	var (
		imBuf []byte
		err   error
	)
	var tests = []struct {
		imageFilesPath  string
		numImagesInsert uint64
		logSize         uint64
		segSize         uint64
		frameRate       uint64
		fillType        string
	}{
		{"../../test_images/1000_images/", 1000, 10, 100, 33, "NO"},
		{"../../test_images/1000_images/", 500, 10, 100, 33, "NO"},
		{"../../test_images/1000_images/", 545, 10, 30, 33, "O"},
		{"../../test_images/1000_images/", 545, 10, 30, 33, "O2"},
		{"../../test_images/1000_images/", 15, 2, 5, 33, "O1"},
		{"../../test_images/1000_images/", 15, 2, 5, 33, "O2"},
	}

	for _, test := range tests {

		//new memlog
		memlog := NewMemLog(test.segSize, test.logSize)

		//Obtain filenames in the directory path given
		errMsg, fileList := walkAllFilesInDir(test.imageFilesPath)
		if errMsg != "file read success" {
			t.Fatalf("File read failed - %v\n", errMsg)
		}

		start := time.Now()

		var tsFirstOverWrittenIndex time.Time
		var tsLastImage time.Time

		go func() {

			//read all files to buffer (conversion to bytes)
			//Append image sizes to the slice
			//Append to log
			//read images count
			ts := time.Now()
			var readImCount uint64
			for _, file := range fileList {

				if readImCount == test.numImagesInsert {
					tsLastImage = ts
					break
				}
				imBuf, err = ioutil.ReadFile(file)

				if err != nil {
					t.Fatalf("Cannot read image file %v\n", err)

				}
				readImCount++
				ts = time.Now()
				if readImCount == (test.segSize*test.logSize)+1 {
					tsFirstOverWrittenIndex = ts
				}

				//fmt.Println(ts)
				memlog.Append(Image(imBuf), ts)
				//Specify frame rate here
				time.Sleep(time.Duration(test.frameRate) * time.Millisecond)

			}
		}()

		if test.numImagesInsert == 15 {
			time.Sleep(1000 * time.Millisecond)
		} else {
			time.Sleep(time.Duration((test.numImagesInsert*test.frameRate)+2000) * time.Millisecond)
		}

		stop := time.Now()
		if test.fillType == "O1" {
			start = tsFirstOverWrittenIndex.Add(10 * time.Millisecond)
			stop = tsLastImage.Add(10 * time.Millisecond)
		} else if test.fillType == "O2" {
			start = tsFirstOverWrittenIndex.Add(10 * time.Millisecond)
			stop = tsLastImage.Add(-10 * time.Millisecond)
		}

		imts := make(chan ImageTimestamp, 200*1024)
		errch := make(chan error)
		defer close(imts)
		defer close(errch)

		go memlog.Read(imts, start, stop, errch)

		errc := <-errch

		//To count total number of images read
		var readImageCount uint64
		//last read time stamp from the log
		var lastTsRead time.Time

		if errc != nil {
			t.Errorf("Read error %s", errc)
		} else {
			ok := true
			for ok {
				select {
				case it := <-imts:
					readImageCount++
					lastTsRead = it.Ts
					ok = true

				default:
					ok = false
				}
			}
		}

		if test.fillType == "NO" {
			if !(stop.After(lastTsRead)) {
				t.Errorf("case 1: Read failed - did not read until tstop")
			} else if !(readImageCount == test.numImagesInsert) {
				t.Errorf("case 2: Read failed - #images read do not match #images inserted")
			}
		} else {

			if start.After(tsFirstOverWrittenIndex) {

				if stop.After(tsLastImage) {

					if !(readImageCount <= test.numImagesInsert-((test.segSize*test.logSize)+1)) {
						t.Errorf("case 3: Read failed - #images read do not match #images inserted")
					}
				} else {

					if !(readImageCount <= test.numImagesInsert-((test.segSize*test.logSize)+1)) {
						t.Errorf("case 4: Read failed - #images read do not match #images inserted")
					}
				}

			} else {
				if !(stop.After(lastTsRead)) {
					t.Errorf("case5: Read failed - did not read until tstop")
				} else if !(readImageCount == test.segSize*test.logSize) {
					t.Errorf("case 6: Read failed - #images read do not match #images inserted")
				}
			}

		}
	}

}

/*
Tests storage Read for already appended images for multiple readers
- for partially, fully and over written logs
- All readers with same Tstart and Tend
- for images with size 2.1M to 1K
*/
func TestMultipleReadersSameTstartAndTEnd(t *testing.T) {

	var (
		imBuf []byte
		err   error
	)
	var tests = []struct {
		imageFilesPath  string
		numImagesInsert uint64
		logSize         uint64
		segSize         uint64
		frameRate       uint64
		fillType        string
		numReader       uint64
	}{
		{"../../test_images/1000_images/", 1000, 10, 100, 33, "NO", 5},
		{"../../test_images/1000_images/", 500, 10, 100, 33, "NO", 10},
		{"../../test_images/1000_images/", 545, 10, 30, 33, "O", 10},
	}

	for _, test := range tests {
		//new memlog
		memlog := NewMemLog(test.segSize, test.logSize)

		//Obtain filenames in the directory path given
		errMsg, fileList := walkAllFilesInDir(test.imageFilesPath)
		if errMsg != "file read success" {
			t.Fatalf("File read failed - %v\n", errMsg)
		}

		start := time.Now()

		go func() {

			//read all files to buffer (conversion to bytes)
			//Append image sizes to the slice
			//Append to log
			//read images count
			ts := time.Now()
			var readImCount uint64
			for _, file := range fileList {

				if readImCount == test.numImagesInsert {
					break
				}
				imBuf, err = ioutil.ReadFile(file)

				if err != nil {
					t.Fatalf("Cannot read image file %v\n", err)

				}
				readImCount++
				ts = time.Now()

				memlog.Append(Image(imBuf), ts)
				//Specify frame rate here
				time.Sleep(time.Duration(test.frameRate) * time.Millisecond)

			}

		}()

		if test.numImagesInsert == 15 {
			time.Sleep(1000 * time.Millisecond)
		} else {
			time.Sleep(time.Duration((test.numImagesInsert*test.frameRate)+2000) * time.Millisecond)
		}

		stop := time.Now()

		var k uint64
		var imts [100]chan ImageTimestamp
		var errch [100]chan error
		done := make(chan bool)
		readErrMsg := make(chan string)
		for k = 0; k < test.numReader; k++ {

			imts[k] = make(chan ImageTimestamp, 200*1024)
			errch[k] = make(chan error)

			go memlog.Read(imts[k], start, stop, errch[k])
			go readFromChannelSameStartEnd(k, imts[k], errch[k], t, done, stop, test.fillType, test.numImagesInsert,
				test.segSize, test.logSize, readErrMsg)
			defer close(imts[k])
			defer close(errch[k])

		}

		var readStatus string
		var readStatusLock sync.Mutex
		for k = 0; k < test.numReader; k++ {

			<-done
			readStatusLock.Lock()
			readStatus = <-readErrMsg

			if readStatus != "memlog read success" {
				t.Errorf("Read failed for %v - %v\n", test.numReader, readStatus)
			}

			readStatusLock.Unlock()

		}

	}

}

/*
Tests storage Read for already appended images for multiple readers for
- for partially, fully and over written logs
- All readers with dissimilar Tstart and Tend
- for images with size 2.1M to 1K
- Test supports only numReaders<=5 readers
*/
func TestMultipleReadersDissimilarTstartAndTEnd(t *testing.T) {

	var (
		imBuf []byte
		err   error
	)
	var tests = []struct {
		imageFilesPath  string
		numImagesInsert uint64
		logSize         uint64
		segSize         uint64
		frameRate       uint64
		fillType        string
		numReader       uint64
	}{
		{"../../test_images/1000_images/", 1000, 10, 100, 33, "NO", 5},
		{"../../test_images/1000_images/", 500, 10, 100, 33, "NO", 5},
		{"../../test_images/1000_images/", 545, 10, 30, 33, "O", 5},
	}

	const startTime1 = 100
	const startTime2 = 200
	const startTime3 = 300
	const startTime4 = 400

	for _, test := range tests {
		//new memlog
		memlog := NewMemLog(test.segSize, test.logSize)

		//Obtain filenames in the directory path given
		errMsg, fileList := walkAllFilesInDir(test.imageFilesPath)
		if errMsg != "file read success" {
			t.Fatalf("File read failed - %v\n", errMsg)
		}

		var k uint64
		var start [50]time.Time
		start[0] = time.Now()
		start[1] = time.Now().Add((startTime1) * time.Millisecond)
		start[2] = time.Now().Add((startTime2) * time.Millisecond)
		start[3] = time.Now().Add((startTime3) * time.Millisecond)
		start[4] = time.Now().Add((startTime4) * time.Millisecond)

		go func() {

			//read all files to buffer (conversion to bytes)
			//Append image sizes to the slice
			//Append to log
			//read images count
			ts := time.Now()
			var readImCount uint64
			for _, file := range fileList {

				if readImCount == test.numImagesInsert {
					break
				}
				imBuf, err = ioutil.ReadFile(file)

				if err != nil {
					t.Fatalf("Cannot read image file %v\n", err)
				}
				readImCount++
				ts = time.Now()
				memlog.Append(Image(imBuf), ts)
				time.Sleep(time.Duration(test.frameRate) * time.Millisecond)
			}
		}()

		if test.numImagesInsert == 15 {
			time.Sleep(1000 * time.Millisecond)
		} else {
			time.Sleep(time.Duration((test.numImagesInsert*test.frameRate)+2000) * time.Millisecond)
		}
		stop := time.Now()

		var imts [100]chan ImageTimestamp
		var errch [100]chan error

		idcount := rIdImCount{}
		readerIdAndreadImages := make(chan rIdImCount)

		for k = 0; k < test.numReader; k++ {

			imts[k] = make(chan ImageTimestamp, 200*1024)
			errch[k] = make(chan error)

			go memlog.Read(imts[k], start[k], stop, errch[k])
			go readFromChannelDissimilarStartEnd(k, imts[k], errch[k], t, stop, test.fillType, test.numImagesInsert,
				test.segSize, test.logSize, readerIdAndreadImages)
			defer close(imts[k])
			defer close(errch[k])

		}

		readImageMap := make(map[uint64]uint64)

		var readlk sync.Mutex

		for k = 0; k < test.numReader; k++ {

			readlk.Lock()
			idcount = <-readerIdAndreadImages
			readImageMap[idcount.readerId] = idcount.readImages
			readlk.Unlock()

		}

		for k = 0; k < test.numReader-1; k++ {
			if readImageMap[k] < readImageMap[k+1] {
				t.Errorf("Test for memlog read for readers with dissimilar Tstart and Tstop failed")
			}
		}
	}

}

/*
Tests storage Read concurrently with Append for images
- for partially, fully and over written logs for a single reader
- for image size from 2.1M to 1K
*/
func TestReadConcurrent(t *testing.T) {
	var (
		imBuf []byte
		err   error
	)
	var tests = []struct {
		imageFilesPath  string
		numImagesInsert uint64
		logSize         uint64
		segSize         uint64
		frameRate       uint64
		fillType        string
	}{
		{"../../test_images/1000_images/", 1000, 10, 100, 33, "NO"},
		{"../../test_images/1000_images/", 500, 10, 100, 33, "NO"},
		{"../../test_images/1000_images/", 545, 10, 30, 33, "O"},
	}

	//sleep time should be less than NO_IM*f0
	const sleepTime = 70
	//waitTime should be >f0*segsize*logsize
	const waitTime = 10000

	for _, test := range tests {
		//new memlog
		memlog := NewMemLog(test.segSize, test.logSize)

		//Obtain filenames in the directory path given
		errMsg, fileList := walkAllFilesInDir(test.imageFilesPath)
		if errMsg != "file read success" {
			t.Fatalf("File read failed - %v\n", errMsg)
		}

		start := time.Now()

		go func() {

			//read all files to buffer (conversion to bytes)
			//Append image sizes to the slice
			//Append to log
			//read images count
			ts := time.Now()
			var readImCount uint64
			for _, file := range fileList {

				if readImCount == test.numImagesInsert {
					break
				}
				imBuf, err = ioutil.ReadFile(file)

				if err != nil {
					t.Fatalf("Cannot read image file %v\n", err)
				}
				readImCount++
				ts = time.Now()

				memlog.Append(Image(imBuf), ts)

				time.Sleep(time.Duration(test.frameRate) * time.Millisecond)

			}

		}()

		stop := time.Now()
		if test.fillType == "NO" {
			time.Sleep(sleepTime * time.Millisecond)
			stop = start.Add(time.Millisecond * (sleepTime))
		} else {
			time.Sleep((waitTime + sleepTime) * time.Millisecond)
			stop = start.Add(time.Millisecond * (waitTime + sleepTime))
		}

		imts := make(chan ImageTimestamp, 200*1024)
		errch := make(chan error)
		defer close(imts)
		defer close(errch)

		go memlog.Read(imts, start, stop, errch)

		errc := <-errch

		//To count total number of images read
		var readImageCount uint64

		if errc != nil {
			t.Fatalf("Memlog Read internal error - %s", errc)
		} else {
			ok := true
			for ok {
				select {
				case _ = <-imts:
					readImageCount++
					ok = true

				default:
					ok = false
				}
			}
		}

		if test.fillType == "NO" {
			if !(readImageCount == ((sleepTime / test.frameRate) + 1)) {
				t.Errorf("Memlog Read failed - #images read do not match #images inserted")
			}
		} else {
			if !(readImageCount == test.segSize*test.logSize) {
				t.Errorf("Memlog Read failed - #images read do not match #images inserted")
			}
		}
	}

}

/*
Tests storage Read concurrently with appended images for multiple readers
- for partially, fully and over written logs
- All readers with same Tstart and Tend
- image size from 2.1M to 1K
*/
func TestMultipleConcurrentReadersSameTstartAndTEnd(t *testing.T) {

	var (
		imBuf []byte
		err   error
	)
	var tests = []struct {
		imageFilesPath  string
		numImagesInsert uint64
		logSize         uint64
		segSize         uint64
		frameRate       uint64
		fillType        string
		numReader       uint64
	}{
		{"../../test_images/1000_images/", 1000, 10, 100, 33, "NO", 5},
		{"../../test_images/1000_images/", 500, 10, 100, 33, "NO", 10},
		{"../../test_images/1000_images/", 545, 10, 30, 33, "O", 10},
	}

	//sleep time should be less than NO_IM*f0
	const sleepTime = 70
	//waitTime should be >f0*segsize*logsize
	const waitTime = 10000

	for _, test := range tests {
		//new memlog
		memlog := NewMemLog(test.segSize, test.logSize)

		//Obtain filenames in the directory path given
		errMsg, fileList := walkAllFilesInDir(test.imageFilesPath)
		if errMsg != "file read success" {
			t.Fatalf("File read failed - %v\n", errMsg)
		}

		start := time.Now()

		go func() {

			//read all files to buffer (conversion to bytes)
			//Append image sizes to the slice
			//Append to log
			//read images count
			ts := time.Now()
			var readImCount uint64
			for _, file := range fileList {

				if readImCount == test.numImagesInsert {
					break
				}
				imBuf, err = ioutil.ReadFile(file)

				if err != nil {
					t.Fatalf("Cannot read image file %v\n", err)

				}
				readImCount++
				ts = time.Now()

				memlog.Append(Image(imBuf), ts)

				time.Sleep(time.Duration(test.frameRate) * time.Millisecond)

			}

		}()

		stop := time.Now()
		if test.fillType == "NO" {
			time.Sleep(sleepTime * time.Millisecond)
			stop = start.Add(time.Millisecond * (sleepTime))
		} else {
			time.Sleep((waitTime + sleepTime) * time.Millisecond)
			stop = start.Add(time.Millisecond * (waitTime + sleepTime))
		}

		var k uint64
		var imts [100]chan ImageTimestamp
		var errch [100]chan error
		readIms := make(chan uint64)
		for k = 0; k < test.numReader; k++ {

			imts[k] = make(chan ImageTimestamp, 200*1024)
			errch[k] = make(chan error)

			go memlog.Read(imts[k], start, stop, errch[k])
			go readFromChannelConcurrent(k, imts[k], errch[k], t, readIms, stop, test.fillType,
				test.numImagesInsert, test.segSize, test.logSize)
			defer close(imts[k])
			defer close(errch[k])

		}

		var readlk sync.Mutex

		for k = 0; k < test.numReader; k++ {

			readlk.Lock()
			countIms := <-readIms
			if test.fillType == "NO" {
				if !(countIms == ((sleepTime / test.frameRate) + 1)) {
					t.Errorf("Memlog Read failed - #images read do not match #images inserted")
				}
			} else {
				if !(countIms == test.segSize*test.logSize) {
					t.Errorf("Memlog Read failed - #images read do not match #images inserted")
				}
			}
			readlk.Unlock()

		}
	}

}

//*****************************Storage tests end***************************************
