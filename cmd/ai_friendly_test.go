package cmd

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestDryRunWouldIsArray(t *testing.T) {
	root := NewRoot()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"config", "publish", "--data-id", "x.yaml", "--content", "k: v", "--dry-run"})
	if err := root.Execute(); err != nil {
		t.Fatalf("dry-run: %v", err)
	}
	var env map[string]any
	if err := json.Unmarshal(buf.Bytes(), &env); err != nil {
		t.Fatal(err, buf.String())
	}
	would, ok := env["would"].([]any)
	if !ok || len(would) < 1 {
		t.Fatalf("expected would[], got %T %v", env["would"], env["would"])
	}
	item, _ := would[0].(map[string]any)
	if item["action"] != "config.publish" {
		t.Fatalf("%v", item)
	}
}

func TestExamplesCommand(t *testing.T) {
	root := NewRoot()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"examples", "--command", "publish"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	var env map[string]any
	if err := json.Unmarshal(buf.Bytes(), &env); err != nil {
		t.Fatal(err, buf.String())
	}
	if env["ok"] != true {
		t.Fatalf("%v", env)
	}
	res, _ := env["result"].(map[string]any)
	samples, _ := res["samples"].([]any)
	if len(samples) < 1 {
		t.Fatalf("samples=%v", res)
	}
	if res["capabilities_schema"] == nil {
		t.Fatalf("missing capabilities_schema: %v", res)
	}
}

func TestTraceMeta(t *testing.T) {
	root := NewRoot()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"--trace", "commands"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	var env map[string]any
	if err := json.Unmarshal(buf.Bytes(), &env); err != nil {
		t.Fatal(err, buf.String())
	}
	meta, _ := env["meta"].(map[string]any)
	if meta == nil || meta["trace"] != true {
		t.Fatalf("meta=%v env=%v", meta, env)
	}
	if _, ok := meta["duration_ms"]; !ok {
		t.Fatalf("duration_ms missing: %v", meta)
	}
}

func TestAuthLoginNonTTYMissingPasswordCoded(t *testing.T) {
	root := NewRoot()
	buf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(errBuf)
	// Non-interactive stdin (bytes.Buffer), username set, no password.
	root.SetIn(bytes.NewBuffer(nil))
	root.SetArgs([]string{"auth", "login", "--server", "http://127.0.0.1:9", "--username", "nacos"})
	err := root.Execute()
	if err == nil {
		t.Fatal("expected error")
	}
	raw := buf.String() + errBuf.String()
	if !strings.Contains(raw, "password_required") && !strings.Contains(raw, "password required") {
		t.Fatalf("expected password_required code/message, got: %s", raw)
	}
	var env map[string]any
	_ = json.Unmarshal([]byte(raw), &env)
	if env == nil {
		// may be on stderr as envelope
		for _, line := range strings.Split(raw, "\n") {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "{") {
				_ = json.Unmarshal([]byte(line), &env)
				break
			}
		}
	}
	if errObj, ok := env["error"].(map[string]any); ok {
		if errObj["code"] != "password_required" {
			t.Fatalf("code=%v env=%v", errObj["code"], env)
		}
	}
}

func TestCommandsPayloadHasCapabilities(t *testing.T) {
	root := NewRoot()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"commands"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	var env map[string]any
	if err := json.Unmarshal(buf.Bytes(), &env); err != nil {
		t.Fatal(err)
	}
	res, _ := env["result"].(map[string]any)
	if res["capabilities_schema"] == nil {
		t.Fatalf("%v", res)
	}
	if res["would"] != "array" {
		t.Fatalf("would shape=%v", res["would"])
	}
	flags, _ := res["global_flags"].([]any)
	joined := ""
	for _, f := range flags {
		joined += f.(string) + " "
	}
	if !strings.Contains(joined, "--trace") {
		t.Fatalf("global_flags missing --trace: %v", flags)
	}
}
