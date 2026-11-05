package audit

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestAppendAndQuery(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "audit.jsonl")
	s := Store{Path: path}
	e1 := Entry{
		Time:    time.Date(2026, 10, 29, 1, 0, 0, 0, time.UTC).Format(time.RFC3339),
		Command: "nacos config get",
		Risk:    "read",
		Profile: "lab",
		OK:      true,
	}
	e2 := Entry{
		Time:    time.Date(2026, 10, 29, 2, 0, 0, 0, time.UTC).Format(time.RFC3339),
		Command: "nacos profile delete",
		Risk:    "high-risk-write",
		Profile: "prod",
		OK:      false,
		Error:   "refused",
	}
	if err := s.Append(e1); err != nil {
		t.Fatal(err)
	}
	if err := s.Append(e2); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(string(raw)), "\n")
	if len(lines) != 2 {
		t.Fatalf("lines=%d body=%s", len(lines), raw)
	}

	all, err := s.Query(Filter{})
	if err != nil || len(all) != 2 {
		t.Fatalf("all=%d err=%v", len(all), err)
	}
	byCmd, err := s.Query(Filter{CommandContains: "config get"})
	if err != nil || len(byCmd) != 1 || byCmd[0].Command != e1.Command {
		t.Fatalf("%+v err=%v", byCmd, err)
	}
	byProf, err := s.Query(Filter{Profile: "prod"})
	if err != nil || len(byProf) != 1 {
		t.Fatalf("%+v", byProf)
	}
	since := time.Date(2026, 10, 29, 1, 30, 0, 0, time.UTC)
	byTime, err := s.Query(Filter{Since: &since})
	if err != nil || len(byTime) != 1 || byTime[0].Profile != "prod" {
		t.Fatalf("%+v err=%v", byTime, err)
	}
}

func TestAppendCreatesDir(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nested", "audit.jsonl")
	s := Store{Path: path}
	if err := s.Append(Entry{Time: time.Now().UTC().Format(time.RFC3339), Command: "nacos whoami", OK: true}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatal(err)
	}
}

func TestQueryMissingFile(t *testing.T) {
	s := Store{Path: filepath.Join(t.TempDir(), "missing.jsonl")}
	got, err := s.Query(Filter{})
	if err != nil || len(got) != 0 {
		t.Fatalf("got=%v err=%v", got, err)
	}
}

func TestRedactErrorMessage(t *testing.T) {
	msg := RedactError("login failed token=abc12345 password=secret", []string{"abc12345", "secret"})
	if strings.Contains(msg, "abc12345") || strings.Contains(msg, "secret") {
		t.Fatalf("not redacted: %s", msg)
	}
	if !strings.Contains(msg, "***") {
		t.Fatalf("expected mask: %s", msg)
	}
}

func TestDefaultPath(t *testing.T) {
	p, err := DefaultPath()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(p, "audit.jsonl") {
		t.Fatalf("%s", p)
	}
}
