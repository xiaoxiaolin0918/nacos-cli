// Package backup writes local config snapshots under backups/<stamp>/.
package backup

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

// Item is one exported config file in a backup.
type Item struct {
	DataID   string `json:"data_id"`
	Group    string `json:"group"`
	Namespace string `json:"namespace,omitempty"`
	RelPath  string `json:"rel_path"`
	Bytes    int    `json:"bytes"`
}

// Manifest describes a backup directory.
type Manifest struct {
	CreatedAt string `json:"created_at"`
	Server    string `json:"server,omitempty"`
	Namespace string `json:"namespace,omitempty"`
	Group     string `json:"group,omitempty"`
	DataID    string `json:"data_id_filter,omitempty"`
	Items     []Item `json:"items"`
}

// Info is a listed backup folder.
type Info struct {
	Name      string `json:"name"`
	Path      string `json:"path"`
	ItemCount int    `json:"item_count,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
}

var unsafeName = regexp.MustCompile(`[^A-Za-z0-9._@+-]+`)

// Root returns the backups directory (NACOS_BACKUP_DIR or <cwd>/backups).
func Root(cwd string) string {
	if v := strings.TrimSpace(os.Getenv("NACOS_BACKUP_DIR")); v != "" {
		return v
	}
	if cwd == "" {
		cwd, _ = os.Getwd()
	}
	return filepath.Join(cwd, "backups")
}

// SafeName turns dataId into a filesystem-safe single path segment.
func SafeName(dataID string) string {
	s := strings.TrimSpace(dataID)
	if s == "" {
		return "unnamed"
	}
	s = strings.ReplaceAll(s, "/", "__")
	s = strings.ReplaceAll(s, `\`, "__")
	s = unsafeName.ReplaceAllString(s, "_")
	if s == "" || s == "." || s == ".." {
		return "unnamed"
	}
	return s
}

// NewStamp returns a local timestamp folder name.
func NewStamp(now time.Time) string {
	if now.IsZero() {
		now = time.Now()
	}
	return now.Format("20060102-150405")
}

// CreateDir makes backups/<stamp> and returns its absolute path.
func CreateDir(root, stamp string) (string, error) {
	if strings.TrimSpace(stamp) == "" {
		return "", errors.New("stamp is required")
	}
	dir := filepath.Join(root, stamp)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		return dir, nil
	}
	return abs, nil
}

// WriteConfig writes group__dataId.txt (or similar) under dir.
func WriteConfig(dir, dataID, group, content string) (Item, error) {
	if group == "" {
		group = "DEFAULT_GROUP"
	}
	rel := filepath.Join(SafeName(group), SafeName(dataID))
	full := filepath.Join(dir, rel)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		return Item{}, err
	}
	b := []byte(content)
	if err := os.WriteFile(full, b, 0o644); err != nil {
		return Item{}, err
	}
	return Item{
		DataID:  dataID,
		Group:   group,
		RelPath: filepath.ToSlash(rel),
		Bytes:   len(b),
	}, nil
}

// WriteManifest persists manifest.json.
func WriteManifest(dir string, m Manifest) error {
	if m.Items == nil {
		m.Items = []Item{}
	}
	b, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "manifest.json"), append(b, '\n'), 0o644)
}

// List returns backup subdirectories (newest name first when sortable).
func List(root string) ([]Info, error) {
	ents, err := os.ReadDir(root)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []Info{}, nil
		}
		return nil, err
	}
	var out []Info
	for _, e := range ents {
		if !e.IsDir() {
			continue
		}
		info := Info{Name: e.Name(), Path: filepath.Join(root, e.Name())}
		if m, err := ReadManifest(info.Path); err == nil {
			info.ItemCount = len(m.Items)
			info.CreatedAt = m.CreatedAt
		}
		out = append(out, info)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name > out[j].Name })
	return out, nil
}

// ReadManifest loads manifest.json if present.
func ReadManifest(dir string) (Manifest, error) {
	b, err := os.ReadFile(filepath.Join(dir, "manifest.json"))
	if err != nil {
		return Manifest{}, err
	}
	var m Manifest
	if err := json.Unmarshal(b, &m); err != nil {
		return Manifest{}, err
	}
	return m, nil
}

// EnsureWritable creates root if needed and checks write access.
func EnsureWritable(root string) error {
	if err := os.MkdirAll(root, 0o755); err != nil {
		return fmt.Errorf("backup root: %w", err)
	}
	probe := filepath.Join(root, ".write-probe")
	if err := os.WriteFile(probe, []byte("ok"), 0o644); err != nil {
		return fmt.Errorf("backup root not writable: %w", err)
	}
	_ = os.Remove(probe)
	return nil
}
