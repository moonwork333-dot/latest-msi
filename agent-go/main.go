package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

const (
	defaultServerURL      = "wss://your-rmm-server.com/agent"
	defaultReconnectDelay = 5 * time.Second
	configFileName        = "agent-config.json"
)

// Config holds the agent configuration loaded from disk
type Config struct {
	MachineID         string `json:"machineId"`
	ServerURL         string `json:"serverUrl"`
	AuthToken         string `json:"authToken"`
	ReconnectInterval int    `json:"reconnectInterval"`
	ScreenshotQuality int    `json:"screenshotQuality"`
	LogLevel          string `json:"logLevel"`
}

func loadConfig() (*Config, error) {
	// Look for config relative to the executable
	exePath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("could not determine executable path: %w", err)
	}
	exeDir := filepath.Dir(exePath)
	configPath := filepath.Join(exeDir, "config", configFileName)

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("could not read config file at %s: %w", configPath, err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("could not parse config file: %w", err)
	}

	// Apply defaults for missing values
	if cfg.ServerURL == "" {
		cfg.ServerURL = defaultServerURL
	}
	if cfg.ReconnectInterval == 0 {
		cfg.ReconnectInterval = 5000
	}
	if cfg.ScreenshotQuality == 0 {
		cfg.ScreenshotQuality = 80
	}

	return &cfg, nil
}

func setupLogger(logLevel string) *log.Logger {
	exePath, _ := os.Executable()
	exeDir := filepath.Dir(exePath)
	logDir := filepath.Join(exeDir, "logs")

	// Ensure logs directory exists
	_ = os.MkdirAll(logDir, 0755)

	logFile := filepath.Join(logDir, "agent.log")
	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		// Fall back to stdout
		return log.New(os.Stdout, "[RemoteAgent] ", log.LstdFlags)
	}

	return log.New(f, "[RemoteAgent] ", log.LstdFlags)
}

func main() {
	installService := flag.Bool("install-service", false, "Install as a Windows service (writes config and registers)")
	uninstallService := flag.Bool("uninstall-service", false, "Uninstall the Windows service")
	flag.Parse()

	if *installService {
		if err := installAndConfigureService(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to install service: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Service installed and configured successfully")
		return
	}

	if *uninstallService {
		if err := uninstallWindowsService(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to uninstall service: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Service uninstalled successfully")
		return
	}

	// Load configuration
	cfg, err := loadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Config error: %v\n", err)
		os.Exit(1)
	}

	logger := setupLogger(cfg.LogLevel)
	logger.Printf("Remote Agent starting - MachineID: %s", cfg.MachineID)
	logger.Printf("Connecting to: %s", cfg.ServerURL)

	// Create agent and start
	agent := NewAgent(cfg, logger)

	// Handle OS signals for graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigCh
		logger.Printf("Received signal: %v - shutting down", sig)
		agent.Stop()
	}()

	// Run with auto-reconnect
	for {
		if agent.stopped {
			break
		}
		logger.Println("Connecting to server...")
		if err := agent.Connect(); err != nil {
			logger.Printf("Connection error: %v - retrying in %v", err, defaultReconnectDelay)
		}
		if !agent.stopped {
			time.Sleep(defaultReconnectDelay)
		}
	}

	logger.Println("Remote Agent stopped")
}
