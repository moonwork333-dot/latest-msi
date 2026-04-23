package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const serviceName = "RemoteAgent"

// installAndConfigureService writes the config file and registers RemoteAgent.exe as a Windows service.
func installAndConfigureService() error {
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("could not determine exe path: %w", err)
	}
	exeDir := filepath.Dir(exePath)

	// Create directories
	configDir := filepath.Join(exeDir, "config")
	logsDir := filepath.Join(exeDir, "logs")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("could not create config dir: %w", err)
	}
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		return fmt.Errorf("could not create logs dir: %w", err)
	}

	// Write config file
	configPath := filepath.Join(configDir, "agent-config.json")
	serverURL := "wss://watson-parts.com/agenthub/ws"
	machineID := generateMachineID()

	cfg := map[string]interface{}{
		"machineId":         machineID,
		"serverUrl":         serverURL,
		"authToken":         "",
		"reconnectInterval": 5000,
		"screenshotQuality": 75,
		"logLevel":          "info",
		"installPath":       exeDir,
	}

	cfgBytes, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("could not marshal config: %w", err)
	}
	if err := os.WriteFile(configPath, cfgBytes, 0644); err != nil {
		return fmt.Errorf("could not write config file: %w", err)
	}

	fmt.Printf("Config written to: %s\n", configPath)
	fmt.Printf("Machine ID: %s\n", machineID)

	// Stop and remove any existing service
	exec.Command("sc", "stop", serviceName).Run()
	time.Sleep(2 * time.Second)
	exec.Command("sc", "delete", serviceName).Run()

	// Create the service pointing to this exe
	out, err := exec.Command("sc", "create", serviceName,
		"binPath=", exePath,
		"start=", "auto",
		"DisplayName=", "Remote Agent Service",
	).CombinedOutput()
	if err != nil {
		return fmt.Errorf("sc create failed: %w\n%s", err, string(out))
	}

	// Set description
	exec.Command("sc", "description", serviceName,
		"Remote control and monitoring agent",
	).Run()

	// Configure auto-restart on failure
	exec.Command("sc", "failure", serviceName,
		"reset=", "60",
		"actions=", "restart/5000/restart/10000/restart/30000",
	).Run()

	// Start the service
	out, err = exec.Command("sc", "start", serviceName).CombinedOutput()
	if err != nil {
		return fmt.Errorf("sc start failed: %w\n%s", err, string(out))
	}

	return nil
}

// uninstallWindowsService stops and removes the Windows service.
func uninstallWindowsService() error {
	exec.Command("sc", "stop", serviceName).Run()
	time.Sleep(2 * time.Second)

	out, err := exec.Command("sc", "delete", serviceName).CombinedOutput()
	if err != nil {
		return fmt.Errorf("sc delete failed: %w\n%s", err, string(out))
	}
	return nil
}

// generateMachineID creates a unique machine identifier from hostname and MAC address.
func generateMachineID() string {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	out, err := exec.Command("getmac", "/fo", "csv", "/nh").Output()
	if err != nil {
		return fmt.Sprintf("%s-%d", strings.ToLower(hostname), time.Now().UnixMilli())
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) == 0 {
		return strings.ToLower(hostname)
	}

	parts := strings.Split(strings.Trim(lines[0], `"`), `","`)
	mac := ""
	if len(parts) > 0 {
		mac = strings.ReplaceAll(strings.Trim(parts[0], `"`), "-", "")
	}

	id := fmt.Sprintf("%s-%s", strings.ToLower(hostname), strings.ToLower(mac))
	var sanitized strings.Builder
	for _, ch := range id {
		if (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') || ch == '-' {
			sanitized.WriteRune(ch)
		} else {
			sanitized.WriteRune('-')
		}
	}
	return sanitized.String()
}
