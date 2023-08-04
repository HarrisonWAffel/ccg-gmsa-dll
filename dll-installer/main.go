//go:build windows

package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"golang.org/x/sys/windows/registry"
)

//go:embed RanchergMSACredentialProvider.dll
var dll []byte

//go:embed install-plugin.ps1
var installer []byte

//go:embed uninstall-plugin.ps1
var uninstaller []byte

const (
	// baseDir is where we expect the dll to live
	baseDir = "C:\\Program Files\\RanchergMSACredentialProvider"
	// CCGCOMClassKey is the Windows Registry Key which is used by CCG to invoke the dll
	CCGCOMClassKey = "SYSTEM\\CurrentControlSet\\Control\\CCG\\COMClasses\\{e4781092-f116-4b79-b55e-28eb6a224e26}"
	// ClassesRootKey is the Windows Registry Key which is added by regsvc upon registering the dll, and is also used to invoke the dll
	ClassesRootKey = "CLSID\\{E4781092-F116-4B79-B55E-28EB6A224E26}"

	installFileName   = "install-plugin.ps1"
	uninstallFileName = "uninstall-plugin.ps1"
	dllFileName       = "RanchergMSACredentialProvider.dll"
)

// todo; This should be an init container alongside a pause container!

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		fmt.Println("no argument supplied, defaulting to 'install'")
		if err := install(); err != nil {
			fmt.Println(fmt.Sprintf("failed to install plugin: %v", err))
		}
	}

	switch args[0] {
	case "install":
		if err := install(); err != nil {
			fmt.Println(fmt.Sprintf("failed to install plugin: %v", err))
		}
	case "uninstall":
		if err := uninstall(); err != nil {
			fmt.Println(fmt.Sprintf("failed to uninstall plugin: %v", err))
		}
	case "upgrade":
		if err := upgrade(); err != nil {
			fmt.Println(fmt.Sprintf("failed to upgrade plugin: %v", err))
		}
	default:
		panic(fmt.Sprintf("unknown argument %s", args[0]))
	}

	time.Sleep(10 * time.Minute)
}

func upgrade() error {

	// Some documentation on upgrading DLL's:
	//  https://serverfault.com/questions/503721/replacing-dll-files-while-the-application-is-running
	//	https://learn.microsoft.com/en-us/windows/win32/dlls/dynamic-link-library-updates

	b, err := os.ReadFile(fmt.Sprintf("%s/%s", baseDir, dllFileName))
	if err != nil {
		return err
	}

	// todo; we need to test this, in theory should work
	if bytes.Equal(b, dll) {
		return nil
	}

	oldDll := fmt.Sprintf("old-%s", dllFileName)
	dllPath := fmt.Sprintf("%s/%s", baseDir, dllFileName)
	oldDllPath := fmt.Sprintf("%s/%s", baseDir, oldDll)

	fmt.Println("New plugin version detected, attempting upgrade")
	_, err = os.Stat(oldDllPath)
	if err != nil {
		fmt.Printf("detected %s, deleting\n", oldDll)
		// we found an old DLL, we should remove it before upgrading.
		if err = os.Remove(oldDllPath); err != nil {
			return fmt.Errorf("failed to remove old plugin dll: %v", err)
		}
	}

	// rename the file, https://learn.microsoft.com/en-us/windows/win32/dlls/dynamic-link-library-updates
	err = os.Rename(dllPath, oldDllPath)

	// write the new file
	err = os.WriteFile(dllPath, dll, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to write dll file: %v", err)
	}

	fmt.Println("upgrade complete")
	return nil
}

func uninstall() error {

	installed, err := alreadyInstalled()
	if err != nil {
		return fmt.Errorf("failed to determine installation status: %v", err)
	}

	if !installed {
		fmt.Println("Did not find anything to uninstall")
		return nil
	}

	fmt.Println("Beginning uninstallation process...")
	err = os.WriteFile(fmt.Sprintf("%s\\%s", baseDir, uninstallFileName), uninstaller, os.ModePerm)
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
	// continuously try to remove the actual files.
	// We retry this process a few times because there may
	// still be instances of CCG referencing the DLL. Windows
	// will prevent the file from being deleted if any references
	// still exist. Eventually, the CCG instances will terminate and
	// all references will disappear, at which point the file can be
	// deleted. It goes without saying that if you're uninstalling this plugin,
	// you shouldn't be running workloads which need to use the plugin.
	successfulRemoval := false
	for i := 0; i < 10; i++ {
		err = os.RemoveAll(baseDir)
		if err == nil {
			successfulRemoval = true
			break
		}
		fmt.Println("encountered error removing DLL directory, some CCG instances may still be referencing the plugin. Will retry in 1 minute")
		time.Sleep(1 * time.Minute)
	}

	if !successfulRemoval {
		fmt.Printf("ERROR: Failed to remove DLL directory: %v\n", err)
	} else {
		fmt.Println("Uninstallation complete")
	}

	return nil
}

func install() error {
	installed, err := alreadyInstalled()
	if err != nil {
		return fmt.Errorf("failed to detect installation status: %v", err)
	}

	if installed {
		fmt.Println("plugin already installed")
		return nil
	}

	fmt.Println("beginning installation")

	err = writeArtifacts()
	if err != nil {
		return fmt.Errorf("failed to write artifacts: %v", err)
	}

	err = executeInstaller()
	if err != nil {
		return fmt.Errorf("failed to execute installation script: %v", err)
	}

	fmt.Println("Installation successful!")
	return nil
}

func alreadyInstalled() (bool, error) {
	// 1. Check that the DLL exists in the expected directory C:\Program Files\RanchergMSACredentialProvider
	_, err := os.Stat(fmt.Sprintf(baseDir))
	directoryDoesNotExist := err != nil

	_, err = os.Stat(fmt.Sprintf("%s\\%s", baseDir, dllFileName))
	fileDoesNotExist := err != nil

	// 2. Check the registry for a key in HKEY_LOCAL_MACHINE\SYSTEM\CurrentControlSet\Control\CCG\COMClasses\{e4781092-f116-4b79-b55e-28eb6a224e26}
	CCGEntryExists, err := CCGCOMClassExists(CCGCOMClassKey)
	if err != nil {
		return false, fmt.Errorf("failed to query CCG Com class key: %v", err)
	}

	// 3. Check the CLSID HKEY_CLASSES_ROOT\CLSID\{E4781092-F116-4B79-B55E-28EB6A224E26}
	ClassesRootKeyExists, err := CLSIDExists(ClassesRootKey)
	if err != nil {
		return false, fmt.Errorf("failed to query CLSID registry key: %v", err)
	}

	return !directoryDoesNotExist && !fileDoesNotExist && CCGEntryExists && ClassesRootKeyExists, nil
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
func CCGCOMClassExists(registryKey string) (bool, error) {
	var access uint32 = registry.QUERY_VALUE
	_, err := registry.OpenKey(registry.LOCAL_MACHINE, registryKey, access)
	if err != nil {
		if err != registry.ErrNotExist {
			return false, err
		}
		return false, nil
	}

	return true, nil
}

// CLSIDExists is used to get the CLSID value
func CLSIDExists(registryKey string) (bool, error) {
	var access uint32 = registry.QUERY_VALUE
	_, err := registry.OpenKey(registry.CLASSES_ROOT, registryKey, access)
	if err != nil {
		if err != registry.ErrNotExist {
			return false, err
		}
		return false, nil
	}

	return true, nil
}
