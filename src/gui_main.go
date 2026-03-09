//go:build gui

package main

import (
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	app := NewApp()
	syncSvc := &SyncService{}
	secretsSvc := &SecretsService{}
	discoverySvc := &DiscoveryService{}
	keysSvc := &KeysService{}
	modeSvc := &ModeService{}
	peerSvc := &PeerService{}
	statusSvc := &StatusService{}
	serviceSvc := &ServiceMgmtService{}
	cronSvc := &CronService{}
	backupSvc := &BackupService{}
	logSvc := &LogService{}

	err := wails.Run(&options.App{
		Title:     "env-sync",
		Width:     1200,
		Height:    800,
		MinWidth:  900,
		MinHeight: 600,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 24, G: 24, B: 27, A: 1},
		OnStartup:        app.startup,
		OnShutdown:       app.shutdown,
		Bind: []interface{}{
			app,
			syncSvc,
			secretsSvc,
			discoverySvc,
			keysSvc,
			modeSvc,
			peerSvc,
			statusSvc,
			serviceSvc,
			cronSvc,
			backupSvc,
			logSvc,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
