package pkg

import (
	"fmt"
	"os"
	"runtime"

	"github.com/pkg/errors"
)

// this file creates a dynamic directory within the
// /var/lib/rancher/gmsa directory. The name of this
// directory needs to be the same as the release namespace
// of the chart this operator is packaged in.

const baseDir = "/var/lib/rancher/gmsa"

func CreateDir(dirName, port string) error {
	if runtime.GOOS != "windows" {
		// this program should
		// not run on linux
		// or mac
		return fmt.Errorf("unsupported OS")
	}

	// TODO: Adjust Directory Permissions
	if _, err := os.Stat(baseDir); os.IsNotExist(err) {
		err = os.Mkdir(baseDir, os.ModePerm)
		if err != nil {
			return fmt.Errorf("failed to create base directory: %v", err)
		}
	}

	if _, err := os.Stat(fmt.Sprintf("%s/%s", baseDir, dirName)); os.IsNotExist(err) {
		err = os.Mkdir(fmt.Sprintf("%s/%s", baseDir, dirName), os.ModePerm)
		if err != nil {
			return fmt.Errorf("failed to create dynamic sub directory: %v", err)
		}
	}

	portFile := fmt.Sprintf("%s/%s/%s", baseDir, dirName, "port.txt")
	// TODO: adjust file permissions
	if _, err := os.Stat(portFile); os.IsNotExist(err) {
		// create the file
		err = os.WriteFile(portFile, []byte(port), os.ModePerm)
		if err != nil {
			return fmt.Errorf("failed to create port.txt: %v", err)
		}
	}

	contents, err := os.ReadFile(portFile)
	if err != nil {
		return fmt.Errorf("failed to read the contents of %s: %v", portFile, err)
	}

	if string(contents) == port {
		return nil
	}

	// update file with new port
	f, err := os.OpenFile(portFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return errors.Wrap(err, "failed to open port.txt file")
	}

	_, err = f.WriteString(port)
	if err != nil {
		return errors.Wrap(err, "failed to update port.txt file")
	}

	return f.Close()
}
