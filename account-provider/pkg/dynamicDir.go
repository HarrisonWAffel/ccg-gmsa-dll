package pkg

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

const (
	gmsaDirectory = "/var/lib/rancher/gmsa"

	serverKey = "%s/%s/container/ssl/server/tls.key"

	containerSslDir = "%s/%s/container"

	containerServerCa  = "%s/%s/container/ssl/server/ca.crt"
	containerServerCrt = "%s/%s/container/ssl/server/tls.crt"

	containerClientDir = containerSslDir + "/ssl/client"
	containerClientCa  = containerClientDir + "/ca.crt"
	containerClientCrt = containerClientDir + "/tls.crt"
	containerClientKey = containerClientDir + "/tls.key"

	containerRootCaDir = "%s/%s/container/ssl/ca"
	containerRootCa    = containerRootCaDir + "/ca.crt"
	containerRootCrt   = containerRootCaDir + "/tls.crt"

	hostSslDir = "%s/%s/ssl"

	hostCaDir   = hostSslDir + "/ca"
	hostRootCa  = hostCaDir + "/ca.crt"
	hostRootCrt = hostCaDir + "/tls.crt"

	hostClientDir = hostSslDir + "/client"
	hostClientCa  = hostClientDir + "/ca.crt"
	hostClientCrt = hostClientDir + "/tls.crt"
	hostClientKey = hostClientDir + "/tls.key"

	hostServerDir = hostSslDir + "/server"
	hostServerCa  = hostServerDir + "/ca.crt"
	hostServerCrt = hostServerDir + "/tls.crt"
)

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
	// TODO: adjust certFile permissions
	if _, err := os.Stat(portFile); os.IsNotExist(err) {
		// create the certFile
		err = os.WriteFile(portFile, []byte(port), os.ModePerm)
		if err != nil {
			return fmt.Errorf("failed to create port.txt: %v", err)
		}
	}

	// update certFile with new port
	f, err := os.OpenFile(portFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return fmt.Errorf("failed to open port certFile: %v", err)
	}

	_, err = f.WriteString(port)
	if err != nil {
		return fmt.Errorf("failed to update port certFile: %v", err)
	}

	return f.Close()
}

func createDirectory(dirName, directory string) error {
	fullDirectory := fmt.Sprintf(directory, gmsaDirectory, dirName)
	err := os.Mkdir(fullDirectory, os.ModePerm)
	if err != nil && !strings.Contains(err.Error(), "already exists") {
		return fmt.Errorf("failed to create directory %s: %v", fullDirectory, err)
	}
	return nil
}

func WriteCerts(dirName string) error {

	directories := []string{
		hostSslDir,
		hostClientDir,
		hostServerDir,
		hostCaDir,
	}

	for _, e := range directories {
		err := createDirectory(dirName, e)
		if err != nil {
			return fmt.Errorf("failed to setup base host certificate directories: %v", err)
		}
	}

	type certFile struct {
		// isKey indicates the file should be written to the host
		// but not imported as a certificate
		isKey bool
		// pfxConvert indicates that the certificate should be
		// passed to certutil. If a certificate has pfxConvert = true
		// then there needs to be an associated key file in the same directory
		// with the same name (tls.crt & tls.key)
		pfxConvert bool
		// where in the container fs the file is
		containerDir string
		// where in the host fs the file should be written to
		hostDir string
	}

	//todo; trim this down to only what is needed

	files := []certFile{
		{
			containerDir: fmt.Sprintf(containerRootCa, gmsaDirectory, dirName),
			hostDir:      fmt.Sprintf(hostRootCa, gmsaDirectory, dirName),
		},
		{
			containerDir: fmt.Sprintf(containerRootCrt, gmsaDirectory, dirName),
			hostDir:      fmt.Sprintf(hostRootCrt, gmsaDirectory, dirName),
		},
		{
			isKey:        true,
			containerDir: fmt.Sprintf(containerClientKey, gmsaDirectory, dirName),
			hostDir:      fmt.Sprintf(hostClientKey, gmsaDirectory, dirName),
		},
		{
			pfxConvert:   true,
			containerDir: fmt.Sprintf(containerClientCrt, gmsaDirectory, dirName),
			hostDir:      fmt.Sprintf(hostClientCrt, gmsaDirectory, dirName),
		},
		{
			containerDir: fmt.Sprintf(containerServerCrt, gmsaDirectory, dirName),
			hostDir:      fmt.Sprintf(hostServerCrt, gmsaDirectory, dirName),
		},
		{
			containerDir: fmt.Sprintf(containerServerCa, gmsaDirectory, dirName),
			hostDir:      fmt.Sprintf(hostServerCa, gmsaDirectory, dirName),
		},
	}

	for _, e := range files {
		bytes, err := os.ReadFile(e.containerDir)
		if err != nil {
			return fmt.Errorf("failed to read %s from container: %v", e.hostDir, err)
		}

		err = os.WriteFile(e.hostDir, bytes, os.ModePerm)
		if err != nil {
			return fmt.Errorf("failed to write %s to host: %v", e.hostDir, err)
		}

		if e.isKey {
			continue
		}

		// we need to create a PFX file for use by the C# dll. Using a PFX file simplifies the association between a client certificate and its private key,
		// and both are required for mTLS. The password we use to encrypt the pfx is arbitrary, but certutil requires it to be passed.
		// TODO: The DLL seems to be able to access this file without needing the password, likely because we import it using powershell
		if e.pfxConvert {
			cmd := exec.Command("powershell", "-Command", "cd", fmt.Sprintf(hostClientDir, gmsaDirectory, dirName), ";", "certutil", "-p", "\"password\"", "-MergePFX", "tls.crt", "tls.pfx")
			fmt.Println("generating PFX certFile: ", cmd.String())
			out, err := cmd.CombinedOutput()
			fmt.Println("PFX generation logs: ", string(out))
			fmt.Println("PFX generation error: ", err)

			// import the pfx cert onto the system
			cmd = exec.Command("powershell", "-Command", "cd", fmt.Sprintf(hostClientDir, gmsaDirectory, dirName), ";", "$secureString = ConvertTo-SecureString password -AsPlainText -Force", ";", "Import-PfxCertificate", "-Filepath", "tls.pfx", "-CertStoreLocation", "Cert:\\LocalMachine\\Root", "-Password", "$secureString")
			fmt.Println("Importing PFX certFile: ", cmd.String())
			out, err = cmd.CombinedOutput()
			fmt.Println("PFX Import logs: ", string(out))
			fmt.Println("PFX Image Error: ", err)

			// todo; destroy the key on host? Probably.

			continue
		}

		cmd := exec.Command("powershell", "-Command", "Import-Certificate", "-FilePath", e.hostDir, "-CertStoreLocation", "Cert:\\LocalMachine\\Root", "-Verbose")
		fmt.Println("Importing certificate: ", cmd.String())
		o, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Println(fmt.Errorf("failed to add certificate to windows cert store: %v", err))
		}
		fmt.Println("Import certificate logs: ", string(o))

	}

	return nil
}
