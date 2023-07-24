package pkg

import (
	"fmt"
	"os"
	"runtime"
)

// this file creates a dynamic directory within the
// /var/lib/rancher/gmsa directory. The name of this
// directory needs to be the same as the release namespace
// of the chart this operator is packaged in.

const baseDir = "/var/lib/rancher/gmsa"

func CreateDir(dirName string) error {
	if runtime.GOOS != "windows" {
		// this program should
		// not run on linux
		// or mac
		return nil
	}

	// TODO: Adjust Directory Permissions
	if _, err := os.Stat(baseDir); os.IsNotExist(err) {
		err = os.Mkdir(baseDir, os.ModePerm)
	}

	if _, err := os.Stat(fmt.Sprintf("%s/%s", baseDir, dirName)); os.IsNotExist(err) {
		err = os.Mkdir(fmt.Sprintf("%s/%s", baseDir, dirName), os.ModePerm)
	}

	return nil
}

func WritePortFile(port string) error {
	if runtime.GOOS != "windows" {
		// this program should
		// not run on linux
		// or mac
		return nil
	}

	// TODO: adjust file permissions
	if _, err := os.Stat(fmt.Sprintf("%s/%s", baseDir, "port.txt")); os.IsNotExist(err) {
		// create the file
		err = os.WriteFile(fmt.Sprintf("%s/%s", baseDir, "port.txt"), []byte(port), os.ModePerm)
	}

	// verify the files contents

	return nil
}
