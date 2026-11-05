package credstore

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSaveLoadLoginRoundTrip(t *testing.T) {
	dir := t.TempDir()
	err := SaveLogin(dir, LoginPassword{Username: "alice", Password: "s3cret-roundtrip", AutoLogin: true})
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(LoginPath(dir))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), "s3cret-roundtrip") {
		t.Fatal("plaintext password landed in login.cred")
	}
	if !strings.Contains(string(raw), `"auto_login": true`) {
		t.Fatalf("missing auto_login: %s", raw)
	}
	got, ok, err := LoadLogin(dir)
	if err != nil || !ok {
		t.Fatalf("load: ok=%v err=%v", ok, err)
	}
	if got.Username != "alice" || got.Password != "s3cret-roundtrip" || !got.AutoLogin {
		t.Fatalf("%+v", got)
	}
	u, auto, ok := PeekLogin(dir)
	if !ok || u != "alice" || !auto {
		t.Fatalf("peek %q %v %v", u, auto, ok)
	}
	if !HasLogin(dir) {
		t.Fatal("HasLogin")
	}
}

func TestLoadLoginMissing(t *testing.T) {
	_, ok, err := LoadLogin(t.TempDir())
	if err != nil || ok {
		t.Fatalf("ok=%v err=%v", ok, err)
	}
}

func TestClearLogin(t *testing.T) {
	dir := t.TempDir()
	if err := SaveLogin(dir, LoginPassword{Username: "u", Password: "p", AutoLogin: false}); err != nil {
		t.Fatal(err)
	}
	if err := ClearLogin(dir); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(LoginPath(dir)); !os.IsNotExist(err) {
		t.Fatalf("cred still present: %v", err)
	}
	if err := ClearLogin(dir); err != nil {
		t.Fatal(err)
	}
}

func TestSaveLoginRejectsEmpty(t *testing.T) {
	dir := t.TempDir()
	if err := SaveLogin(dir, LoginPassword{Username: "", Password: "p"}); err == nil {
		t.Fatal("expected username error")
	}
	if err := SaveLogin(dir, LoginPassword{Username: "u", Password: ""}); err == nil {
		t.Fatal("expected password error")
	}
}

func TestDirFromCredentialsPath(t *testing.T) {
	got := DirFromCredentialsPath(filepath.Join("a", "b", "credentials.json"))
	if got != filepath.Join("a", "b") {
		t.Fatalf("%s", got)
	}
}
