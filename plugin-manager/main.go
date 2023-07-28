//go:build windows

package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"golang.org/x/sys/windows/registry"
	"os"
	"os/exec"
	"strings"
	"time"
)

//go:embed RanchergMSACredentialProvider.dll
var dll []byte

//go:embed install-plugin.ps1
var installer []byte

//go:embed uninstall.ps1
var uninstaller []byte

const (
	// baseDir is where we expect the dll to live
	baseDir = "C:\\Program Files\\RanchergMSACredentialProvider"
	// CCGCOMClassKey is the Windows Registry Key which is used by CCG to invoke the dll
	CCGCOMClassKey = "SYSTEM\\CurrentControlSet\\Control\\CCG\\COMClasses\\{e4781092-f116-4b79-b55e-28eb6a224e26}"
	// ClassesRootKey is the Windows Registry Key which is added by regsvc upon registering the dll, and is also used to invoke the dll
	ClassesRootKey = "CLSID\\{E4781092-F116-4B79-B55E-28EB6A224E26}"

	installFileName   = "install-plugin.ps1"
	uninstallFileName = "uninstall.ps1"
	dllFileName       = "RanchergMSACredentialProvider.dll"
)

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		watchInstall()
	} else if args[0] == "uninstall" {
		if err := uninstall(); err != nil {
			panic(err)
		}
	} else if args[0] == "upgrade" {
		// todo
		panic("not yet implemented")
	} else {
		panic(fmt.Sprintf("unknown argument %s", args[0]))
	}
}

func upgrade() {
	// somehow check if the dll is out of date, and if a newer version needs to be installed
	//    4a. might be able to just check the bytes of the file, if there is any difference add the new file
	//    4b. We need to understand how to hot-swap dll's (doesn't seem super hard)
	//   		https://serverfault.com/questions/503721/replacing-dll-files-while-the-application-is-running
	//			https://learn.microsoft.com/en-us/windows/win32/dlls/dynamic-link-library-updates

	b, err := os.ReadFile(fmt.Sprintf("%s/%s", baseDir, dllFileName))
	if err != nil {
		panic(err)
	}

	changed := !bytes.Equal(b, dll)
	if !changed {
		return
	}

	err = os.Rename(fmt.Sprintf("%s/%s", baseDir, dllFileName), fmt.Sprintf("%s/%s", baseDir, fmt.Sprintf("old-%s", dllFileName)))

	err = os.WriteFile(fmt.Sprintf("%s\\%s", baseDir, dllFileName), dll, os.ModePerm)
	if err != nil && !strings.Contains(err.Error(), "already exists") {
		panic(fmt.Errorf("failed to write dll file: %v", err))
	}
}

func uninstall() error {

	fmt.Println("Beginning uninstallation process...")
	err := os.WriteFile(fmt.Sprintf("%s\\%s", baseDir, uninstallFileName), uninstaller, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to write install script: %v", err)
	}

	fmt.Println("executing uninstallation script...")
	cmd := exec.Command("powershell.exe", "-File", fmt.Sprintf("%s\\%s", baseDir, uninstallFileName))
	_, err = cmd.CombinedOutput()
	if err != nil {
		return err
	}

	fmt.Println("successfully executed uninstallation script")
	fmt.Println("attempting to remove DLL directory: ", baseDir)
	// continuously try to remove the actual files
	// we should do that part in go because
	// there may be instances of CCG still referencing the file. In which case a simple rm will fail
	// so we need to continuously try to rm it until all ccg instances are gone. We avoid workload
	// outages due to the use of regsvc. While old CCG instances may still have the DLL loaded into memory
	// they will eventually exit (needs to be confirmed via testing),
	// and new instances will be unable to load the DLL anymore since its reference will be removed from the registry.
	successfulRemoval := false
	for i := 0; i < 10; i++ {
		err = os.RemoveAll(baseDir)
		if err == nil {
			successfulRemoval = true
			break
		}
		fmt.Println("encountered error removing DLL directory, will retry in 1 minute")
		time.Sleep(1 * time.Minute)
	}

	if !successfulRemoval {
		fmt.Printf("ERROR: Failed to remove DLL directory: %v\n", err)
	} else {
		fmt.Println("Uninstallation complete")
	}

	return nil
}

func watchInstall() {
	for {
		fmt.Println("Checking installation...")
		if notAlreadyInstalled() {
			fmt.Println("Plugin is not installed, beginning installation in 30 seconds")
			time.Sleep(30 * time.Second)
			if installErr := install(); installErr != nil {
				fmt.Println(fmt.Sprintf("error encountered during installation: %v", installErr))
			} else {
				fmt.Println("Installation successful!")
			}
		} else {
			fmt.Println("Plugin already installed")
		}
		fmt.Println("Done checking installation")
		// wait around for a bit before double-checking the installation
		time.Sleep(5 * time.Minute)
	}
}

func install() error {
	err := writeArtifacts()
	if err != nil {
		return fmt.Errorf("failed to write artifacts: %v", err)
	}

	err = executeInstaller()
	if err != nil {
		return fmt.Errorf("failed to execute installation script: %v", err)
	}

	return nil
}

func notAlreadyInstalled() bool {

	// 1. Check that the DLL exists in the expected directory C:\Program Files\RanchergMSACredentialProvider
	_, err := os.Stat(fmt.Sprintf(baseDir))
	directoryDoesNotExist := err != nil

	_, err = os.Stat(fmt.Sprintf("%s\\%s", baseDir, dllFileName))
	fileDoesNotExist := err != nil

	// 2. Check the registry for a key in HKEY_LOCAL_MACHINE\SYSTEM\CurrentControlSet\Control\CCG\COMClasses\{e4781092-f116-4b79-b55e-28eb6a224e26}
	CCGEntryExists := CCGCOMClassExists(CCGCOMClassKey)

	// 3. Check the CLSID HKEY_CLASSES_ROOT\CLSID\{E4781092-F116-4B79-B55E-28EB6A224E26}
	ClassesRootKeyExists := CLSIDExists(ClassesRootKey)

	return directoryDoesNotExist || fileDoesNotExist || !CCGEntryExists || !ClassesRootKeyExists
}

func writeArtifacts() error {

	err := os.Mkdir(baseDir, os.ModePerm)
	if err != nil && !strings.Contains(err.Error(), "already exists") {
		return fmt.Errorf("failed to create base directory: %v", err)
	}

	err = os.WriteFile(fmt.Sprintf("%s\\%s", baseDir, dllFileName), dll, os.ModePerm)
	if err != nil && !strings.Contains(err.Error(), "already exists") {
		return fmt.Errorf("failed to write dll file: %v", err)
	}

	err = os.WriteFile(fmt.Sprintf("%s\\%s", baseDir, installFileName), installer, os.ModePerm)
	if err != nil && !strings.Contains(err.Error(), "already exists") {
		return fmt.Errorf("failed to write install script: %v", err)
	}

	return nil
}

func executeInstaller() error {
	// run installation command
	cmd := exec.Command("powershell.exe", "-File", fmt.Sprintf("%s\\%s", baseDir, installFileName))
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
