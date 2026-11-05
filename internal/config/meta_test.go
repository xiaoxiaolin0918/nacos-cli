package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMetaPathIn(t *testing.T) {
	dir := t.TempDir()
	p, err := metaPathIn(dir, "lab")
	if err != nil {
		t.Fatal(err)
	}
	if p != filepath.Join(dir, "lab.meta.yaml") {
		t.Fatalf("got %s", p)
	}
	if _, err := metaPathIn(dir, "../evil"); err == nil {
		t.Fatal("expected reject")
	}
}

func TestSaveLoadMeta(t *testing.T) {
	dir := t.TempDir()
	m := ProfileMeta{
		DisplayName: "支付实验室",
		EnvType:     "lab",
		Color:       "#1677ff",
		Owner:       "xiaolin",
		Notes:       "只读演练",
	}
	path, err := SaveMetaIn(dir, "lab", m)
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(path) != "lab.meta.yaml" {
		t.Fatalf("path=%s", path)
	}
	got, ok, err := LoadMetaIn(dir, "lab")
	if err != nil || !ok {
		t.Fatalf("ok=%v err=%v", ok, err)
	}
	if got.DisplayName != m.DisplayName || got.EnvType != "lab" || got.Color != "#1677ff" || got.Owner != "xiaolin" || got.Notes != "只读演练" {
		t.Fatalf("%+v", got)
	}
	raw, _ := os.ReadFile(path)
	s := string(raw)
	for _, key := range []string{"displayName:", "envType:", "color:", "owner:", "notes:"} {
		if !strings.Contains(s, key) {
			t.Fatalf("yaml missing %s: %s", key, s)
		}
	}
}

func TestLoadMetaMissing(t *testing.T) {
	dir := t.TempDir()
	got, ok, err := LoadMetaIn(dir, "none")
	if err != nil || ok {
		t.Fatalf("ok=%v err=%v got=%+v", ok, err, got)
	}
}

func TestNormalizeMetaEnvType(t *testing.T) {
	m := ProfileMeta{EnvType: "PROD"}
	m.normalize()
	if m.EnvType != "prod" {
		t.Fatalf("%q", m.EnvType)
	}
	m = ProfileMeta{EnvType: "staging"}
	if err := m.validate(); err == nil {
		t.Fatal("expected invalid envType")
	}
	m = ProfileMeta{EnvType: "lab"}
	if err := m.validate(); err != nil {
		t.Fatal(err)
	}
	m = ProfileMeta{EnvType: ""}
	if err := m.validate(); err != nil {
		t.Fatal(err)
	}
}

func TestDeleteMetaIn(t *testing.T) {
	dir := t.TempDir()
	_, _ = SaveMetaIn(dir, "tmp", ProfileMeta{DisplayName: "x"})
	if err := DeleteMetaIn(dir, "tmp"); err != nil {
		t.Fatal(err)
	}
	_, ok, err := LoadMetaIn(dir, "tmp")
	if err != nil || ok {
		t.Fatalf("ok=%v err=%v", ok, err)
	}
}
