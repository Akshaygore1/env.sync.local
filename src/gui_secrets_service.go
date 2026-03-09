//go:build gui

package main

import (
	"encoding/json"
	"envsync/internal/config"
	"envsync/internal/keys"
	"envsync/internal/metadata"
	"envsync/internal/secrets"
	"fmt"
	"strings"
)

// SecretsService provides secrets CRUD operations
type SecretsService struct{}

// SecretEntry represents a single secret key-value pair
type SecretEntry struct {
	Key       string `json:"key"`
	Value     string `json:"value"`
	UpdatedAt string `json:"updatedAt"`
}

// List returns all secrets from the file
func (s *SecretsService) List() ([]SecretEntry, error) {
	file := config.SecretsFile()
	content, err := secrets.GetSecretsContent(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read secrets: %w", err)
	}

	encrypted := keys.IsFileEncrypted(file)
	return parseSecrets(content, encrypted)
}

// Get returns a single secret by key
func (s *SecretsService) Get(key string) (SecretEntry, error) {
	entries, err := s.List()
	if err != nil {
		return SecretEntry{}, err
	}

	for _, e := range entries {
		if e.Key == key {
			return e, nil
		}
	}
	return SecretEntry{}, fmt.Errorf("key not found: %s", key)
}

// Add adds or updates a secret
func (s *SecretsService) Add(key, value string) error {
	file := config.SecretsFile()
	content, err := secrets.GetSecretsContent(file)
	if err != nil {
		return fmt.Errorf("failed to read secrets: %w", err)
	}

	timestamp := secrets.GetTimestamp()
	encrypted := keys.IsFileEncrypted(file)

	var finalValue string
	if encrypted {
		recipients := keys.GetAllKnownRecipients()
		if len(recipients) == 0 {
			return fmt.Errorf("no recipients available for encryption")
		}
		enc, err := keys.EncryptValue(value, recipients)
		if err != nil {
			return fmt.Errorf("failed to encrypt value: %w", err)
		}
		finalValue = enc
	} else {
		finalValue = value
	}

	newLine := fmt.Sprintf("%s=\"%s\" # ENVSYNC_UPDATED_AT=%s", key, finalValue, timestamp)

	lines := strings.Split(content, "\n")
	found := false
	for i, line := range lines {
		parts := strings.SplitN(strings.TrimSpace(line), "=", 2)
		if len(parts) >= 1 && parts[0] == key {
			lines[i] = newLine
			found = true
			break
		}
	}

	if !found {
		lines = append(lines, newLine)
	}

	newContent := strings.Join(lines, "\n")
	if err := secrets.SetSecretsContent(file, newContent); err != nil {
		return fmt.Errorf("failed to write secrets: %w", err)
	}

	return metadata.UpdateChecksum(file)
}

// Remove deletes a secret by key
func (s *SecretsService) Remove(key string) error {
	file := config.SecretsFile()
	content, err := secrets.GetSecretsContent(file)
	if err != nil {
		return fmt.Errorf("failed to read secrets: %w", err)
	}

	lines := strings.Split(content, "\n")
	var newLines []string
	found := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		parts := strings.SplitN(trimmed, "=", 2)
		if len(parts) >= 1 && parts[0] == key {
			found = true
			continue
		}
		newLines = append(newLines, line)
	}

	if !found {
		return fmt.Errorf("key not found: %s", key)
	}

	newContent := strings.Join(newLines, "\n")
	if err := secrets.SetSecretsContent(file, newContent); err != nil {
		return fmt.Errorf("failed to write secrets: %w", err)
	}

	return metadata.UpdateChecksum(file)
}

// ExportEnv exports secrets in .env format
func (s *SecretsService) ExportEnv() (string, error) {
	entries, err := s.List()
	if err != nil {
		return "", err
	}

	var lines []string
	for _, e := range entries {
		lines = append(lines, fmt.Sprintf("%s=\"%s\"", e.Key, e.Value))
	}
	return strings.Join(lines, "\n"), nil
}

// ExportJSON exports secrets as JSON
func (s *SecretsService) ExportJSON() (string, error) {
	entries, err := s.List()
	if err != nil {
		return "", err
	}

	m := make(map[string]string)
	for _, e := range entries {
		m[e.Key] = e.Value
	}

	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// Initialize creates a new secrets file
func (s *SecretsService) Initialize(encrypted bool) error {
	file := config.SecretsFile()
	timestamp := config.InitTimestamp()

	if err := secrets.InitSecretsFile(file, timestamp); err != nil {
		return fmt.Errorf("failed to initialize secrets file: %w", err)
	}

	if encrypted {
		if err := keys.GenerateKey(); err != nil {
			return fmt.Errorf("failed to generate encryption key: %w", err)
		}
		hostname := secrets.GetHostname()
		if err := metadata.EnsureEncryptedMetadata(file, hostname); err != nil {
			return fmt.Errorf("failed to set encryption metadata: %w", err)
		}
	}

	return nil
}

// EncryptExisting encrypts an existing plaintext secrets file
func (s *SecretsService) EncryptExisting() error {
	file := config.SecretsFile()
	if keys.IsFileEncrypted(file) {
		return fmt.Errorf("file is already encrypted")
	}

	if err := keys.GenerateKey(); err != nil {
		return fmt.Errorf("failed to generate key: %w", err)
	}

	hostname := secrets.GetHostname()
	if err := metadata.EnsureEncryptedMetadata(file, hostname); err != nil {
		return fmt.Errorf("failed to set encryption metadata: %w", err)
	}

	content, err := secrets.GetSecretsContent(file)
	if err != nil {
		return err
	}

	recipients := keys.GetAllKnownRecipients()
	lines := strings.Split(content, "\n")
	var newLines []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			newLines = append(newLines, line)
			continue
		}

		parts := strings.SplitN(trimmed, "=", 2)
		if len(parts) != 2 {
			newLines = append(newLines, line)
			continue
		}

		key := parts[0]
		valuePart := parts[1]
		value := valuePart
		comment := ""
		if idx := strings.Index(valuePart, " # "); idx != -1 {
			value = valuePart[:idx]
			comment = valuePart[idx:]
		}
		value = strings.Trim(value, "\"")

		enc, err := keys.EncryptValue(value, recipients)
		if err != nil {
			return fmt.Errorf("failed to encrypt %s: %w", key, err)
		}

		newLines = append(newLines, fmt.Sprintf("%s=\"%s\"%s", key, enc, comment))
	}

	return secrets.SetSecretsContent(file, strings.Join(newLines, "\n"))
}

// IsEncrypted returns whether the secrets file is encrypted
func (s *SecretsService) IsEncrypted() bool {
	return keys.IsFileEncrypted(config.SecretsFile())
}

func parseSecrets(content string, encrypted bool) ([]SecretEntry, error) {
	var entries []SecretEntry
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		parts := strings.SplitN(trimmed, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := parts[0]
		valuePart := parts[1]

		updatedAt := ""
		if idx := strings.Index(valuePart, "ENVSYNC_UPDATED_AT="); idx != -1 {
			updatedAt = strings.TrimSpace(valuePart[idx+len("ENVSYNC_UPDATED_AT="):])
			if hashIdx := strings.LastIndex(valuePart[:idx], "#"); hashIdx != -1 {
				valuePart = strings.TrimSpace(valuePart[:hashIdx])
			}
		}

		value := strings.Trim(valuePart, "\"")

		if encrypted && value != "" {
			dec, err := keys.DecryptValue(value)
			if err == nil {
				value = dec
			}
		}

		entries = append(entries, SecretEntry{
			Key:       key,
			Value:     value,
			UpdatedAt: updatedAt,
		})
	}

	return entries, nil
}
