// Package config loads settings from flags, env, and optional .env file.
package config

import (
	"bufio"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// Settings are resolved connection parameters. Password is never logged.
type Settings struct {
	Server      string
	Username    string
	Password    string
	Namespace   string
	ContextPath string
}

func (s Settings) BaseURL() string {
	server := strings.TrimRight(strings.TrimSpace(s.Server), "/")
	ctx := strings.TrimSpace(s.ContextPath)
	if ctx == "" {
		ctx = "/nacos"
	}
	if !strings.HasPrefix(ctx, "/") {
		ctx = "/" + ctx
	}
	return server + strings.TrimRight(ctx, "/")
}

// LoadDotEnv loads KEY=VAL from path into the process env when the key is unset.
func LoadDotEnv(path string) error {
	f, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") || !strings.Contains(line, "=") {
			continue
		}
		key, val, _ := strings.Cut(line, "=")
		key = strings.TrimSpace(key)
		val = strings.TrimSpace(val)
		val = strings.Trim(val, `"'`)
		if key == "" {
			continue
		}
		if _, ok := os.LookupEnv(key); !ok {
			_ = os.Setenv(key, val)
		}
	}
	return sc.Err()
}

// FromEnv builds Settings from environment (after optional LoadDotEnv).
func FromEnv() Settings {
	ctx := strings.TrimSpace(os.Getenv("NACOS_CONTEXT_PATH"))
	if ctx == "" {
		ctx = "/nacos"
	}
	user := strings.TrimSpace(os.Getenv("NACOS_USERNAME"))
	if user == "" {
		user = "nacos"
	}
	return Settings{
		Server:      strings.TrimSpace(os.Getenv("NACOS_SERVER")),
		Username:    user,
		Password:    os.Getenv("NACOS_PASSWORD"),
		Namespace:   strings.TrimSpace(os.Getenv("NACOS_NAMESPACE")),
		ContextPath: ctx,
	}
}

// ConfigDir returns the nacos-cli config directory.
func ConfigDir() (string, error) {
	return dirFor(runtime.GOOS, os.Getenv("APPDATA"), os.Getenv("XDG_CONFIG_HOME"), userHome())
}

func userHome() string {
	h, _ := os.UserHomeDir()
	return h
}

func dirFor(goos, appdata, xdg, home string) (string, error) {
	if goos == "windows" && appdata != "" {
		return filepath.Join(appdata, "nacos-cli"), nil
	}
	if xdg != "" {
		return filepath.Join(xdg, "nacos-cli"), nil
	}
	if home == "" {
		return "", errors.New("home directory is unknown")
	}
	return filepath.Join(home, ".config", "nacos-cli"), nil
}

// CredentialsPath returns the default credentials.json path (override NACOS_CREDENTIALS_PATH).
func CredentialsPath() (string, error) {
	if v := strings.TrimSpace(os.Getenv("NACOS_CREDENTIALS_PATH")); v != "" {
		return v, nil
	}
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "credentials.json"), nil
}

// FindDotEnv looks for .env beside the executable cwd and common parents.
func FindDotEnv() string {
	candidates := []string{".env"}
	if wd, err := os.Getwd(); err == nil {
		candidates = append(candidates, filepath.Join(wd, ".env"))
	}
	for _, c := range candidates {
		if st, err := os.Stat(c); err == nil && !st.IsDir() {
			return c
		}
	}
	return ".env"
}
