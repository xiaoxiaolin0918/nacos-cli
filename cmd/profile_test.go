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

func TestOutputAliasPretty(t *testing.T) {
	root := NewRoot()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"profile", "path", "default", "--output", "pretty"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	raw := buf.String()
	if !strings.Contains(raw, "\n  ") {
		t.Fatalf("expected pretty indent: %s", raw)
	}
	var env map[string]any
	if err := json.Unmarshal(buf.Bytes(), &env); err != nil {
		t.Fatal(err, raw)
	}
	if env["ok"] != true {
		t.Fatalf("%v", env)
	}
}

func TestOutputRejectsUnknown(t *testing.T) {
	root := NewRoot()
	buf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(errBuf)
	root.SetArgs([]string{"profile", "path", "--output", "xml"})
	err := root.Execute()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestProfileSetShowListDelete(t *testing.T) {
	home := t.TempDir()
	t.Setenv("USERPROFILE", home)
	t.Setenv("HOME", home)
	name := "tdd-lab"
	root := NewRoot()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{
		"profile", "set", name,
		"--server", "http://127.0.0.1:8848",
		"--username", "nacos",
		"--password", "super-secret",
		"--default-format", "pretty",
	})
	if err := root.Execute(); err != nil {
		t.Fatal(err, buf.String())
	}
	if strings.Contains(buf.String(), "super-secret") {
		t.Fatal("password leaked on set")
	}
	dir, err := config.ProfileDir()
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, name+".conf")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("profile missing; ProfileDir=%s home=%s err=%v", dir, home, err)
	}

	buf.Reset()
	root2 := NewRoot()
	root2.SetOut(buf)
	root2.SetErr(&bytes.Buffer{})
	root2.SetArgs([]string{"profile", "show", name})
	if err := root2.Execute(); err != nil {
		t.Fatal(err)
	}
	raw := buf.String()
	if strings.Contains(raw, "super-secret") {
		t.Fatal("password leaked on show")
	}
	var env map[string]any
	if err := json.Unmarshal(buf.Bytes(), &env); err != nil {
		t.Fatal(err, raw)
	}
	res := env["result"].(map[string]any)
	prof := res["profile"].(map[string]any)
	if prof["password"] != "******" {
		t.Fatalf("%v", prof)
	}
	if prof["server"] != "http://127.0.0.1:8848" {
		t.Fatalf("%v", prof)
	}

	buf.Reset()
	root3 := NewRoot()
	root3.SetOut(buf)
	root3.SetErr(&bytes.Buffer{})
	root3.SetArgs([]string{"profile", "list"})
	if err := root3.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), name) {
		t.Fatalf("list missing name: %s", buf.String())
	}

	buf.Reset()
	root4 := NewRoot()
	root4.SetOut(buf)
	root4.SetErr(&bytes.Buffer{})
	root4.SetArgs([]string{"profile", "delete", name})
	err = root4.Execute()
	if err == nil {
		t.Fatal("delete requires --yes")
	}

	buf.Reset()
	root5 := NewRoot()
	root5.SetOut(buf)
	root5.SetErr(&bytes.Buffer{})
	root5.SetArgs([]string{"profile", "delete", name, "--dry-run"})
	if err := root5.Execute(); err != nil {
		t.Fatal(err)
	}
	var e2 map[string]any
	if err := json.Unmarshal(buf.Bytes(), &e2); err != nil {
		t.Fatal(err, buf.String())
	}
	if e2["dry_run"] != true {
		t.Fatalf("%s", buf.String())
	}

	buf.Reset()
	root6 := NewRoot()
	root6.SetOut(buf)
	root6.SetErr(&bytes.Buffer{})
	root6.SetArgs([]string{"profile", "delete", name, "--yes"})
	if err := root6.Execute(); err != nil {
		t.Fatal(err)
	}
}

func TestProfileHostPortCompose(t *testing.T) {
	home := t.TempDir()
	t.Setenv("USERPROFILE", home)
	t.Setenv("HOME", home)
	name := "hostport"
	root := NewRoot()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"profile", "set", name, "--host", "10.0.0.2", "--port", "8849", "--username", "u"})
	if err := root.Execute(); err != nil {
		t.Fatal(err, buf.String())
	}
	dir, err := config.ProfileDir()
	if err != nil {
		t.Fatal(err)
	}
	p, ok, err := config.LoadProfileIn(dir, name)
	if err != nil || !ok {
		t.Fatalf("ok=%v err=%v dir=%s", ok, err, dir)
	}
	if p.Server != "http://10.0.0.2:8849" {
		t.Fatalf("%+v", p)
	}
	_ = config.DeleteProfileIn(dir, name)
}