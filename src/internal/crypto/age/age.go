package age

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	agecrypto "filippo.io/age"

	"envsync/internal/config"
	"envsync/internal/logging"
)

func CheckInstalled() error {
	return nil
}

func GenerateKey() error {
	if err := os.MkdirAll(config.AgeKeyDir(), 0o700); err != nil {
		return err
	}
	if _, err := os.Stat(config.AgeKeyFile()); err == nil {
		logging.Log("WARN", "AGE key already exists at "+config.AgeKeyFile())
		return errors.New("key exists")
	}

	identity, err := agecrypto.GenerateX25519Identity()
	if err != nil {
		return err
	}
	recipient := identity.Recipient().String()
	keyData := fmt.Sprintf("# created: %s\n# public key: %s\n%s\n", time.Now().UTC().Format(time.RFC3339), recipient, identity.String())
	if err := os.WriteFile(config.AgeKeyFile(), []byte(keyData), 0o600); err != nil {
		return err
	}
	if err := os.WriteFile(config.AgePubKeyFile(), []byte(recipient), 0o644); err != nil {
		return err
	}
	logging.Log("SUCCESS", "Generated AGE key pair")
	logging.Log("INFO", "Public key: "+recipient)
	return nil
}

func GetLocalPubkey() string {
	data, err := os.ReadFile(config.AgePubKeyFile())
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func EncryptValue(value string, recipients []string) (string, error) {
	if len(recipients) == 0 {
		logging.Log("ERROR", "No recipients specified for encryption")
		return "", errors.New("no recipients")
	}

	parsedRecipients := make([]agecrypto.Recipient, 0, len(recipients))
	for _, recipient := range recipients {
		parsed, err := agecrypto.ParseX25519Recipient(recipient)
		if err != nil {
			logging.Log("ERROR", "Invalid recipient: "+recipient)
			return "", err
		}
		parsedRecipients = append(parsedRecipients, parsed)
	}

	var out bytes.Buffer
	enc, err := agecrypto.Encrypt(&out, parsedRecipients...)
	if err != nil {
		logging.Log("ERROR", "Encryption failed")
		return "", err
	}
	if _, err := io.WriteString(enc, value); err != nil {
		return "", err
	}
	if err := enc.Close(); err != nil {
		return "", err
	}

	encoded := base64.StdEncoding.EncodeToString(out.Bytes())
	return encoded, nil
}

func DecryptValue(encrypted string) (string, error) {
	if strings.TrimSpace(encrypted) == "" {
		return "", nil
	}
	if _, err := os.Stat(config.AgeKeyFile()); err != nil {
		logging.Log("ERROR", "No private key found for decryption")
		return "", err
	}

	decoded, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return "", err
	}

	return DecryptBytes(decoded)
}

func DecryptBytes(encrypted []byte) (string, error) {
	identities, err := loadIdentities()
	if err != nil {
		return "", err
	}

	dec, err := agecrypto.Decrypt(bytes.NewReader(encrypted), identities...)
	if err != nil {
		return "", err
	}
	data, err := io.ReadAll(dec)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

func loadIdentities() ([]agecrypto.Identity, error) {
	data, err := os.ReadFile(config.AgeKeyFile())
	if err != nil {
		return nil, err
	}
	identities, err := agecrypto.ParseIdentities(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	if len(identities) == 0 {
		return nil, errors.New("no identities found")
	}
	return identities, nil
}
