package storage

import (
	"log"
	"os"
	"testing"
)

//Back up files should be present to run this test

func TestRecoverSameSizeImage(t *testing.T) {

	var tests = []struct {
		imageFilesPath  string
		numImagesInsert uint64
		logSize         uint64
		segSize         uint64
		frameRate       int16
		recoveryAddr    string
		camid           string
	}{
		{"../../test_images/10K/", 100, 5, 10, 200, "../broker/enb.txt", "nil"},
	}

	for _, test := range tests {
		//new memlog
		memlog := NewMemLog(test.segSize, test.logSize)

		recoveryFile, err := os.Open(test.recoveryAddr)
		if err != nil {
			log.Fatalln("cannot open recovery file", err)
		}

		memlog.Recover(recoveryFile, test.camid)

	}

}
