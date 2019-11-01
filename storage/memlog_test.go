package storage

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"
)

//***************************Helper functions start*****************************************

/*To return list of absolute paths of files in a directory
Input - directory path
Output- list of filepaths
*/
func walkAllFilesInDir(dir string) []string {
	fileList := make([]string, 0)
	_ = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// check if it is a regular file (not dir)
		if info.Mode().IsRegular() {
			fileList = append(fileList, path)

		}
		return err
	})

	return fileList
}

/*To compare equality of timestamp slices (Appended and Read)
and to compare equality of image size slices (Appended and Read)
Input - Appended and Read slices
Output- true or false boolean
*/
func sliceEquality(tsAppended, tsRead []time.Time, imSizeAppended, imSizeRead []int) bool {
	if len(tsAppended) != len(tsRead) {
		log.Fatalln("timestamp logs length mismatch")
		return false

	}

	if len(imSizeAppended) != len(imSizeRead) {
		log.Fatalln("image logs length mismatch")
		return false

	}

	for i := 0; i < len(tsAppended); i++ {
		if tsAppended[i] != tsRead[i] {
			log.Fatalln("timestamp logs contents mismatch")
			return false

		}
		if imSizeAppended[i] != imSizeRead[i] {
			log.Fatalln("image logs contents mismatch")
			return false
		}
	}
	return true
}

//**********************************Helper functions end***********************************

//***********************************Storage tests start************************************
/*
First test - tests storage Append and Read (For reference)
*/
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
Tests storage capability to store same sized images at 30fps rate
Tests only storage Append
Test for partially written, fully written logs
To run this test for fully written case use command,
> FP=/path/to/images/ NO_IM=1000 LS=10 SS=100 go test -v -run="TestAppendSameSizeImage"
To run this test for partially written case use command,
> FP=/path/to/images/ NO_IM=500 LS=10 SS=100 go test -v -run="TestAppendSameSizeImage"
FP is path to images, NO_IM is no of images to be inserted
LS is log size and SS is segment size
Frame rate should be specified in code
*/
func TestAppendSameSizeImage(t *testing.T) {

	const f0 = 33

	imageFilesPath := os.Getenv("FP")
	noImages := os.Getenv("NO_IM")
	noImagesInsert, _ := strconv.Atoi(noImages)
	logSz := os.Getenv("LS")
	logSize, _ := strconv.ParseUint(logSz, 10, 64)
	segSz := os.Getenv("SS")
	segSize, _ := strconv.ParseUint(segSz, 10, 64)

	//frame rate should be specified after inserting the timestamp
	var (
		imBuf []byte
		err   error
	)

	//new memlog
	memlog := NewMemLog(segSize, logSize)

	//Obtain filenames in the directory path given
	fileList := walkAllFilesInDir(imageFilesPath)

	//read first file to buffer (conversion to bytes)
	imBuf, err = ioutil.ReadFile(fileList[0])
	if err != nil {
		t.Errorf("Cannot read image file %v\n", err)
		panic(err)

	}

	//slices to store timestamps and image sizes appended
	tsAppended := make([]time.Time, 0)
	imSizeAppended := make([]int, 0)

	//Append image sizes to the slice
	for i := 0; i < noImagesInsert; i++ {
		imSizeAppended = append(imSizeAppended, len(imBuf))
	}

	// Append to log
	for i := 0; i < noImagesInsert; i++ {
		ts := time.Now()
		tsAppended = append(tsAppended, ts)
		memlog.Append(Image(imBuf), ts)
		//Specify frame rate here
		time.Sleep(f0 * time.Millisecond)
		//sz, pos, tins := memlog.AppendStats()
		//imSizeAppended = append(imSizeAppended, sz)
		//t.Logf("Image of size %d inserted in log at position %d at time %v\n", sz, pos, tins)

	}

	//verifying the timestamps and image sizes directly from log
	//count variable to keep track of number of elements in log
	var count int
	//slices to read timestamps and image sizes
	tsRead := make([]time.Time, 0)
	imSizeRead := make([]int, 0)
	var i uint64
	var j uint64
	for i = 0; i < logSize; i++ {
		for j = 0; j < segSize; j++ {
			if count == noImagesInsert {
				break
			}

			tsRead = append(tsRead, memlog.tsmemlog.tslog[i].ts[j])
			imSizeRead = append(imSizeRead, len(memlog.immemlog.imlog[i].im[j]))
			count = count + 1
		}

	}

	if !(sliceEquality(tsAppended, tsRead, imSizeAppended, imSizeRead)) {
		t.Errorf("Memlog append failed")
	}

}

/*
Tests storage capability to store same sized images at 30fps rate
Tests only storage Append
Test for over written logs
To run this test use command,
> FP=/path/to/images/ NO_IM=545 LS=10 SS=50 go test -v -run="TestAppendOverWrittenLog"
FP is path to images, NO_IM is no of images to be inserted
LS is log size and SS is segment size
Frame rate should be specified in code
*/

func TestAppendOverWrittenLog(t *testing.T) {

	const f0 = 33

	imageFilesPath := os.Getenv("FP")
	noImages := os.Getenv("NO_IM")
	noImagesInsert, _ := strconv.ParseUint(noImages, 10, 64)
	logSz := os.Getenv("LS")
	logSize, _ := strconv.ParseUint(logSz, 10, 64)
	segSz := os.Getenv("SS")
	segSize, _ := strconv.ParseUint(segSz, 10, 64)

	//frame rate should be specified after inserting the timestamp
	var (
		imBuf []byte
		err   error
	)

	//new memlog
	memlog := NewMemLog(segSize, logSize)

	//Obtain filenames in the directory path given
	fileList := walkAllFilesInDir(imageFilesPath)

	//read first file to buffer (conversion to bytes)
	imBuf, err = ioutil.ReadFile(fileList[0])
	if err != nil {
		t.Errorf("Cannot read image file %v\n", err)
		panic(err)

	}

	//slices to store timestamps and image sizes appended
	tsAppended := make([]time.Time, 0)
	imSizeAppended := make([]int, 0)

	//Append image sizes to the slice
	var i uint64
	for i = segSize * logSize; i < noImagesInsert; i++ {
		imSizeAppended = append(imSizeAppended, len(imBuf))
	}

	// Append to log
	for i = 0; i < noImagesInsert; i++ {
		ts := time.Now()

		if i >= segSize*logSize {
			tsAppended = append(tsAppended, ts)
		}

		memlog.Append(Image(imBuf), ts)
		//Specify frame rate here
		time.Sleep(f0 * time.Millisecond)
		//sz, pos, tins := memlog.AppendStats()
		//imSizeAppended = append(imSizeAppended, sz)
		//t.Logf("Image of size %d inserted in log at position %d at time %v\n", sz, pos, tins)

	}

	//verifying the timestamps and image sizes directly from log
	//count variable to keep track of number of elements in log
	var count uint64
	//slices to read timestamps and image sizes
	tsRead := make([]time.Time, 0)
	imSizeRead := make([]int, 0)
	var j uint64
	for i = 0; i < logSize; i++ {
		for j = 0; j < segSize; j++ {
			if count == (noImagesInsert - (segSize * logSize)) {
				break
			}

			tsRead = append(tsRead, memlog.tsmemlog.tslog[i].ts[j])
			imSizeRead = append(imSizeRead, len(memlog.immemlog.imlog[i].im[j]))
			count = count + 1
		}

	}

	if !(sliceEquality(tsAppended, tsRead, imSizeAppended, imSizeRead)) {
		t.Errorf("Memlog append failed")
	}

}

/*
Tests storage capability to store same sized images at 24fps rate
Tests only storage Append
Test for partially written, fully written logs
To run this test for fully written case use command,
> FP=/path/to/images/ NO_IM=1000 LS=10 SS=100 go test -v -run="TestAppendSameSizeImage"
To run this test for partially written case use command,
> FP=/path/to/images/ NO_IM=500 LS=10 SS=100 go test -v -run="TestAppendSameSizeImage"
FP is path to images, NO_IM is no of images to be inserted
LS is log size and SS is segment size
Frame rate should be specified in code
*/
func TestAppendSameSizeImageFR24(t *testing.T) {

	const f1 = 42

	imageFilesPath := os.Getenv("FP")
	noImages := os.Getenv("NO_IM")
	noImagesInsert, _ := strconv.Atoi(noImages)
	logSz := os.Getenv("LS")
	logSize, _ := strconv.ParseUint(logSz, 10, 64)
	segSz := os.Getenv("SS")
	segSize, _ := strconv.ParseUint(segSz, 10, 64)

	//frame rate should be specified after inserting the timestamp
	var (
		imBuf []byte
		err   error
	)

	//new memlog
	memlog := NewMemLog(segSize, logSize)

	//Obtain filenames in the directory path given
	fileList := walkAllFilesInDir(imageFilesPath)

	//read first file to buffer (conversion to bytes)
	imBuf, err = ioutil.ReadFile(fileList[0])
	if err != nil {
		t.Errorf("Cannot read image file %v\n", err)
		panic(err)

	}

	//slices to store timestamps and image sizes appended
	tsAppended := make([]time.Time, 0)
	imSizeAppended := make([]int, 0)

	//Append image sizes to the slice
	for i := 0; i < noImagesInsert; i++ {
		imSizeAppended = append(imSizeAppended, len(imBuf))
	}

	// Append to log at 24fps rate
	for i := 0; i < noImagesInsert; i++ {
		ts := time.Now()
		tsAppended = append(tsAppended, ts)
		memlog.Append(Image(imBuf), ts)
		//Specify frame rate here
		time.Sleep(f1 * time.Millisecond)
		//sz, pos, tins := memlog.AppendStats()
		//imSizeAppended = append(imSizeAppended, sz)
		//t.Logf("Image of size %d inserted in log at position %d at time %v\n", sz, pos, tins)

	}

	//verifying the timestamps and image sizes directly from log
	//count variable to keep track of number of elements in log
	var count int
	//slices to read timestamps and image sizes
	tsRead := make([]time.Time, 0)
	imSizeRead := make([]int, 0)
	var i uint64
	var j uint64
	for i = 0; i < logSize; i++ {
		for j = 0; j < segSize; j++ {
			if count == noImagesInsert {
				break
			}

			tsRead = append(tsRead, memlog.tsmemlog.tslog[i].ts[j])
			imSizeRead = append(imSizeRead, len(memlog.immemlog.imlog[i].im[j]))
			count = count + 1
		}

	}

	if !(sliceEquality(tsAppended, tsRead, imSizeAppended, imSizeRead)) {
		t.Errorf("Memlog append failed")
	}

}

/*
Tests storage capability to store same sized images at 24fps rate
Tests only storage Append
Test for over written logs
To run this test use command,
> FP=/path/to/images/ NO_IM=545 LS=10 SS=50 go test -v -run="TestAppendOverWrittenLog"
FP is path to images, NO_IM is no of images to be inserted
LS is log size and SS is segment size
Frame rate should be specified in code
*/

func TestAppendOverWrittenLogFR24(t *testing.T) {

	const f1 = 42

	imageFilesPath := os.Getenv("FP")
	noImages := os.Getenv("NO_IM")
	noImagesInsert, _ := strconv.ParseUint(noImages, 10, 64)
	logSz := os.Getenv("LS")
	logSize, _ := strconv.ParseUint(logSz, 10, 64)
	segSz := os.Getenv("SS")
	segSize, _ := strconv.ParseUint(segSz, 10, 64)

	//frame rate should be specified after inserting the timestamp
	var (
		imBuf []byte
		err   error
	)

	//new memlog
	memlog := NewMemLog(segSize, logSize)

	//Obtain filenames in the directory path given
	fileList := walkAllFilesInDir(imageFilesPath)

	//read first file to buffer (conversion to bytes)
	imBuf, err = ioutil.ReadFile(fileList[0])
	if err != nil {
		t.Errorf("Cannot read image file %v\n", err)
		panic(err)

	}

	//slices to store timestamps and image sizes appended
	tsAppended := make([]time.Time, 0)
	imSizeAppended := make([]int, 0)

	//Append image sizes to the slice
	var i uint64
	for i = segSize * logSize; i < noImagesInsert; i++ {
		imSizeAppended = append(imSizeAppended, len(imBuf))
	}

	// Append to log
	for i = 0; i < noImagesInsert; i++ {
		ts := time.Now()

		if i >= segSize*logSize {
			tsAppended = append(tsAppended, ts)
		}

		memlog.Append(Image(imBuf), ts)
		//Specify frame rate here
		time.Sleep(f1 * time.Millisecond)
		//sz, pos, tins := memlog.AppendStats()
		//imSizeAppended = append(imSizeAppended, sz)
		//t.Logf("Image of size %d inserted in log at position %d at time %v\n", sz, pos, tins)

	}

	//verifying the timestamps and image sizes directly from log
	//count variable to keep track of number of elements in log
	var count uint64
	//slices to read timestamps and image sizes
	tsRead := make([]time.Time, 0)
	imSizeRead := make([]int, 0)
	var j uint64
	for i = 0; i < logSize; i++ {
		for j = 0; j < segSize; j++ {
			if count == (noImagesInsert - (segSize * logSize)) {
				break
			}

			tsRead = append(tsRead, memlog.tsmemlog.tslog[i].ts[j])
			imSizeRead = append(imSizeRead, len(memlog.immemlog.imlog[i].im[j]))
			count = count + 1
		}

	}

	if !(sliceEquality(tsAppended, tsRead, imSizeAppended, imSizeRead)) {
		t.Errorf("Memlog append failed")
	}

}

/*
Tests storage capability to store same sized images at 15fps rate
Tests only storage Append
Test for partially written, fully written logs
To run this test for fully written case use command,
> FP=/path/to/images/ NO_IM=1000 LS=10 SS=100 go test -v -run="TestAppendSameSizeImage"
To run this test for partially written case use command,
> FP=/path/to/images/ NO_IM=500 LS=10 SS=100 go test -v -run="TestAppendSameSizeImage"
FP is path to images, NO_IM is no of images to be inserted
LS is log size and SS is segment size
Frame rate should be specified in code
*/
func TestAppendSameSizeImageFR15(t *testing.T) {

	const f2 = 67

	imageFilesPath := os.Getenv("FP")
	noImages := os.Getenv("NO_IM")
	noImagesInsert, _ := strconv.Atoi(noImages)
	logSz := os.Getenv("LS")
	logSize, _ := strconv.ParseUint(logSz, 10, 64)
	segSz := os.Getenv("SS")
	segSize, _ := strconv.ParseUint(segSz, 10, 64)

	//frame rate should be specified after inserting the timestamp
	var (
		imBuf []byte
		err   error
	)

	//new memlog
	memlog := NewMemLog(segSize, logSize)

	//Obtain filenames in the directory path given
	fileList := walkAllFilesInDir(imageFilesPath)

	//read first file to buffer (conversion to bytes)
	imBuf, err = ioutil.ReadFile(fileList[0])
	if err != nil {
		t.Errorf("Cannot read image file %v\n", err)
		panic(err)

	}

	//slices to store timestamps and image sizes appended
	tsAppended := make([]time.Time, 0)
	imSizeAppended := make([]int, 0)

	//Append image sizes to the slice
	for i := 0; i < noImagesInsert; i++ {
		imSizeAppended = append(imSizeAppended, len(imBuf))
	}

	// Append to log at 24fps rate
	for i := 0; i < noImagesInsert; i++ {
		ts := time.Now()
		tsAppended = append(tsAppended, ts)
		memlog.Append(Image(imBuf), ts)
		//Specify frame rate here
		time.Sleep(f2 * time.Millisecond)
		//sz, pos, tins := memlog.AppendStats()
		//imSizeAppended = append(imSizeAppended, sz)
		//t.Logf("Image of size %d inserted in log at position %d at time %v\n", sz, pos, tins)

	}

	//verifying the timestamps and image sizes directly from log
	//count variable to keep track of number of elements in log
	var count int
	//slices to read timestamps and image sizes
	tsRead := make([]time.Time, 0)
	imSizeRead := make([]int, 0)
	var i uint64
	var j uint64
	for i = 0; i < logSize; i++ {
		for j = 0; j < segSize; j++ {
			if count == noImagesInsert {
				break
			}

			tsRead = append(tsRead, memlog.tsmemlog.tslog[i].ts[j])
			imSizeRead = append(imSizeRead, len(memlog.immemlog.imlog[i].im[j]))
			count = count + 1
		}

	}

	if !(sliceEquality(tsAppended, tsRead, imSizeAppended, imSizeRead)) {
		t.Errorf("Memlog append failed")
	}

}

/*
Tests storage capability to store same sized images at 15fps rate
Tests only storage Append
Test for over written logs
To run this test use command,
> FP=/path/to/images/ NO_IM=545 LS=10 SS=50 go test -v -run="TestAppendOverWrittenLog"
FP is path to images, NO_IM is no of images to be inserted
LS is log size and SS is segment size
Frame rate should be specified in code
*/

func TestAppendOverWrittenLogFR15(t *testing.T) {

	const f2 = 67

	imageFilesPath := os.Getenv("FP")
	noImages := os.Getenv("NO_IM")
	noImagesInsert, _ := strconv.ParseUint(noImages, 10, 64)
	logSz := os.Getenv("LS")
	logSize, _ := strconv.ParseUint(logSz, 10, 64)
	segSz := os.Getenv("SS")
	segSize, _ := strconv.ParseUint(segSz, 10, 64)

	//frame rate should be specified after inserting the timestamp
	var (
		imBuf []byte
		err   error
	)

	//new memlog
	memlog := NewMemLog(segSize, logSize)

	//Obtain filenames in the directory path given
	fileList := walkAllFilesInDir(imageFilesPath)

	//read first file to buffer (conversion to bytes)
	imBuf, err = ioutil.ReadFile(fileList[0])
	if err != nil {
		t.Errorf("Cannot read image file %v\n", err)
		panic(err)

	}

	//slices to store timestamps and image sizes appended
	tsAppended := make([]time.Time, 0)
	imSizeAppended := make([]int, 0)

	//Append image sizes to the slice
	var i uint64
	for i = segSize * logSize; i < noImagesInsert; i++ {
		imSizeAppended = append(imSizeAppended, len(imBuf))
	}

	// Append to log
	for i = 0; i < noImagesInsert; i++ {
		ts := time.Now()

		if i >= segSize*logSize {
			tsAppended = append(tsAppended, ts)
		}

		memlog.Append(Image(imBuf), ts)
		//Specify frame rate here
		time.Sleep(f2 * time.Millisecond)
		//sz, pos, tins := memlog.AppendStats()
		//imSizeAppended = append(imSizeAppended, sz)
		//t.Logf("Image of size %d inserted in log at position %d at time %v\n", sz, pos, tins)

	}

	//verifying the timestamps and image sizes directly from log
	//count variable to keep track of number of elements in log
	var count uint64
	//slices to read timestamps and image sizes
	tsRead := make([]time.Time, 0)
	imSizeRead := make([]int, 0)
	var j uint64
	for i = 0; i < logSize; i++ {
		for j = 0; j < segSize; j++ {
			if count == (noImagesInsert - (segSize * logSize)) {
				break
			}

			tsRead = append(tsRead, memlog.tsmemlog.tslog[i].ts[j])
			imSizeRead = append(imSizeRead, len(memlog.immemlog.imlog[i].im[j]))
			count = count + 1
		}

	}

	if !(sliceEquality(tsAppended, tsRead, imSizeAppended, imSizeRead)) {
		t.Errorf("Memlog append failed")
	}

}

/*
Tests storage capability to store variable sized images at 30fps rate
Tests only storage Append
Test for partially written, fully written logs
To run this test for fully written case use command,
> FP=/path/to/images/ NO_IM=1000 LS=10 SS=100 go test -v -run="TestAppendSameSizeImage"
To run this test for partially written case use command,
> FP=/path/to/images/ NO_IM=500 LS=10 SS=100 go test -v -run="TestAppendSameSizeImage"
FP is path to images, NO_IM is no of images to be inserted
LS is log size and SS is segment size
Frame rate should be specified in code
*/
func TestAppendVariableSizeImage(t *testing.T) {

	const f0 = 33

	imageFilesPath := os.Getenv("FP")
	noImages := os.Getenv("NO_IM")
	noImagesInsert, _ := strconv.ParseUint(noImages, 10, 64)
	logSz := os.Getenv("LS")
	logSize, _ := strconv.ParseUint(logSz, 10, 64)
	segSz := os.Getenv("SS")
	segSize, _ := strconv.ParseUint(segSz, 10, 64)

	//frame rate should be specified after inserting the timestamp
	var (
		imBuf []byte
		err   error
	)

	//new memlog
	memlog := NewMemLog(segSize, logSize)

	//Obtain filenames in the directory path given
	fileList := walkAllFilesInDir(imageFilesPath)

	//slices to store timestamps and image sizes appended
	tsAppended := make([]time.Time, 0)
	imSizeAppended := make([]int, 0)

	//read all files to buffer (conversion to bytes)
	//Append image sizes to the slice
	// Append to log
	//read images count
	var readImCount uint64
	var i uint64
	for _, file := range fileList {
		if readImCount == noImagesInsert {
			break
		}
		imBuf, err = ioutil.ReadFile(file)
		imSizeAppended = append(imSizeAppended, len(imBuf))
		if err != nil {
			t.Errorf("Cannot read image file %v\n", err)
			panic(err)

		}
		readImCount++

		ts := time.Now()
		tsAppended = append(tsAppended, ts)
		memlog.Append(Image(imBuf), ts)
		//Specify frame rate here
		time.Sleep(f0 * time.Millisecond)
		//sz, pos, tins := memlog.AppendStats()
		//imSizeAppended = append(imSizeAppended, sz)
		//t.Logf("Image of size %d inserted in log at position %d at time %v\n", sz, pos, tins)

	}

	//verifying the timestamps and image sizes directly from log
	//count variable to keep track of number of elements in log
	var count uint64
	//slices to read timestamps and image sizes
	tsRead := make([]time.Time, 0)
	imSizeRead := make([]int, 0)
	var j uint64
	for i = 0; i < logSize; i++ {
		for j = 0; j < segSize; j++ {
			if count == noImagesInsert {
				break
			}

			tsRead = append(tsRead, memlog.tsmemlog.tslog[i].ts[j])
			imSizeRead = append(imSizeRead, len(memlog.immemlog.imlog[i].im[j]))
			count = count + 1
		}

	}

	if !(sliceEquality(tsAppended, tsRead, imSizeAppended, imSizeRead)) {
		t.Errorf("Memlog append failed")
	}

}

//*****************************Storage tests end***************************************
