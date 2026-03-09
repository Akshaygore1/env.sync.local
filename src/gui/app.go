package main

import (
	"context"
	"os"

	"envsync/internal/config"
)

// App struct holds the application context
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// shutdown is called when the app is closing
func (a *App) shutdown(ctx context.Context) {
	// cleanup if needed
}

// ConfigPaths holds all configuration directory paths
type ConfigPaths struct {
	ConfigDir   string `json:"configDir"`
	SecretsFile string `json:"secretsFile"`
	BackupDir   string `json:"backupDir"`
	LogDir      string `json:"logDir"`
	KeysDir     string `json:"keysDir"`
}

// GetVersion returns the current env-sync version
func (a *App) GetVersion() string {
	return config.Version
}

// GetConfigPaths returns all configuration directory paths
func (a *App) GetConfigPaths() ConfigPaths {
	return ConfigPaths{
		ConfigDir:   config.ConfigDir(),
		SecretsFile: config.SecretsFile(),
		BackupDir:   config.BackupDir(),
		LogDir:      config.LogDir(),
		KeysDir:     config.AgeKeyDir(),
	}
}

// IsInitialized checks if the secrets file already exists
func (a *App) IsInitialized() bool {
	_, err := os.Stat(config.SecretsFile())
	return err == nil
}
