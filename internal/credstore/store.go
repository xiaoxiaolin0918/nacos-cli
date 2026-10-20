// Package credstore persists Nacos access tokens (never passwords).
package credstore

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

// Cred is the on-disk credential. Password is intentionally absent.
type Cred struct {
	Server      string `json:"server,omitempty"`
	Username    string `json:"username,omitempty"`
	AccessToken string `json:"access_token,omitempty"`
	AuthMode    string `json:"auth_mode,omitempty"`
	Via         string `json:"via,omitempty"`
}

func Load(path string) (Cred, bool, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Cred{}, false, nil
		}
		return Cred{}, false, err
	}
	var c Cred
	if err := json.Unmarshal(b, &c); err != nil {
		return Cred{}, true, err
	}
	return c, true, nil
}

func Save(path string, c Cred) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	b, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n')
	if err := os.WriteFile(path, b, 0o600); err != nil {
		return err
	}
	return os.Chmod(path, 0o600)
}

func Clear(path string) error {
	err := os.Remove(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

func MaskToken(token string) string {
	token = strings.TrimSpace(token)
	if token == "" || token == "AUTH_DISABLED" {
		return ""
	}
	if len(token) <= 8 {
		return "***"
	}
	return token[:4] + "…" + token[len(token)-4:]
}

func Status(path string) map[string]any {
	c, _, _ := Load(path)
	has := c.AccessToken != "" && c.AccessToken != "AUTH_DISABLED"
	return map[string]any{
		"credentials_path": path,
		"has_token":        has,
		"token_masked":     MaskToken(c.AccessToken),
		"username":         c.Username,
		"server":           c.Server,
		"auth_mode":        firstNonEmpty(c.AuthMode, "none"),
		"via":              c.Via,
	}
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}
