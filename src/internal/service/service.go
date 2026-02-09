package service

import (
	"errors"
	"fmt"

	"github.com/kardianos/service"

	"envsync/internal/logging"
)

// StopIfRunning stops the env-sync service if it's running
// Returns true if the service was stopped, false if it wasn't running
func StopIfRunning() (bool, error) {
	svc, err := getService()
	if err != nil {
		return false, err
	}

	status, err := svc.Status()
	if err != nil {
		if errors.Is(err, service.ErrNotInstalled) {
			// Service not installed, nothing to stop
			return false, nil
		}
		return false, err
	}

	if status != service.StatusRunning {
		// Service installed but not running
		return false, nil
	}

	// Service is running, stop it
	if err := svc.Stop(); err != nil {
		return false, fmt.Errorf("failed to stop service: %w", err)
	}

	logging.Log("INFO", "Stopped env-sync service for installation")
	return true, nil
}

// RestartIfNeeded restarts the service if it was previously stopped
func RestartIfNeeded(wasStopped bool) error {
	if !wasStopped {
		return nil
	}

	svc, err := getService()
	if err != nil {
		return err
	}

	if err := svc.Start(); err != nil {
		return fmt.Errorf("failed to restart service: %w", err)
	}

	logging.Log("SUCCESS", "Restarted env-sync service")
	return nil
}

// UninstallIfInstalled uninstalls the service if it's installed
// Returns true if the service was uninstalled
func UninstallIfInstalled() (bool, error) {
	svc, err := getService()
	if err != nil {
		return false, err
	}

	status, err := svc.Status()
	if err != nil {
		if errors.Is(err, service.ErrNotInstalled) {
			// Service not installed
			return false, nil
		}
		return false, err
	}

	// Stop the service if running
	if status == service.StatusRunning {
		if err := svc.Stop(); err != nil {
			return false, fmt.Errorf("failed to stop service before uninstall: %w", err)
		}
	}

	// Uninstall the service
	if err := svc.Uninstall(); err != nil {
		return false, fmt.Errorf("failed to uninstall service: %w", err)
	}

	logging.Log("INFO", "Uninstalled env-sync service")
	return true, nil
}

// getService creates a service instance for management operations
func getService() (service.Service, error) {
	// Create a minimal service config just for management operations
	// We don't need the full program implementation for stop/start/uninstall
	prg := &dummyProgram{}
	svcConfig := &service.Config{
		Name:        "env-sync",
		DisplayName: "env-sync",
		Description: "env-sync HTTP server",
		Option:      service.KeyValue{"UserService": true},
	}
	return service.New(prg, svcConfig)
}

// dummyProgram is a minimal implementation of service.Interface
// Only used for service management operations (stop/start/uninstall)
type dummyProgram struct{}

func (p *dummyProgram) Start(_ service.Service) error {
	// Not used for management operations
	return nil
}

func (p *dummyProgram) Stop(_ service.Service) error {
	// Not used for management operations
	return nil
}
