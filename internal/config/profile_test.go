package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProfileDirFor(t *testing.T) {
	d, err := profileDirFor("windows", `C:\Users\x`)
	want := filepath.Join(`C:\Users\x`, ".nacos-cli")
	if err != nil || d != want {
		t.Fatalf("got %q want %q err=%v", d, want, err)
	}
	d, err = profileDirFor("linux", "/home/u")
	want = filepath.Join("/home/u", ".nacos-cli")
	if err != nil || d != want {
		t.Fatalf("got %q want %q err=%v", d, want, err)
	}
}

func TestProfilePath(t *testing.T) {
	dir := t.TempDir()
	p, err := profilePathIn(dir, "default")
	if err != nil {
		t.Fatal(err)
	}
	if p != filepath.Join(dir, "default.conf") {
		t.Fatalf("%s", p)
	}
	_, err = profilePathIn(dir, "../evil")
	if err == nil {
		t.Fatal("expected path traversal reject")
	}
	_, err = profilePathIn(dir, "bad name")
	if err == nil {
		t.Fatal("expected invalid name reject")
	}
}

func TestSaveLoadProfile(t *testing.T) {
	dir := t.TempDir()
	p := Profile{
		Server:      "http://127.0.0.1:8848",
		Username:    "nacos",
		Password:    "secret-pw",
		Namespace:   "ns1",
		ContextPath: "/nacos",
		Format:      "json",
	}
	path, err := SaveProfileIn(dir, "lab", p)
	if err != nil {
		t.Fatal(err)
	}
	got, ok, err := LoadProfileIn(dir, "lab")
	if err != nil || !ok {
		t.Fatalf("ok=%v err=%v", ok, err)
	}
	if got.Server != p.Server || got.Username != p.Username || got.Password != p.Password {
		t.Fatalf("%+v", got)
	}
	if got.Namespace != "ns1" || got.ContextPath != "/nacos" || got.Format != "json" {
		t.Fatalf("%+v", got)
	}
	raw, _ := os.ReadFile(path)
	if !strings.Contains(string(raw), "server:") {
		t.Fatalf("expected yaml keys: %s", raw)
	}
}

func TestLoadProfileHostPort(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "dev.conf")
	body := "host: 10.0.0.1\nport: 8849\nauthType: nacos\nusername: u1\nnamespace: \"\"\n"
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	got, ok, err := LoadProfileIn(dir, "dev")
	if err != nil || !ok {
		t.Fatalf("ok=%v err=%v", ok, err)
	}
	if got.Server != "http://10.0.0.1:8849" {
		t.Fatalf("server=%q", got.Server)
	}
	if got.Username != "u1" {
		t.Fatalf("user=%q", got.Username)
	}
}

func TestListProfiles(t *testing.T) {
	dir := t.TempDir()
	_, _ = SaveProfileIn(dir, "default", Profile{Server: "http://a:8848"})
	_, _ = SaveProfileIn(dir, "lab", Profile{Server: "http://b:8848"})
	names, err := ListProfilesIn(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(names) != 2 || names[0] != "default" || names[1] != "lab" {
		t.Fatalf("%v", names)
	}
}

func TestMergeSettings(t *testing.T) {
	base := Settings{Server: "http://env:8848", Username: "envuser", Namespace: "envns", ContextPath: "/nacos"}
	prof := Profile{Server: "http://prof:8848", Username: "profuser", Format: "pretty"}
	merged := MergeSettings(base, prof)
	if merged.Server != "http://prof:8848" || merged.Username != "profuser" {
		t.Fatalf("%+v", merged)
	}
	if merged.Namespace != "envns" {
		t.Fatalf("namespace should keep env when profile empty: %+v", merged)
	}
	onlyNS := Profile{Namespace: "from-profile"}
	m2 := MergeSettings(base, onlyNS)
	if m2.Server != "http://env:8848" || m2.Namespace != "from-profile" {
		t.Fatalf("%+v", m2)
	}
}

func TestSecretKey(t *testing.T) {
	if !SecretKey("password") || !SecretKey("Password") {
		t.Fatal("password should be secret")
	}
	if SecretKey("server") {
		t.Fatal("server is not secret")
	}
}

func TestNormalizeFormat(t *testing.T) {
	cases := map[string]string{
		"":       "json",
		"json":   "json",
		"pretty": "pretty",
		"JSON":   "json",
		"Pretty": "pretty",
	}
	for in, want := range cases {
		got, err := NormalizeFormat(in)
		if err != nil || got != want {
			t.Fatalf("%q -> %q err=%v want %q", in, got, err, want)
		}
	}
	if _, err := NormalizeFormat("table"); err == nil {
		t.Fatal("table not supported yet")
	}
	if _, err := NormalizeFormat("xml"); err == nil {
		t.Fatal("expected error")
	}
}

func TestDeleteProfile(t *testing.T) {
	dir := t.TempDir()
	_, _ = SaveProfileIn(dir, "tmp", Profile{Server: "http://x:8848"})
	if err := DeleteProfileIn(dir, "tmp"); err != nil {
		t.Fatal(err)
	}
	_, ok, err := LoadProfileIn(dir, "tmp")
	if err != nil || ok {
		t.Fatalf("ok=%v err=%v", ok, err)
	}
}