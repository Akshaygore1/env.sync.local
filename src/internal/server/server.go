package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/kardianos/service"

	"envsync/internal/config"
	"envsync/internal/discovery"
	"envsync/internal/logging"
	"envsync/internal/metadata"
	"envsync/internal/secrets"
	syncer "envsync/internal/sync"
)

type Options struct {
	Port         string
	Daemon       bool
	Quiet        bool
	Service      bool
	SyncInterval time.Duration
}

func Run(opts Options) error {
	if opts.Port == "" {
		opts.Port = config.EnvSyncPort()
	}
	if opts.SyncInterval == 0 {
		opts.SyncInterval = 30 * time.Minute
	}

	if err := secrets.ValidateSecretsFile(config.SecretsFile()); err != nil {
		logging.Log("ERROR", "Invalid secrets file")
		return err
	}

	if opts.Service {
		return runService(opts)
	}
	if opts.Daemon {
		return runDaemon(opts)
	}

	return runServer(opts)
}

func runDaemon(opts Options) error {
	svc, err := newService(opts)
	if err != nil {
		return err
	}
	status, err := svc.Status()
	if err != nil {
		if !errors.Is(err, service.ErrNotInstalled) {
			return err
		}
		if err := svc.Install(); err != nil {
			return err
		}
	} else if status == service.StatusRunning {
		logging.Log("INFO", "Server service already running")
		return nil
	}

	if err := svc.Start(); err != nil {
		return err
	}
	logging.Log("SUCCESS", "Server service started")
	return nil
}

func runServer(opts Options) error {
	server, err := buildServer(opts)
	if err != nil {
		return err
	}
	return server.ListenAndServe()
}

func runService(opts Options) error {
	svc, err := newService(opts)
	if err != nil {
		return err
	}
	return svc.Run()
}

func newService(opts Options) (service.Service, error) {
	prg := &serviceProgram{opts: opts}
	svcConfig := &service.Config{
		Name:        "env-sync",
		DisplayName: "env-sync",
		Description: "env-sync HTTP server",
		Arguments:   serviceArgs(opts),
		Option:      service.KeyValue{"UserService": true},
	}
	return service.New(prg, svcConfig)
}

func serviceArgs(opts Options) []string {
	args := []string{"serve", "--service"}
	if opts.Port != "" {
		args = append(args, "--port", opts.Port)
	}
	if opts.Quiet {
		args = append(args, "--quiet")
	}
	return args
}

type serviceProgram struct {
	opts       Options
	server     *http.Server
	cancel     context.CancelFunc
	advertiser *discovery.Advertiser
}

func (p *serviceProgram) Start(_ service.Service) error {
	ctx, cancel := context.WithCancel(context.Background())
	p.cancel = cancel

	server, err := buildServer(p.opts)
	if err != nil {
		return err
	}
	p.server = server
	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logging.Log("ERROR", fmt.Sprintf("Server stopped: %v", err))
		}
	}()
	if adv, err := discovery.StartAdvertiser(p.opts.Port); err == nil {
		p.advertiser = adv
	} else {
		logging.Log("WARN", fmt.Sprintf("Failed to start mDNS advertisement: %v", err))
	}
	go startSyncLoop(ctx, p.opts)
	return nil
}

func (p *serviceProgram) Stop(_ service.Service) error {
	if p.cancel != nil {
		p.cancel()
	}
	if p.advertiser != nil {
		_ = p.advertiser.Stop()
	}
	if p.server == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return p.server.Shutdown(ctx)
}

func startSyncLoop(ctx context.Context, opts Options) {
	syncInterval := opts.SyncInterval
	if syncInterval <= 0 {
		syncInterval = 30 * time.Minute
	}

	runSync := func() {
		if _, err := os.Stat(config.SecretsFile()); err != nil {
			logging.Log("WARN", "Skipping sync: secrets file not found")
			return
		}
		syncOpts := syncer.Options{Quiet: true}
		if err := syncer.Run(syncOpts); err != nil {
			logging.Log("WARN", fmt.Sprintf("Background sync failed: %v", err))
		}
	}

	runSync()
	ticker := time.NewTicker(syncInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			runSync()
		}
	}
}

func buildServer(opts Options) (*http.Server, error) {
	if err := checkPort(opts.Port); err != nil {
		logging.Log("WARN", fmt.Sprintf("Port %s is already in use", opts.Port))
		return nil, err
	}

	if err := os.MkdirAll(config.ConfigDir(), 0o700); err == nil {
		_ = os.WriteFile(config.ServerPidFile(), []byte(fmt.Sprintf("%d", os.Getpid())), 0o600)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/secrets.env", secretsHandler)
	mux.HandleFunc("/", notFoundHandler)

	logging.Log("INFO", fmt.Sprintf("Starting env-sync server on port %s", opts.Port))
	logging.Log("SUCCESS", fmt.Sprintf("Server listening on http://0.0.0.0:%s", opts.Port))
	logging.Log("INFO", "Endpoints:")
	logging.Log("INFO", "  - GET /health       - Health check")
	logging.Log("INFO", "  - GET /secrets.env  - Get secrets file")

	server := &http.Server{Addr: ":" + opts.Port, Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	return server, nil
}

func healthHandler(w http.ResponseWriter, _ *http.Request) {
	version := metadata.GetFileVersion(config.SecretsFile())
	timestamp := metadata.GetFileTimestamp(config.SecretsFile())
	host := metadata.GetFileHost(config.SecretsFile())
	response := fmt.Sprintf(`{"status": "ok", "version": "%s", "timestamp": "%s", "host": "%s"}`, version, timestamp, host)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(response)))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(response))
}

func secretsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	content, err := os.ReadFile(config.SecretsFile())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	version := metadata.GetFileVersion(config.SecretsFile())
	timestamp := metadata.GetFileTimestamp(config.SecretsFile())
	host := metadata.GetFileHost(config.SecretsFile())
	checksum := metadata.GetFileChecksum(config.SecretsFile())
	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(content)))
	w.Header().Set("X-EnvSync-Version", version)
	w.Header().Set("X-EnvSync-Timestamp", timestamp)
	w.Header().Set("X-EnvSync-Host", host)
	w.Header().Set("X-EnvSync-Checksum", checksum)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(content)
}

func notFoundHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusNotFound)
	_, _ = w.Write([]byte("Not Found"))
}

func checkPort(port string) error {
	ln, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}
	_ = ln.Close()
	return nil
}
