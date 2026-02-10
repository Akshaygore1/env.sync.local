package metadata

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"regexp"
	"sort"
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

func EnsureEncryptedMetadata(file string, hostname string) error {
	content, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	lines := strings.Split(string(content), "\n")
	output := make([]string, 0, len(lines)+3)
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
				inserted = true
			}
			inMetadata = false
			output = append(output, line)
			continue
		}
		if inMetadata && (strings.HasPrefix(line, "# ENCRYPTED:") || strings.HasPrefix(line, "# RECIPIENTS:") || strings.HasPrefix(line, "# PUBLIC_KEYS:")) {
			continue
		}
		if inMetadata && strings.HasPrefix(line, "# HOST:") {
			output = append(output, "# HOST: "+hostname)
			if !inserted {
				output = append(output, "# ENCRYPTED: true")
				inserted = true
			}
			continue
		}
		output = append(output, line)
	}

	return os.WriteFile(file, []byte(strings.Join(output, "\n")), 0o600)
}

// EnsurePublicKeysMetadata adds or updates the PUBLIC_KEYS metadata field
func EnsurePublicKeysMetadata(file string, publicKeysMap map[string]string) error {
	if len(publicKeysMap) == 0 {
		return nil
	}

	content, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	// Build the PUBLIC_KEYS value as "hostname1:key1,hostname2:key2,..."
	pairs := make([]string, 0, len(publicKeysMap))
	for hostname, pubkey := range publicKeysMap {
		pairs = append(pairs, hostname+":"+pubkey)
	}
	// Sort for deterministic output
	sort.Strings(pairs)
	publicKeysValue := strings.Join(pairs, ",")

	lines := strings.Split(string(content), "\n")
	output := make([]string, 0, len(lines)+1)
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
				output = append(output, "# PUBLIC_KEYS: "+publicKeysValue)
				inserted = true
			}
			inMetadata = false
			output = append(output, line)
			continue
		}
		if inMetadata && strings.HasPrefix(line, "# PUBLIC_KEYS:") {
			// Skip existing PUBLIC_KEYS line, we'll add new one
			continue
		}
		if inMetadata && strings.HasPrefix(line, "# ENCRYPTED:") {
			// Add PUBLIC_KEYS right after ENCRYPTED
			output = append(output, line)
			if !inserted {
				output = append(output, "# PUBLIC_KEYS: "+publicKeysValue)
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
		if strings.HasPrefix(line, "# ENCRYPTED:") || strings.HasPrefix(line, "# RECIPIENTS:") || strings.HasPrefix(line, "# PUBLIC_KEYS:") {
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
