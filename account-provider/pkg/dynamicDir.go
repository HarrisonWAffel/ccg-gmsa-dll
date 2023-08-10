package pkg

import (
	"fmt"
	"os"
	"runtime"
)

const gmsaDirectory = "/var/lib/rancher/gmsa"

func CreateDir(dirName string) error {
	if runtime.GOOS != "windows" {
		// this program should
		// not run on linux
		// or mac
		return fmt.Errorf("unsupported OS")
	}

	// TODO: Adjust Directory Permissions
	if _, err := os.Stat(gmsaDirectory); os.IsNotExist(err) {
		err = os.Mkdir(gmsaDirectory, os.ModePerm)
		if err != nil {
			return fmt.Errorf("failed to create base directory: %v", err)
		}
	}

	if _, err := os.Stat(fmt.Sprintf("%s/%s", gmsaDirectory, dirName)); os.IsNotExist(err) {
		err = os.Mkdir(fmt.Sprintf("%s/%s", gmsaDirectory, dirName), os.ModePerm)
		if err != nil {
			return fmt.Errorf("failed to create dynamic sub directory: %v", err)
		}
	}

	return nil
}

func WritePortFile(dirName, port string) error {
	portFile := fmt.Sprintf("%s/%s/%s", gmsaDirectory, dirName, "port.txt")
	// TODO: adjust file permissions
	if _, err := os.Stat(portFile); os.IsNotExist(err) {
		// create the port file
		err = os.WriteFile(portFile, []byte(port), os.ModePerm)
		if err != nil {
			return fmt.Errorf("failed to create port.txt: %v", err)
		}
	}

	// update file with new port
	f, err := os.OpenFile(portFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return fmt.Errorf("failed to open port mTLSFile: %v", err)
	}

	_, err = f.WriteString(port)
	if err != nil {
		return fmt.Errorf("failed to update port mTLSFile: %v", err)
	}

	return f.Close()
}

func HardenDir() error {

	// todo; setup proper ACE / ACL for files and folders

	return nil
}
