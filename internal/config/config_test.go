package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBaseURL(t *testing.T) {
	s := Settings{Server: "http://127.0.0.1:8848/", ContextPath: "/nacos"}
	if got := s.BaseURL(); got != "http://127.0.0.1:8848/nacos" {
		t.Fatalf("%s", got)
	}
	s.ContextPath = ""
	if got := s.BaseURL(); got != "http://127.0.0.1:8848/nacos" {
		t.Fatalf("%s", got)
	}
}

func TestLoadDotEnv(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, ".env")
	if err := os.WriteFile(p, []byte("NACOS_SERVER=http://example:8848\nNACOS_USERNAME=u1\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	_ = os.Unsetenv("NACOS_SERVER")
	_ = os.Unsetenv("NACOS_USERNAME")
	if err := LoadDotEnv(p); err != nil {
		t.Fatal(err)
	}
	if os.Getenv("NACOS_SERVER") != "http://example:8848" {
		t.Fatal(os.Getenv("NACOS_SERVER"))
	}
	s := FromEnv()
	if s.Server != "http://example:8848" || s.Username != "u1" {
		t.Fatalf("%+v", s)
	}
}

func TestDirFor(t *testing.T) {
	d, err := dirFor("windows", `C:\Users\x\AppData\Roaming`, "", "")
	wantWin := filepath.Join(`C:\Users\x\AppData\Roaming`, "nacos-cli")
	if err != nil || d != wantWin {
		t.Fatalf("got %q want %q err=%v", d, wantWin, err)
	}
	d, err = dirFor("linux", "", "/xdg", "/home/u")
	wantLin := filepath.Join("/xdg", "nacos-cli")
	if err != nil || d != wantLin {
		t.Fatalf("got %q want %q err=%v", d, wantLin, err)
	}
}
