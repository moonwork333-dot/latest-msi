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

type AgentConfig struct {
	MachineID         string `json:"machineId"`
	ServerURL         string `json:"serverUrl"`
	AuthToken         string `json:"authToken"`
	ReconnectInterval int    `json:"reconnectInterval"`
	ScreenshotQuality int    `json:"screenshotQuality"`
	LogLevel          string `json:"logLevel"`
	InstallPath       string `json:"installPath"`
}

func main() {
	installDir, err := os.Executable()
	if err != nil {
		fatalf("Could not get executable path: %v", err)
	}
	installDir = filepath.Dir(installDir)

	logPath := filepath.Join(installDir, "logs", "install.log")
	_ = os.MkdirAll(filepath.Join(installDir, "logs"), 0755)
	_ = os.MkdirAll(filepath.Join(installDir, "config"), 0755)

	logf := func(msg string, args ...interface{}) {
		line := fmt.Sprintf("[%s] %s\n", time.Now().Format(time.RFC3339), fmt.Sprintf(msg, args...))
		fmt.Print(line)
		f, _ := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if f != nil {
			f.WriteString(line)
			f.Close()
		}
	}

	logf("=== Remote Agent Post-Install ===")
	logf("Install directory: %s", installDir)

	configPath := filepath.Join(installDir, "config", "agent-config.json")
	serverURL := getEnvOrDefault("AGENT_SERVER_URL", "wss://watson-parts.com/agenthub/ws")
	authToken := getEnvOrDefault("AGENT_AUTH_TOKEN", "")
	machineID := generateMachineID()

	cfg := AgentConfig{
		MachineID:         machineID,
		ServerURL:         serverURL,
		AuthToken:         authToken,
		ReconnectInterval: 5000,
		ScreenshotQuality: 75,
		LogLevel:          "info",
		InstallPath:       installDir,
	}

	cfgBytes, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		fatalf("Failed to marshal config: %v", err)
	}
	if err := os.WriteFile(configPath, cfgBytes, 0644); err != nil {
		fatalf("Failed to write config: %v", err)
	}
	logf("Config written: %s", configPath)
	logf("Machine ID: %s", machineID)

	agentExe := filepath.Join(installDir, "RemoteAgent.exe")
	logf("Registering service pointing to: %s", agentExe)

	exec.Command("sc", "stop", serviceName).Run()
	time.Sleep(3 * time.Second)
	exec.Command("sc", "delete", serviceName).Run()

	out, err := exec.Command("sc", "create", serviceName,
		"binPath=", agentExe,
		"start=", "auto",
		"DisplayName=", "Remote Agent Service",
	).CombinedOutput()
	if err != nil {
		fatalf("sc create failed: %v\n%s", err, string(out))
	}
	logf("Service created")

	exec.Command("sc", "description", serviceName, "Remote control and monitoring agent").Run()
	exec.Command("sc", "failure", serviceName, "reset=", "60", "actions=", "restart/5000/restart/10000/restart/30000").Run()
	logf("Failure recovery configured")

	out, err = exec.Command("sc", "start", serviceName).CombinedOutput()
	if err != nil {
		fatalf("sc start failed: %v\n%s", err, string(out))
	}
	logf("Service started successfully")
	logf("=== Post-install complete ===")
}

func getEnvOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

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

func fatalf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "FATAL: "+format+"\n", args...)
	os.Exit(1)
}
