package cmd

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestWriteGateDryRunAndYes(t *testing.T) {
	root := NewRoot()
	buf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(errBuf)

	root.SetArgs([]string{"config", "publish", "--data-id", "x.yaml", "--content", "k: v", "--dry-run"})
	if err := root.Execute(); err != nil {
		t.Fatalf("dry-run: %v", err)
	}
	var env map[string]any
	if err := json.Unmarshal(buf.Bytes(), &env); err != nil {
		t.Fatal(err, buf.String())
	}
	if env["ok"] != true || env["dry_run"] != true {
		t.Fatalf("%v", env)
	}

	buf.Reset()
	errBuf.Reset()
	root2 := NewRoot()
	root2.SetOut(buf)
	root2.SetErr(errBuf)
	root2.SetArgs([]string{"config", "publish", "--data-id", "x.yaml", "--content", "k: v"})
	err := root2.Execute()
	if err == nil {
		t.Fatal("expected confirmation error")
	}
	pe, ok := err.(*PrintedError)
	if !ok || pe.ExitCode() != 10 {
		t.Fatalf("%T %v", err, err)
	}
}

func TestACLUserCreateGateMasksPassword(t *testing.T) {
	root := NewRoot()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"user", "create", "--username", "alice", "--password", "super-secret", "--dry-run"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	raw := buf.String()
	if raw == "" {
		t.Fatal("empty")
	}
	if contains(raw, "super-secret") {
		t.Fatal("password leaked in dry-run output")
	}
	var env map[string]any
	if err := json.Unmarshal(buf.Bytes(), &env); err != nil {
		t.Fatal(err, raw)
	}
	if env["dry_run"] != true {
		t.Fatalf("%v", env)
	}
	would, _ := env["would"].(map[string]any)
	if would["username"] != "alice" {
		t.Fatalf("%v", would)
	}
	if _, ok := would["password"]; ok {
		t.Fatal("password key must not appear")
	}
}

func TestACLPermissionGrantRequiresYes(t *testing.T) {
	root := NewRoot()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"permission", "grant", "--role", "dev", "--resource", "*:*:*", "--action", "r"})
	err := root.Execute()
	if err == nil {
		t.Fatal("expected confirmation")
	}
	pe, ok := err.(*PrintedError)
	if !ok || pe.ExitCode() != 10 {
		t.Fatalf("%T %v", err, err)
	}
}

func contains(s, sub string) bool {
	return len(sub) == 0 || (len(s) >= len(sub) && (s == sub || len(s) > 0 && (func() bool {
		for i := 0; i+len(sub) <= len(s); i++ {
			if s[i:i+len(sub)] == sub {
				return true
			}
		}
		return false
	})()))
}
