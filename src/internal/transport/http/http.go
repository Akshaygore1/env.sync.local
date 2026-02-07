package httptransport

import (
    "encoding/json"
    "errors"
    "fmt"
    "io"
    "net/http"
    "time"

    "envsync/internal/config"
)

type HealthResponse struct {
    Status    string `json:"status"`
    Version   string `json:"version"`
    Timestamp string `json:"timestamp"`
    Host      string `json:"host"`
}

func FetchHealth(host string) (HealthResponse, error) {
    url := fmt.Sprintf("http://%s:%s/health", host, config.EnvSyncPort())
    client := &http.Client{Timeout: 2 * time.Second}
    resp, err := client.Get(url)
    if err != nil {
        return HealthResponse{}, err
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return HealthResponse{}, err
    }
    var health HealthResponse
    if err := json.Unmarshal(body, &health); err != nil {
        return HealthResponse{}, err
    }
    if health.Status == "" {
        return HealthResponse{}, errors.New("missing status")
    }
    return health, nil
}

func FetchSecrets(host string) ([]byte, error) {
    url := fmt.Sprintf("http://%s:%s/secrets.env", host, config.EnvSyncPort())
    client := &http.Client{Timeout: 10 * time.Second}
    resp, err := client.Get(url)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("unexpected status: %s", resp.Status)
    }
    return io.ReadAll(resp.Body)
}
