package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// Profile is one named connection preset under ~/.nacos-cli/<name>.conf (YAML).
// Passwords may live here for local convenience; never commit profile files to git.
type Profile struct {
	Server      string `yaml:"server,omitempty" json:"server,omitempty"`
	Host        string `yaml:"host,omitempty" json:"host,omitempty"`
	Port        int    `yaml:"port,omitempty" json:"port,omitempty"`
	AuthType    string `yaml:"authType,omitempty" json:"authType,omitempty"`
	Username    string `yaml:"username,omitempty" json:"username,omitempty"`
	Password    string `yaml:"password,omitempty" json:"password,omitempty"`
	Namespace   string `yaml:"namespace,omitempty" json:"namespace,omitempty"`
	ContextPath string `yaml:"contextPath,omitempty" json:"contextPath,omitempty"`
	Format      string `yaml:"format,omitempty" json:"format,omitempty"`
}

var profileNameRE = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.-]*$`)

// ProfileDir returns ~/.nacos-cli (official CLI layout; not the same as ConfigDir).
func ProfileDir() (string, error) {
	return profileDirFor(runtime.GOOS, userHome())
}

func profileDirFor(goos, home string) (string, error) {
	_ = goos
	if home == "" {
		return "", errors.New("home directory is unknown")
	}
	return filepath.Join(home, ".nacos-cli"), nil
}

// ProfilePath returns ~/.nacos-cli/<name>.conf.
func ProfilePath(name string) (string, error) {
	dir, err := ProfileDir()
	if err != nil {
		return "", err
	}
	return profilePathIn(dir, name)
}

func profilePathIn(dir, name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		name = "default"
	}
	if !profileNameRE.MatchString(name) {
		return "", fmt.Errorf("invalid profile name %q (use letters, digits, _ . -)", name)
	}
	cleaned := filepath.Clean(name)
	if cleaned != name || strings.Contains(name, "..") || strings.ContainsAny(name, `/\`) {
		return "", fmt.Errorf("invalid profile name %q", name)
	}
	return filepath.Join(dir, name+".conf"), nil
}

// LoadProfile loads a named profile from ProfileDir.
func LoadProfile(name string) (Profile, bool, error) {
	dir, err := ProfileDir()
	if err != nil {
		return Profile{}, false, err
	}
	return LoadProfileIn(dir, name)
}

// LoadProfileIn loads <dir>/<name>.conf.
func LoadProfileIn(dir, name string) (Profile, bool, error) {
	path, err := profilePathIn(dir, name)
	if err != nil {
		return Profile{}, false, err
	}
	return loadProfileFile(path)
}

// LoadProfileFile loads an explicit --config path.
func LoadProfileFile(path string) (Profile, bool, error) {
	return loadProfileFile(path)
}

func loadProfileFile(path string) (Profile, bool, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Profile{}, false, nil
		}
		return Profile{}, false, err
	}
	var p Profile
	if err := yaml.Unmarshal(b, &p); err != nil {
		return Profile{}, true, fmt.Errorf("parse %s: %w", path, err)
	}
	p.normalize()
	return p, true, nil
}

func (p *Profile) normalize() {
	p.Server = strings.TrimSpace(p.Server)
	p.Host = strings.TrimSpace(p.Host)
	p.Username = strings.TrimSpace(p.Username)
	p.Namespace = strings.TrimSpace(p.Namespace)
	p.ContextPath = strings.TrimSpace(p.ContextPath)
	p.AuthType = strings.TrimSpace(p.AuthType)
	p.Format = strings.TrimSpace(p.Format)
	if p.Server == "" && p.Host != "" {
		port := p.Port
		if port <= 0 {
			port = 8848
		}
		p.Server = "http://" + p.Host + ":" + strconv.Itoa(port)
	}
	p.Server = strings.TrimRight(p.Server, "/")
}

// SaveProfile writes ~/.nacos-cli/<name>.conf (mode 0600).
func SaveProfile(name string, p Profile) (string, error) {
	dir, err := ProfileDir()
	if err != nil {
		return "", err
	}
	return SaveProfileIn(dir, name, p)
}

// SaveProfileIn writes <dir>/<name>.conf.
func SaveProfileIn(dir, name string, p Profile) (string, error) {
	path, err := profilePathIn(dir, name)
	if err != nil {
		return "", err
	}
	p.normalize()
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	b, err := yaml.Marshal(&p)
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(path, b, 0o600); err != nil {
		return "", err
	}
	_ = os.Chmod(path, 0o600)
	return path, nil
}

// ListProfiles lists profile names in ProfileDir (sorted).
func ListProfiles() ([]string, error) {
	dir, err := ProfileDir()
	if err != nil {
		return nil, err
	}
	return ListProfilesIn(dir)
}

// ListProfilesIn lists *.conf basenames without extension.
func ListProfilesIn(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []string{}, nil
		}
		return nil, err
	}
	names := []string{}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(name, ".conf") {
			continue
		}
		base := strings.TrimSuffix(name, ".conf")
		if profileNameRE.MatchString(base) {
			names = append(names, base)
		}
	}
	sort.Strings(names)
	return names, nil
}

// DeleteProfileIn removes <dir>/<name>.conf.
func DeleteProfileIn(dir, name string) error {
	path, err := profilePathIn(dir, name)
	if err != nil {
		return err
	}
	err = os.Remove(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}

// DeleteProfile removes a named profile from ProfileDir.
func DeleteProfile(name string) error {
	dir, err := ProfileDir()
	if err != nil {
		return err
	}
	return DeleteProfileIn(dir, name)
}

// MergeSettings overlays non-empty profile fields onto base (env/defaults).
func MergeSettings(base Settings, p Profile) Settings {
	out := base
	if strings.TrimSpace(p.Server) != "" {
		out.Server = strings.TrimRight(strings.TrimSpace(p.Server), "/")
	}
	if strings.TrimSpace(p.Username) != "" {
		out.Username = strings.TrimSpace(p.Username)
	}
	if p.Password != "" {
		out.Password = p.Password
	}
	if strings.TrimSpace(p.Namespace) != "" {
		out.Namespace = NormalizeNamespaceID(p.Namespace)
	}
	if strings.TrimSpace(p.ContextPath) != "" {
		out.ContextPath = strings.TrimSpace(p.ContextPath)
	}
	return out
}

// SecretKey reports keys that must never be printed by profile get.
func SecretKey(key string) bool {
	switch strings.ToLower(strings.TrimSpace(key)) {
	case "password", "login_password", "secret", "token", "accesskey", "secretkey", "access_key", "secret_key":
		return true
	default:
		return false
	}
}

// NormalizeFormat accepts json|pretty (case-insensitive). Empty -> json.
func NormalizeFormat(v string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "", "json":
		return "json", nil
	case "pretty":
		return "pretty", nil
	default:
		return "", fmt.Errorf("unknown format %q (want json or pretty)", v)
	}
}

// RedactedView returns a map safe for profile show (password masked).
func (p Profile) RedactedView() map[string]any {
	pw := ""
	if p.Password != "" {
		pw = "******"
	}
	return map[string]any{
		"server":       p.Server,
		"host":         p.Host,
		"port":         p.Port,
		"authType":     p.AuthType,
		"username":     p.Username,
		"password":     pw,
		"namespace":    p.Namespace,
		"contextPath":  p.ContextPath,
		"format":       p.Format,
		"password_set": p.Password != "",
	}
}
