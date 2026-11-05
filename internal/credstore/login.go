package credstore

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	loginFileName = "login.cred"
	loginVersion  = 1
)

// LoginPassword is the decrypted in-memory remembered password. Never log Password.
type LoginPassword struct {
	Username  string
	Password  string
	AutoLogin bool
}

type loginDiskFile struct {
	Version           int    `json:"version"`
	Username          string `json:"username"`
	PasswordProtected string `json:"password_protected"`
	Protect           string `json:"protect"`
	AutoLogin         bool   `json:"auto_login"`
}

// LoginPath returns the encrypted login.cred path under configDir.
func LoginPath(configDir string) string {
	return filepath.Join(configDir, loginFileName)
}

// DirFromCredentialsPath returns the directory that holds credentials.json (and login.cred).
func DirFromCredentialsPath(credPath string) string {
	if credPath == "" {
		return ""
	}
	return filepath.Dir(credPath)
}

// SaveLogin encrypts the password and writes login.cred with mode 0600.
func SaveLogin(configDir string, c LoginPassword) error {
	if strings.TrimSpace(configDir) == "" {
		return errors.New("credstore: config dir is empty")
	}
	user := strings.TrimSpace(c.Username)
	if user == "" {
		return errors.New("credstore: username is required")
	}
	if c.Password == "" {
		return errors.New("credstore: password is required")
	}
	if err := os.MkdirAll(configDir, 0o700); err != nil {
		return err
	}
	blob, scheme, err := protect([]byte(c.Password), configDir)
	if err != nil {
		return err
	}
	df := loginDiskFile{
		Version:           loginVersion,
		Username:          user,
		PasswordProtected: base64.StdEncoding.EncodeToString(blob),
		Protect:           scheme,
		AutoLogin:         c.AutoLogin,
	}
	b, err := json.MarshalIndent(df, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n')
	path := LoginPath(configDir)
	if err := os.WriteFile(path, b, 0o600); err != nil {
		return err
	}
	return os.Chmod(path, 0o600)
}

// LoadLogin decrypts login.cred. Missing file returns ok=false and no error.
func LoadLogin(configDir string) (LoginPassword, bool, error) {
	path := LoginPath(configDir)
	b, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return LoginPassword{}, false, nil
		}
		return LoginPassword{}, false, err
	}
	var df loginDiskFile
	if err := json.Unmarshal(b, &df); err != nil {
		return LoginPassword{}, true, fmt.Errorf("credstore: parse %s: %w", path, err)
	}
	if df.Version != 0 && df.Version != loginVersion {
		return LoginPassword{}, true, fmt.Errorf("credstore: unsupported login.cred version %d", df.Version)
	}
	raw, err := base64.StdEncoding.DecodeString(df.PasswordProtected)
	if err != nil {
		return LoginPassword{}, true, fmt.Errorf("credstore: corrupt ciphertext: %w", err)
	}
	plain, err := unprotect(raw, df.Protect, configDir)
	if err != nil {
		return LoginPassword{}, true, err
	}
	return LoginPassword{
		Username:  strings.TrimSpace(df.Username),
		Password:  string(plain),
		AutoLogin: df.AutoLogin,
	}, true, nil
}

// ClearLogin removes login.cred and login.key when present. Missing is OK.
func ClearLogin(configDir string) error {
	if strings.TrimSpace(configDir) == "" {
		return nil
	}
	_ = os.Remove(localKeyPath(configDir))
	err := os.Remove(LoginPath(configDir))
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

// PeekLogin reads username and auto_login without decrypting the password.
func PeekLogin(configDir string) (username string, autoLogin bool, ok bool) {
	df, exists, err := peekLogin(configDir)
	if err != nil || !exists {
		return "", false, false
	}
	return strings.TrimSpace(df.Username), df.AutoLogin, true
}

// HasLogin reports whether login.cred exists (password presence; never decrypts).
func HasLogin(configDir string) bool {
	_, err := os.Stat(LoginPath(configDir))
	return err == nil
}

func peekLogin(configDir string) (loginDiskFile, bool, error) {
	b, err := os.ReadFile(LoginPath(configDir))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return loginDiskFile{}, false, nil
		}
		return loginDiskFile{}, false, err
	}
	var df loginDiskFile
	if err := json.Unmarshal(b, &df); err != nil {
		return loginDiskFile{}, true, err
	}
	return df, true, nil
}

// ProfileLoginDir returns <profileDir>/.login/<name> for per-profile encrypted passwords.
func ProfileLoginDir(profileDir, name string) string {
	return filepath.Join(profileDir, ".login", strings.TrimSpace(name))
}
