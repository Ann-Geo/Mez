package storage

import (
	"bufio"
	"log"
	"os"
	"path/filepath"
)

func (memlog *MemLog) Recover(recoveryFile *os.File) {

	var recoveryPath string

	//Get the recoveryPath frrom recovery file
	scanner := bufio.NewScanner(recoveryFile)
	for scanner.Scan() {
		recoveryPath = scanner.Text()
		break
	}

	//Obtain filenames in the recovery path given
	errMsg, fileList := walkFilesInDir(recoveryPath)
	if errMsg != "file read success" {
		log.Fatalln("File read failed", errMsg)
	}

	for _, _ = range fileList {

	}

}

func walkFilesInDir(dir string) (string, []string) {
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
