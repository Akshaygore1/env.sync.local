package metadata

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"regexp"
	"strings"
)

var checksumRegex = regexp.MustCompile(`(?m)^# CHECKSUM: .*?$`)

func ExtractMetadata(file string, key string) string {
	content, err := os.ReadFile(file)
	if err != nil {
		return ""
	}

	prefix := "# " + key + ": "
	for _, line := range strings.Split(string(content), "\n") {
		if strings.HasPrefix(line, prefix) {
			return strings.TrimSpace(strings.TrimPrefix(line, prefix))
		}
	}
	return ""
}

func GetFileVersion(file string) string {
	return ExtractMetadata(file, "VERSION")
}

func GetFileTimestamp(file string) string {
	return ExtractMetadata(file, "TIMESTAMP")
}

func GetFileHost(file string) string {
	return ExtractMetadata(file, "HOST")
}

func GetFileChecksum(file string) string {
	return ExtractMetadata(file, "CHECKSUM")
}

func GenerateChecksum(file string) (string, error) {
	content, err := os.ReadFile(file)
	if err != nil {
		return "", err
	}
	normalized := checksumRegex.ReplaceAll(content, []byte("# CHECKSUM: "))
	sum := sha256.Sum256(normalized)
	return hex.EncodeToString(sum[:]), nil
}

func UpdateChecksum(file string) error {
	content, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	cleared := checksumRegex.ReplaceAll(content, []byte("# CHECKSUM: "))
	checksum := sha256.Sum256(cleared)
	updated := checksumRegex.ReplaceAll(cleared, []byte("# CHECKSUM: "+hex.EncodeToString(checksum[:])))
	return os.WriteFile(file, updated, 0o600)
}

func UpdateMetadataFields(file string, updates map[string]string) error {
	content, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	lines := strings.Split(string(content), "\n")
	for i, line := range lines {
		for key, value := range updates {
			prefix := "# " + key + ":"
			if strings.HasPrefix(line, prefix) {
				lines[i] = "# " + key + ": " + value
			}
		}
	}
	return os.WriteFile(file, []byte(strings.Join(lines, "\n")), 0o600)
}

func EnsureEncryptedMetadata(file string, hostname string, recipients string) error {
	content, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	lines := strings.Split(string(content), "\n")
	output := make([]string, 0, len(lines)+2)
	inMetadata := false
	inserted := false

	for _, line := range lines {
		if strings.HasPrefix(line, "# === ENV_SYNC_METADATA ===") {
			inMetadata = true
			output = append(output, line)
			continue
		}
		if strings.HasPrefix(line, "# === END_METADATA ===") {
			if !inserted {
				output = append(output, "# ENCRYPTED: true")
				output = append(output, "# RECIPIENTS: "+recipients)
				inserted = true
			}
			inMetadata = false
			output = append(output, line)
			continue
		}
		if inMetadata && (strings.HasPrefix(line, "# ENCRYPTED:") || strings.HasPrefix(line, "# RECIPIENTS:")) {
			continue
		}
		if inMetadata && strings.HasPrefix(line, "# HOST:") {
			output = append(output, "# HOST: "+hostname)
			if !inserted {
				output = append(output, "# ENCRYPTED: true")
				output = append(output, "# RECIPIENTS: "+recipients)
				inserted = true
			}
			continue
		}
		output = append(output, line)
	}

	return os.WriteFile(file, []byte(strings.Join(output, "\n")), 0o600)
}

func ClearEncryptedMetadata(file string) error {
	content, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	lines := strings.Split(string(content), "\n")
	output := make([]string, 0, len(lines))
	for _, line := range lines {
		if strings.HasPrefix(line, "# ENCRYPTED:") || strings.HasPrefix(line, "# RECIPIENTS:") {
			continue
		}
		output = append(output, line)
	}
	return os.WriteFile(file, []byte(strings.Join(output, "\n")), 0o600)
}

func ReplaceMetadataValue(file string, key string, value string) error {
	content, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	lines := strings.Split(string(content), "\n")
	prefix := "# " + key + ":"
	replaced := false
	for i, line := range lines {
		if strings.HasPrefix(line, prefix) {
			lines[i] = "# " + key + ": " + value
			replaced = true
		}
	}
	if !replaced {
		return errors.New("metadata key not found")
	}
	return os.WriteFile(file, []byte(strings.Join(lines, "\n")), 0o600)
}
