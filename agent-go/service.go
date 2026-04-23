package main

import (
	"fmt"
	"os"
	"os/exec"
)

const serviceName = "RemoteAgent"

func installService() error {
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("could not determine exe path: %w", err)
	}

	exec.Command("sc", "stop", serviceName).Run()
	exec.Command("timeout", "/t", "3", "/nobreak").Run()
	exec.Command("sc", "delete", serviceName).Run()

	out, err := exec.Command("sc", "create", serviceName,
		"binPath=", exePath,
		"start=", "auto",
		"DisplayName=", "Remote Agent Service",
	).CombinedOutput()
	if err != nil {
		return fmt.Errorf("sc create failed: %w\n%s", err, string(out))
	}

	exec.Command("sc", "description", serviceName, "Remote control and monitoring agent").Run()
	exec.Command("sc", "failure", serviceName, "reset=", "60", "actions=", "restart/5000/restart/10000/restart/30000").Run()

	out, err = exec.Command("sc", "start", serviceName).CombinedOutput()
	if err != nil {
		return fmt.Errorf("sc start failed: %w\n%s", err, string(out))
	}
	return nil
}

func uninstallService() error {
	exec.Command("sc", "stop", serviceName).Run()
	exec.Command("timeout", "/t", "3", "/nobreak").Run()

	out, err := exec.Command("sc", "delete", serviceName).CombinedOutput()
	if err != nil {
		return fmt.Errorf("sc delete failed: %w\n%s", err, string(out))
	}
	return nil
}
