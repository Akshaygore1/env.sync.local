//go:build gui

package main

import (
	"envsync/internal/config"
	"envsync/internal/server"
	"envsync/internal/service"
	"fmt"
	"os"
	"strings"
)

// ServiceMgmtService provides background server management
type ServiceMgmtService struct{}

// StartServer starts the HTTP server
func (s *ServiceMgmtService) StartServer(port int, daemon bool) error {
	portStr := fmt.Sprintf("%d", port)
	if port <= 0 {
		portStr = config.EnvSyncPort()
	}

	return server.Run(server.Options{
		Port:   portStr,
		Daemon: daemon,
		Quiet:  true,
	})
}

// StopServer stops the running server
func (s *ServiceMgmtService) StopServer() error {
	_, err := service.StopIfRunning()
	return err
}

// RestartServer restarts the server
func (s *ServiceMgmtService) RestartServer() error {
	wasStopped, err := service.StopIfRunning()
	if err != nil {
		return fmt.Errorf("failed to stop server: %w", err)
	}
	return service.RestartIfNeeded(wasStopped)
}

// UninstallService uninstalls the system service
func (s *ServiceMgmtService) UninstallService() error {
	_, err := service.UninstallIfInstalled()
	return err
}

// GetServerPort returns the configured server port
func (s *ServiceMgmtService) GetServerPort() int {
	port := config.EnvSyncPort()
	var p int
	fmt.Sscanf(port, "%d", &p)
	return p
}

// IsServerRunning checks if the server process is alive
func (s *ServiceMgmtService) IsServerRunning() bool {
	pidFile := config.ServerPidFile()
	data, err := os.ReadFile(pidFile)
	if err != nil {
		return false
	}

	var pid int
	if _, err := fmt.Sscanf(strings.TrimSpace(string(data)), "%d", &pid); err != nil {
		return false
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	return process.Signal(nil) == nil
}
