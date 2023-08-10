package pkg

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	log "github.com/sirupsen/logrus"
)

const (
	serverKey = "%s/%s/container/ssl/server/tls.key"

	containerSslDir = "%s/%s/container"

	containerServerCa  = "%s/%s/container/ssl/server/ca.crt"
	containerServerCrt = "%s/%s/container/ssl/server/tls.crt"

	containerClientDir      = containerSslDir + "/ssl/client"
	containerClientCrt      = containerClientDir + "/tls.crt"
	containerClientKey      = containerClientDir + "/tls.key"
	containerClientKeystore = containerClientDir + "/keystore.p12"

	containerRootCaDir = "%s/%s/container/ssl/ca"
	containerRootCa    = containerRootCaDir + "/ca.crt"
	containerRootCrt   = containerRootCaDir + "/tls.crt"

	hostSslDir = "%s/%s/ssl"

	hostRootCaDir = hostSslDir + "/ca"
	hostRootCa    = hostRootCaDir + "/ca.crt"
	hostRootCrt   = hostRootCaDir + "/tls.crt"

	hostClientDir      = hostSslDir + "/client"
	hostClientCrt      = hostClientDir + "/tls.crt"
	hostClientKey      = hostClientDir + "/tls.key"
	hostClientKeystore = hostClientDir + "/keystore.p12"

	hostServerDir = hostSslDir + "/server"
	hostServerCa  = hostServerDir + "/ca.crt"
	hostServerCrt = hostServerDir + "/tls.crt"
)

// mTLSFile represents a file used to set up mTLS for the provider
// these files are utilized by the plugin to perform the client side
// handshake
type mTLSFile struct {
	// isKey indicates the file should be written to the host
	// but not imported as a certificate
	isKey bool
	// where in the container fs the file is
	containerFile string
	// the base directory on the host which the file will be placed into
	hostDir string
	// the full path of the file to be written on the host
	hostFile string
}

func WriteCerts(dirName string) error {
	err := os.Mkdir(fmt.Sprintf(hostSslDir, gmsaDirectory, dirName), os.ModePerm)
	if err != nil && !strings.Contains(err.Error(), "already exists") {
		return fmt.Errorf("failed to create base ssl dir: %v", err)
	}

	files := []mTLSFile{
		{
			containerFile: fmt.Sprintf(containerRootCa, gmsaDirectory, dirName),
			hostFile:      fmt.Sprintf(hostRootCa, gmsaDirectory, dirName),
			hostDir:       fmt.Sprintf(hostRootCaDir, gmsaDirectory, dirName),
		},
		{
			containerFile: fmt.Sprintf(containerRootCrt, gmsaDirectory, dirName),
			hostFile:      fmt.Sprintf(hostRootCrt, gmsaDirectory, dirName),
			hostDir:       fmt.Sprintf(hostRootCaDir, gmsaDirectory, dirName),
		},
		{
			containerFile: fmt.Sprintf(containerClientCrt, gmsaDirectory, dirName),
			hostFile:      fmt.Sprintf(hostClientCrt, gmsaDirectory, dirName),
			hostDir:       fmt.Sprintf(hostClientDir, gmsaDirectory, dirName),
		},
		{
			containerFile: fmt.Sprintf(containerServerCrt, gmsaDirectory, dirName),
			hostFile:      fmt.Sprintf(hostServerCrt, gmsaDirectory, dirName),
			hostDir:       fmt.Sprintf(hostServerDir, gmsaDirectory, dirName),
		},
		{
			containerFile: fmt.Sprintf(containerServerCa, gmsaDirectory, dirName),
			hostFile:      fmt.Sprintf(hostServerCa, gmsaDirectory, dirName),
			hostDir:       fmt.Sprintf(hostServerDir, gmsaDirectory, dirName),
		},
		{
			isKey:         true,
			containerFile: fmt.Sprintf(containerClientKeystore, gmsaDirectory, dirName),
			hostFile:      fmt.Sprintf(hostClientKeystore, gmsaDirectory, dirName),
			hostDir:       fmt.Sprintf(hostClientDir, gmsaDirectory, dirName),
		},
	}

	for _, e := range files {
		err := createDirectory(e.hostDir)
		if err != nil {
			return fmt.Errorf("failed to setup base host certificate directories: %v", err)
		}
	}

	for _, file := range files {
		bytes, err := os.ReadFile(file.containerFile)
		if err != nil {
			return fmt.Errorf("failed to read %s from container: %v", file.hostDir, err)
		}

		err = os.WriteFile(file.hostFile, bytes, os.ModePerm)
		if err != nil {
			return fmt.Errorf("failed to write %s to host: %v", file.hostDir, err)
		}

		if file.isKey {
			continue
		}

		err = importCertificate(file)
		if err != nil {
			log.Errorf("failed to import certificate %s: %v", file.hostFile, err)
		}
	}

	return nil
}

// createDirectory creates a Windows directory. If the directory already exists, it will return nil.
func createDirectory(name string) error {
	err := os.Mkdir(name, os.ModePerm)
	if err != nil && !strings.Contains(err.Error(), "already exists") {
		return fmt.Errorf("failed to create directory %s: %v", name, err)
	}
	return nil
}

// importCertificate adds a certificate file to the windows certificate store located in Cert:\LocalMachine\Root
func importCertificate(file mTLSFile) error {
	cmd := exec.Command("powershell", "-Command", "Import-Certificate", "-FilePath", file.hostFile, "-CertStoreLocation", "Cert:\\LocalMachine\\Root", "-Verbose")
	log.Debugf("Importing certificate: %s", cmd.String())
	o, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to add certificate to windows cert store: %v", err)
	}
	log.Debugf("Import certificate logs: %s", string(o))
	return nil
}
