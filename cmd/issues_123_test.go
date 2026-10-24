package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xiaoxiaolin0918/nacos-cli/internal/config"
)

func TestPublishDryRunInfersTypeAndEffectivePublicNS(t *testing.T) {
	home := t.TempDir()
	t.Setenv("USERPROFILE", home)
	t.Setenv("HOME", home)
	t.Setenv("NACOS_SERVER", "")
	t.Setenv("NACOS_NAMESPACE", "")

	name := "lab-public"
	root := NewRoot()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{
		"profile", "set", name,
		"--server", "http://127.0.0.1:8848",
		"--namespace", "public",
	})
	if err := root.Execute(); err != nil {
		t.Fatal(err, buf.String())
	}

	tmp := t.TempDir()
	file := filepath.Join(tmp, "app.yaml")
	if err := os.WriteFile(file, []byte("a: 1\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	buf.Reset()
	root2 := NewRoot()
	root2.SetOut(buf)
	root2.SetErr(&bytes.Buffer{})
	root2.SetArgs([]string{
		"--profile", name,
		"config", "publish",
		"--data-id", "app.yaml",
		"--group", "DEFAULT_GROUP",
		"--file", file,
		"--dry-run",
	})
	if err := root2.Execute(); err != nil {
		t.Fatal(err, buf.String())
	}
	var env map[string]any
	if err := json.Unmarshal(buf.Bytes(), &env); err != nil {
		t.Fatal(err, buf.String())
	}
	if env["dry_run"] != true {
		t.Fatalf("%s", buf.String())
	}
	would := env["would"].(map[string]any)
	if would["type"] != "yaml" {
		t.Fatalf("expected inferred yaml, got %v", would["type"])
	}
	if ns, _ := would["namespace"].(string); ns != "" {
		t.Fatalf("namespace=%q want empty", ns)
	}
	if eff, _ := would["effective_namespace"].(string); eff != "" {
		t.Fatalf("effective_namespace=%q want empty", eff)
	}
}

func TestPublishRequiresTypeWhenUnknownExt(t *testing.T) {
	home := t.TempDir()
	t.Setenv("USERPROFILE", home)
	t.Setenv("HOME", home)
	t.Setenv("NACOS_SERVER", "http://127.0.0.1:8848")

	tmp := t.TempDir()
	file := filepath.Join(tmp, "app.dat")
	if err := os.WriteFile(file, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	root := NewRoot()
	buf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(errBuf)
	root.SetArgs([]string{
		"--server", "http://127.0.0.1:8848",
		"config", "publish",
		"--data-id", "app.dat",
		"--file", file,
		"--dry-run",
	})
	err := root.Execute()
	if err == nil {
		t.Fatal("expected error for unknown extension")
	}
	raw := buf.String() + errBuf.String()
	if !strings.Contains(raw, "--type") && !strings.Contains(err.Error(), "--type") {
		t.Fatalf("expected --type hint: err=%v out=%s", err, raw)
	}
}

func TestProfileSetRejectsMSYSContextPath(t *testing.T) {
	home := t.TempDir()
	t.Setenv("USERPROFILE", home)
	t.Setenv("HOME", home)
	root := NewRoot()
	buf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(errBuf)
	root.SetArgs([]string{
		"profile", "set", "badctx",
		"--server", "http://127.0.0.1:8848",
		"--context-path", "D:/Program Files/Git/nacos",
	})
	err := root.Execute()
	if err == nil {
		t.Fatal("expected reject")
	}
	raw := buf.String() + errBuf.String() + err.Error()
	if !strings.Contains(raw, "context-path") && !strings.Contains(raw, "MSYS_NO_PATHCONV") {
		t.Fatalf("expected context-path/MSYS hint: %s", raw)
	}
	dir, _ := config.ProfileDir()
	if _, err := os.Stat(filepath.Join(dir, "badctx.conf")); err == nil {
		t.Fatal("profile should not be saved")
	}
}

func TestProfileSetNormalizesPublicNamespace(t *testing.T) {
	home := t.TempDir()
	t.Setenv("USERPROFILE", home)
	t.Setenv("HOME", home)
	root := NewRoot()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{
		"profile", "set", "pub",
		"--server", "http://127.0.0.1:8848",
		"--namespace", "public",
		"--context-path", "/nacos",
	})
	if err := root.Execute(); err != nil {
		t.Fatal(err, buf.String())
	}
	dir, err := config.ProfileDir()
	if err != nil {
		t.Fatal(err)
	}
	p, ok, err := config.LoadProfileIn(dir, "pub")
	if err != nil || !ok {
		t.Fatalf("ok=%v err=%v", ok, err)
	}
	if p.Namespace != "" {
		t.Fatalf("stored namespace=%q want empty", p.Namespace)
	}
	if p.ContextPath != "/nacos" {
		t.Fatalf("contextPath=%q", p.ContextPath)
	}
}
