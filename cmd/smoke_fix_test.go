package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestConfigExportDryRunWritesNothing(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "sub", "x.yaml")
	root := NewRoot()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"--server", "http://127.0.0.1:1", "config", "export", "--data-id", "x.yaml", "--out", out, "--dry-run"})
	if err := root.Execute(); err != nil {
		t.Fatalf("%v %s", err, buf.String())
	}
	if _, err := os.Stat(out); err == nil {
		t.Fatal("dry-run must not write the export file")
	}
	var env map[string]any
	if err := json.Unmarshal(buf.Bytes(), &env); err != nil {
		t.Fatal(err, buf.String())
	}
	if env["dry_run"] != true {
		t.Fatalf("%v", env)
	}
}

func TestConfigDeleteGate(t *testing.T) {
	root := NewRoot()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"config", "delete", "--data-id", "x.yaml", "--dry-run"})
	if err := root.Execute(); err != nil {
		t.Fatalf("%v", err)
	}
	var env map[string]any
	_ = json.Unmarshal(buf.Bytes(), &env)
	if env["dry_run"] != true {
		t.Fatalf("%s", buf.String())
	}
	root2 := NewRoot()
	root2.SetOut(&bytes.Buffer{})
	root2.SetErr(&bytes.Buffer{})
	root2.SetArgs([]string{"config", "delete", "--data-id", "x.yaml"})
	err := root2.Execute()
	pe, ok := err.(*PrintedError)
	if !ok || pe.ExitCode() != 10 {
		t.Fatalf("want confirmation exit 10, got %T %v", err, err)
	}
}

func TestProfileListEmptyIsArray(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	root := NewRoot()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"profile", "list"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	var env struct {
		Result struct {
			Profiles []string `json:"profiles"`
		} `json:"result"`
	}
	raw := buf.String()
	if err := json.Unmarshal([]byte(raw), &env); err != nil {
		t.Fatal(err)
	}
	if env.Result.Profiles == nil || !bytes.Contains([]byte(raw), []byte(`"profiles":[]`)) {
		t.Fatalf("profiles should be [] not null: %s", raw)
	}
}
