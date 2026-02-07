package server

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"envsync/internal/config"
	"envsync/internal/logging"
	"envsync/internal/metadata"
	"envsync/internal/secrets"
)

type Options struct {
	Port   string
	Daemon bool
	Quiet  bool
}

func Run(opts Options) error {
	if opts.Port == "" {
		opts.Port = config.EnvSyncPort()
	}

	if err := secrets.ValidateSecretsFile(config.SecretsFile()); err != nil {
		logging.Log("ERROR", "Invalid secrets file")
		return err
	}

	if opts.Daemon {
		return runDaemon(opts)
	}

	return runServer(opts)
}

func runDaemon(opts Options) error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	args := []string{"serve", "--port", opts.Port}
	if opts.Quiet {
		args = append(args, "--quiet")
	}
	cmd := exec.Command(exe, args...)
	devNull, _ := os.OpenFile(os.DevNull, os.O_RDWR, 0o600)
	cmd.Stdout = devNull
	cmd.Stderr = devNull
	cmd.Stdin = devNull
	if err := cmd.Start(); err != nil {
		return err
	}
	logging.Log("SUCCESS", fmt.Sprintf("Server started in background (PID: %d)", cmd.Process.Pid))
	return nil
}

func runServer(opts Options) error {
	if err := checkPort(opts.Port); err != nil {
		logging.Log("WARN", fmt.Sprintf("Port %s is already in use", opts.Port))
		return err
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
	return server.ListenAndServe()
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

func PathToBinary() string {
	exe, err := os.Executable()
	if err != nil {
		return ""
	}
	if filepath.Base(exe) == "env-sync" {
		return exe
	}
	return exe
}
