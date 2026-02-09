package service

import (
	"testing"
)

// TestStopIfRunning tests the StopIfRunning function
// This test verifies that the function handles a non-installed service gracefully
func TestStopIfRunning(t *testing.T) {
	// When service is not installed, should return false and no error
	wasStopped, err := StopIfRunning()
	if err != nil {
		t.Errorf("StopIfRunning() returned error when service not installed: %v", err)
	}
	if wasStopped {
		t.Errorf("StopIfRunning() returned true when service was not running")
	}
}

// TestRestartIfNeeded tests the RestartIfNeeded function
func TestRestartIfNeeded(t *testing.T) {
	// When wasStopped is false, should do nothing
	err := RestartIfNeeded(false)
	if err != nil {
		t.Errorf("RestartIfNeeded(false) returned error: %v", err)
	}
}

// TestUninstallIfInstalled tests the UninstallIfInstalled function
func TestUninstallIfInstalled(t *testing.T) {
	// When service is not installed, should return false and no error
	wasUninstalled, err := UninstallIfInstalled()
	if err != nil {
		t.Errorf("UninstallIfInstalled() returned error when service not installed: %v", err)
	}
	if wasUninstalled {
		t.Errorf("UninstallIfInstalled() returned true when service was not installed")
	}
}

// TestGetService tests that getService creates a valid service instance
func TestGetService(t *testing.T) {
	svc, err := getService()
	if err != nil {
		t.Fatalf("getService() returned error: %v", err)
	}
	if svc == nil {
		t.Fatal("getService() returned nil service")
	}
}
