package age

import (
    "bytes"
    "encoding/base64"
    "errors"
    "os"
    "os/exec"
    "strings"

    "envsync/internal/config"
    "envsync/internal/logging"
)

func CheckInstalled() error {
    if _, err := exec.LookPath("age"); err != nil {
        logging.Log("ERROR", "age is not installed. Install with: brew install age (macOS) or apt install age (Linux)")
        return err
    }
    if _, err := exec.LookPath("age-keygen"); err != nil {
        logging.Log("ERROR", "age-keygen is not installed. Install age package.")
        return err
    }
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

    cmd := exec.Command("age-keygen", "-o", config.AgeKeyFile())
    cmd.Stdout = nil
    cmd.Stderr = nil
    if err := cmd.Run(); err != nil {
        return err
    }
    _ = os.Chmod(config.AgeKeyFile(), 0o600)

    pub, err := exec.Command("age-keygen", "-y", config.AgeKeyFile()).Output()
    if err != nil {
        return err
    }
    if err := os.WriteFile(config.AgePubKeyFile(), bytes.TrimSpace(pub), 0o644); err != nil {
        return err
    }
    logging.Log("SUCCESS", "Generated AGE key pair")
    logging.Log("INFO", "Public key: "+strings.TrimSpace(string(pub)))
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

    args := []string{}
    for _, recipient := range recipients {
        args = append(args, "-r", recipient)
    }

    cmd := exec.Command("age", args...)
    cmd.Stdin = strings.NewReader(value)
    var out bytes.Buffer
    cmd.Stdout = &out
    cmd.Stderr = nil
    if err := cmd.Run(); err != nil {
        logging.Log("ERROR", "Encryption failed")
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

    cmd := exec.Command("age", "-d", "-i", config.AgeKeyFile())
    cmd.Stdin = bytes.NewReader(decoded)
    var out bytes.Buffer
    cmd.Stdout = &out
    if err := cmd.Run(); err != nil {
        return "", err
    }
    return strings.TrimSpace(out.String()), nil
}

func DecryptBytes(encrypted []byte) (string, error) {
    cmd := exec.Command("age", "-d", "-i", config.AgeKeyFile())
    cmd.Stdin = bytes.NewReader(encrypted)
    var out bytes.Buffer
    cmd.Stdout = &out
    if err := cmd.Run(); err != nil {
        return "", err
    }
    return strings.TrimSpace(out.String()), nil
}
