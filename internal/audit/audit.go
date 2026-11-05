package audit

import (
	"bufio"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/xiaoxiaolin0918/nacos-cli/internal/config"
)

// Entry is one JSONL audit record. Never store password or token values here.
type Entry struct {
	Time    string `json:"time"`
	Command string `json:"command"`
	Risk    string `json:"risk,omitempty"`
	Profile string `json:"profile,omitempty"`
	OK      bool   `json:"ok"`
	Error   string `json:"error,omitempty"`
}

// Filter selects audit rows for the local web UI.
type Filter struct {
	CommandContains string
	Profile         string
	Since           *time.Time
	Limit           int
}

// Store appends/queries a JSONL audit file.
type Store struct {
	Path string
	mu   sync.Mutex
}

// DefaultPath returns ~/.nacos-cli/audit.jsonl.
func DefaultPath() (string, error) {
	dir, err := config.ProfileDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "audit.jsonl"), nil
}

// Default returns a Store pointing at DefaultPath (best-effort).
func Default() (*Store, error) {
	p, err := DefaultPath()
	if err != nil {
		return nil, err
	}
	return &Store{Path: p}, nil
}

// Append writes one JSON line. Creates parent dirs as needed.
func (s *Store) Append(e Entry) error {
	if s == nil || strings.TrimSpace(s.Path) == "" {
		return errors.New("audit store path empty")
	}
	if strings.TrimSpace(e.Time) == "" {
		e.Time = time.Now().UTC().Format(time.RFC3339)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := os.MkdirAll(filepath.Dir(s.Path), 0o700); err != nil {
		return err
	}
	f, err := os.OpenFile(s.Path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	defer f.Close()
	b, err := json.Marshal(e)
	if err != nil {
		return err
	}
	if _, err := f.Write(append(b, '\n')); err != nil {
		return err
	}
	_ = os.Chmod(s.Path, 0o600)
	return nil
}

// Query returns matching entries newest-last (file order). Limit<=0 means no limit.
func (s *Store) Query(f Filter) ([]Entry, error) {
	if s == nil || strings.TrimSpace(s.Path) == "" {
		return nil, errors.New("audit store path empty")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	file, err := os.Open(s.Path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []Entry{}, nil
		}
		return nil, err
	}
	defer file.Close()

	cmdSub := strings.ToLower(strings.TrimSpace(f.CommandContains))
	prof := strings.TrimSpace(f.Profile)
	out := []Entry{}
	sc := bufio.NewScanner(file)
	// allow long lines
	buf := make([]byte, 0, 64*1024)
	sc.Buffer(buf, 1024*1024)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		var e Entry
		if err := json.Unmarshal([]byte(line), &e); err != nil {
			continue
		}
		if cmdSub != "" && !strings.Contains(strings.ToLower(e.Command), cmdSub) {
			continue
		}
		if prof != "" && e.Profile != prof {
			continue
		}
		if f.Since != nil {
			ts, err := time.Parse(time.RFC3339, e.Time)
			if err != nil {
				continue
			}
			if ts.Before(*f.Since) {
				continue
			}
		}
		out = append(out, e)
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	if f.Limit > 0 && len(out) > f.Limit {
		out = out[len(out)-f.Limit:]
	}
	return out, nil
}

// RedactError masks known secrets in an error message.
func RedactError(msg string, secrets []string) string {
	s := msg
	for _, sec := range secrets {
		if len(sec) < 4 {
			continue
		}
		s = strings.ReplaceAll(s, sec, "***")
	}
	return s
}
