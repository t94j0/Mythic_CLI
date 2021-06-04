package util

import (
	"fmt"
	"os"
	"path/filepath"
)

func DatabaseReset() {
	fmt.Printf("[*] Stopping Mythic\n")
	StopMythic([]string{})
	workingPath := GetCwdFromExe()
	fmt.Printf("[*] Removing database files\n")
	err := os.RemoveAll(filepath.Join(workingPath, "postgres-docker", "database"))
	if err != nil {
		fmt.Printf("[-] Failed to remove database files\n")
	} else {
		fmt.Printf("[+] Successfully reset datbase files\n")
	}
}
