package util

import (
	"fmt"
	"os"
	"path/filepath"
)

func RabbitmqReset() {
	fmt.Printf("[*] Stopping Mythic\n")
	StopMythic([]string{})
	workingPath := GetCwdFromExe()
	fmt.Printf("[*] Removing rabbitmq files\n")
	err := os.RemoveAll(filepath.Join(workingPath, "rabbitmq-docker", "storage"))
	if err != nil {
		fmt.Printf("[-] Failed to reset rabbitmq files\n")
	} else {
		fmt.Printf("[+] Successfully reset rabbitmq files\n")
	}
}
