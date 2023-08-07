package pkg

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

const (
	baseDir = "/var/lib/rancher/gmsa"

	serverCa  = "%s/%s/container/ssl/server/ca.crt"
	serverCrt = "%s/%s/container/ssl/server/tls.crt"
	serverKey = "%s/%s/container/ssl/server/tls.key"

	containerClientDir = "%s/%s/container/ssl/client"
	containerClientCa  = containerClientDir + "/ca.crt"
	containerClientCrt = containerClientDir + "/tls.crt"
	containerClientKey = containerClientDir + "/tls.key"

	containerRootCaDir = containerClientDir + "/ca"
	containerRootCa    = containerRootCaDir + "/ca.crt"

	hostSslDir = "%s/%s/ssl"

	hostCaDir  = hostSslDir + "/ca"
	hostRootCa = hostCaDir + "/ca.crt"

	hostClientDir = hostSslDir + "/client"
	hostClientCa  = hostClientDir + "/ca.crt"
	hostClientCrt = hostClientDir + "/tls.crt"
	hostClientKey = hostClientDir + "/tls.key"

	hostServerDir = hostSslDir + "/server"
	hostServerCa  = hostServerDir + "/ca.crt"
	hostServerCrt = hostServerDir + "/tls.crt"
	hostServerKey = hostServerDir + "/tls.key"

	fingerPrintDir = hostSslDir + "/fingerprint.txt"
)

func CreateDir(dirName string) error {
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

	return nil
}

func WritePortFile(dirName, port string) error {
	portFile := fmt.Sprintf("%s/%s/%s", baseDir, dirName, "port.txt")
	// TODO: adjust file permissions
	if _, err := os.Stat(portFile); os.IsNotExist(err) {
		// create the file
		err = os.WriteFile(portFile, []byte(port), os.ModePerm)
		if err != nil {
			return fmt.Errorf("failed to create port.txt: %v", err)
		}
	}

	// update file with new port
	f, err := os.OpenFile(portFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return fmt.Errorf("failed to open port file: %v", err)
	}

	_, err = f.WriteString(port)
	if err != nil {
		return fmt.Errorf("failed to update port file: %v", err)
	}

	return f.Close()
}

// WriteClientCerts copies the client tls certificate and key from the container
// filesystem to the host filesystem so that it may be used by the CCG plugin
func WriteClientCerts(dirName string) error {
	containerCrt := fmt.Sprintf(containerClientCrt, baseDir, dirName)
	containerKey := fmt.Sprintf(containerClientKey, baseDir, dirName)
	containerCa := fmt.Sprintf(containerClientCa, baseDir, dirName)

	err := os.Mkdir(fmt.Sprintf(hostSslDir, baseDir, dirName), os.ModePerm)
	if err != nil && !strings.Contains(err.Error(), "already exists") {
		return fmt.Errorf("failed to create client certificate directory: %v", err)
	}

	err = os.Mkdir(fmt.Sprintf(hostClientDir, baseDir, dirName), os.ModePerm)
	if err != nil && !strings.Contains(err.Error(), "already exists") {
		return fmt.Errorf("failed to create client certificate directory: %v", err)
	}

	err = os.Mkdir(fmt.Sprintf(hostServerDir, baseDir, dirName), os.ModePerm)
	if err != nil && !strings.Contains(err.Error(), "already exists") {
		return fmt.Errorf("failed to create client certificate directory: %v", err)
	}

	certBytes, err := os.ReadFile(containerCrt)
	if err != nil {
		return fmt.Errorf("failed to read client certificate file: %v", err)
	}

	err = os.WriteFile(fmt.Sprintf(hostClientCrt, baseDir, dirName), certBytes, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to write client certificate to host: %v", err)
	}

	keyBytes, err := os.ReadFile(containerKey)
	if err != nil {
		return fmt.Errorf("failed to read client key file: %v", err)
	}

	err = os.WriteFile(fmt.Sprintf(hostClientKey, baseDir, dirName), keyBytes, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to write client key to host: %v", err)
	}

	caBytes, err := os.ReadFile(containerCa)
	if err != nil {
		return fmt.Errorf("failed to read client key file: %v", err)
	}

	err = os.WriteFile(fmt.Sprintf(hostClientCa, baseDir, dirName), caBytes, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to write client key to host: %v", err)
	}

	sCrt, err := os.ReadFile(fmt.Sprintf(serverCrt, baseDir, dirName))
	if err != nil {
		return fmt.Errorf("failed to write server crt to host: %v", err)
	}

	err = os.WriteFile(fmt.Sprintf(hostServerCa, baseDir, dirName), sCrt, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to write server crt to host: %v", err)
	}

	imports := []string{
		fmt.Sprintf(hostClientDir, baseDir, dirName),
		fmt.Sprintf(hostServerCa, baseDir, dirName),
	}

	for _, e := range imports {
		cmd := exec.Command("powershell", "-Command", "Import-Certificate", "-FilePath", e, "-CertStoreLocation", "Cert:\\LocalMachine\\Root", "-Verbose")
		fmt.Println(cmd.String())
		o, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Println(fmt.Errorf("failed to add server certificate to windows cert store: %v", err))
		}
		fmt.Println(string(o))
	}

	return nil
}
