//go:build windows

package main

import (
	_ "embed"
	"fmt"
	"golang.org/x/sys/windows/registry"
	"os"
	"os/exec"
	"time"
)

//go:embed RanchergMSACredentialProvider.dll
var dll []byte

//go:embed install-plugin.ps1
var installer []byte

const (
	// baseDir is where we expect the dll to live
	baseDir = "C:\\Program Files\\RanchergMSACredentialProvider"
	// CCGCOMClassKey is the Windows Registry Key which is used by CCG to invoke the dll
	CCGCOMClassKey = "SYSTEM\\CurrentControlSet\\Control\\CCG\\COMClasses\\{e4781092-f116-4b79-b55e-28eb6a224e26}"
	// ClassesRootKey is the Windows Registry Key which is added by regsvc upon registering the dll, and is also used to invoke the dll
	ClassesRootKey = "CLSID\\{E4781092-F116-4B79-B55E-28EB6A224E26}"
)

// todo; uninstall
func main() {
	for {
		fmt.Println("Checking installation...")
		if notAlreadyInstalled() {
			fmt.Println("Plugin is not installed, beginning installation")
			err := writeArtifacts()
			if err != nil {
				fmt.Println(fmt.Errorf("failed to write artifacts: %v", err))
				continue
			}

			err = executeInstaller()
			if err != nil {
				fmt.Println(fmt.Errorf("failed to execute installation script: %v", err))
				continue
			}

			fmt.Println("Installation successful!")
		}
		fmt.Println("Done checking installation")
		// wait around for a bit before double-checking the installation
		time.Sleep(5 * time.Minute)
	}
}

func notAlreadyInstalled() bool {

	// we should
	// 1. Check that the DLL exists in the expected directory C:\Program Files\RanchergMSACredentialProvider
	_, err := os.Stat(fmt.Sprintf(baseDir))
	directoryDoesNotExist := err != nil

	_, err = os.Stat(fmt.Sprintf("%s\\%s", baseDir, "RanchergMSACredentialProvider.dll"))
	fileDoesNotExist := err != nil

	// 2. Check the registry for a key in HKEY_LOCAL_MACHINE\SYSTEM\CurrentControlSet\Control\CCG\COMClasses\{e4781092-f116-4b79-b55e-28eb6a224e26}
	CCGEntryExists := CCGCOMClassExists(CCGCOMClassKey)

	// 3. Check the CLSID HKEY_CLASSES_ROOT\CLSID\{E4781092-F116-4B79-B55E-28EB6A224E26}
	ClassesRootKeyExists := CLSIDExists(ClassesRootKey)

	fmt.Println(fmt.Sprintf("Directory does not exist: %t, ccgentry %t, classroot %t", directoryDoesNotExist, CCGEntryExists, ClassesRootKeyExists))

	return directoryDoesNotExist || fileDoesNotExist || !CCGEntryExists || !ClassesRootKeyExists

	// 4. somehow check if the dll is out of date, and if a newer version needs to be installed
	//    4a. might be able to just check the bytes of the file, if there is any difference add the new file
	//    4b. We need to understand how to hot-swap dll's (doesn't seem super hard)
	//   		https://serverfault.com/questions/503721/replacing-dll-files-while-the-application-is-running
	//			https://learn.microsoft.com/en-us/windows/win32/dlls/dynamic-link-library-updates
}

func writeArtifacts() error {

	err := os.Mkdir(baseDir, os.ModePerm)
	if err != nil {
		return err
	}

	err = os.WriteFile(fmt.Sprintf("%s\\%s", baseDir, "RanchergMSACredentialProvider.dll"), dll, os.ModePerm)
	if err != nil {
		return err
	}

	err = os.WriteFile(fmt.Sprintf("%s\\%s", baseDir, "install-plugin.ps1"), installer, os.ModePerm)
	if err != nil {
		return err
	}

	return nil
}

func executeInstaller() error {
	// run installation command
	cmd := exec.Command("powershell.exe", "-File", fmt.Sprintf("%s\\%s", baseDir, "install-plugin.ps1"))
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(string(out))
		return err
	}
	return nil
}

// CCGCOMClassExists is used to get the ccg com entry
func CCGCOMClassExists(registryKey string) bool {
	var access uint32 = registry.QUERY_VALUE
	_, err := registry.OpenKey(registry.LOCAL_MACHINE, registryKey, access)
	if err != nil {
		if err != registry.ErrNotExist {
			panic(err)
		}
		return false
	}
	return true
}

// CLSIDExists is used to get the CLSID value
func CLSIDExists(registryKey string) bool {
	var access uint32 = registry.QUERY_VALUE
	_, err := registry.OpenKey(registry.CLASSES_ROOT, registryKey, access)
	if err != nil {
		if err != registry.ErrNotExist {
			panic(err)
		}
		return false
	}

	return true
}
