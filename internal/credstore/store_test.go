package credstore

import (
	"path/filepath"
	"testing"
)

func TestSaveLoadClear(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "credentials.json")
	if err := Save(p, Cred{Server: "http://x", Username: "u", AccessToken: "abcdef123456", AuthMode: "token", Via: "password"}); err != nil {
		t.Fatal(err)
	}
	c, ok, err := Load(p)
	if err != nil || !ok || c.AccessToken != "abcdef123456" {
		t.Fatalf("%+v %v %v", c, ok, err)
	}
	st := Status(p)
	if st["has_token"] != true || st["token_masked"] == "" {
		t.Fatalf("%v", st)
	}
	if err := Clear(p); err != nil {
		t.Fatal(err)
	}
	_, ok, err = Load(p)
	if err != nil || ok {
		t.Fatalf("expected missing, ok=%v err=%v", ok, err)
	}
}

func TestMaskToken(t *testing.T) {
	if MaskToken("abcd") != "***" {
		t.Fatal("short")
	}
	if MaskToken("AUTH_DISABLED") != "" {
		t.Fatal("disabled")
	}
	m := MaskToken("abcdefghijklmnop")
	if m == "abcdefghijklmnop" || !contains(m, "…") {
		t.Fatal(m)
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || indexOf(s, sub) >= 0)
}
func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
