package main

import (
	"fmt"
	"os/exec"
	"time"
)

const serviceName = "RemoteAgent"

func main() {
	fmt.Println("=== Remote Agent Pre-Uninstall ===")

	fmt.Printf("Stopping service %q...\n", serviceName)
	out, err := exec.Command("sc", "stop", serviceName).CombinedOutput()
	if err != nil {
		fmt.Printf("sc stop (non-fatal): %v\n%s\n", err, string(out))
	} else {
		fmt.Println("Service stopped")
		time.Sleep(3 * time.Second)
	}

	fmt.Printf("Removing service %q...\n", serviceName)
	out, err = exec.Command("sc", "delete", serviceName).CombinedOutput()
	if err != nil {
		fmt.Printf("sc delete (non-fatal): %v\n%s\n", err, string(out))
	} else {
		fmt.Println("Service removed")
	}

	fmt.Println("=== Pre-uninstall complete ===")
}
