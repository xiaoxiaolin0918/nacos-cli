package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// ProfileMeta is display/ops metadata stored beside a profile as <name>.meta.yaml.
// Connection fields stay in <name>.conf; this sidecar is ignored by LoadProfile.
type ProfileMeta struct {
	DisplayName string `yaml:"displayName,omitempty" json:"displayName,omitempty"`
	EnvType     string `yaml:"envType,omitempty" json:"envType,omitempty"` // lab|prod (empty ok)
	Color       string `yaml:"color,omitempty" json:"color,omitempty"`
	Owner       string `yaml:"owner,omitempty" json:"owner,omitempty"`
	Notes       string `yaml:"notes,omitempty" json:"notes,omitempty"`
}

func metaPathIn(dir, name string) (string, error) {
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
	return filepath.Join(dir, name+".meta.yaml"), nil
}

// MetaPath returns ~/.nacos-cli/<name>.meta.yaml.
func MetaPath(name string) (string, error) {
	dir, err := ProfileDir()
	if err != nil {
		return "", err
	}
	return metaPathIn(dir, name)
}

func (m *ProfileMeta) normalize() {
	m.DisplayName = strings.TrimSpace(m.DisplayName)
	m.EnvType = strings.ToLower(strings.TrimSpace(m.EnvType))
	m.Color = strings.TrimSpace(m.Color)
	m.Owner = strings.TrimSpace(m.Owner)
	m.Notes = strings.TrimSpace(m.Notes)
}

func (m ProfileMeta) validate() error {
	switch m.EnvType {
	case "", "lab", "prod":
		return nil
	default:
		return fmt.Errorf("invalid envType %q (want lab or prod)", m.EnvType)
	}
}

// LoadMeta loads sidecar metadata for a named profile.
func LoadMeta(name string) (ProfileMeta, bool, error) {
	dir, err := ProfileDir()
	if err != nil {
		return ProfileMeta{}, false, err
	}
	return LoadMetaIn(dir, name)
}

// LoadMetaIn loads <dir>/<name>.meta.yaml.
func LoadMetaIn(dir, name string) (ProfileMeta, bool, error) {
	path, err := metaPathIn(dir, name)
	if err != nil {
		return ProfileMeta{}, false, err
	}
	b, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return ProfileMeta{}, false, nil
		}
		return ProfileMeta{}, false, err
	}
	var m ProfileMeta
	if err := yaml.Unmarshal(b, &m); err != nil {
		return ProfileMeta{}, true, fmt.Errorf("parse %s: %w", path, err)
	}
	m.normalize()
	return m, true, nil
}

// SaveMeta writes ~/.nacos-cli/<name>.meta.yaml (mode 0600).
func SaveMeta(name string, m ProfileMeta) (string, error) {
	dir, err := ProfileDir()
	if err != nil {
		return "", err
	}
	return SaveMetaIn(dir, name, m)
}

// SaveMetaIn writes <dir>/<name>.meta.yaml.
func SaveMetaIn(dir, name string, m ProfileMeta) (string, error) {
	path, err := metaPathIn(dir, name)
	if err != nil {
		return "", err
	}
	m.normalize()
	if err := m.validate(); err != nil {
		return "", err
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	b, err := yaml.Marshal(&m)
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(path, b, 0o600); err != nil {
		return "", err
	}
	_ = os.Chmod(path, 0o600)
	return path, nil
}

// DeleteMeta removes ~/.nacos-cli/<name>.meta.yaml.
func DeleteMeta(name string) error {
	dir, err := ProfileDir()
	if err != nil {
		return err
	}
	return DeleteMetaIn(dir, name)
}

// DeleteMetaIn removes <dir>/<name>.meta.yaml (missing is ok).
func DeleteMetaIn(dir, name string) error {
	path, err := metaPathIn(dir, name)
	if err != nil {
		return err
	}
	err = os.Remove(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}
