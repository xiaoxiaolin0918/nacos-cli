package backup

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSafeName(t *testing.T) {
	if SafeName("app/config.yaml") != "app__config.yaml" {
		t.Fatal(SafeName("app/config.yaml"))
	}
	if SafeName(`a:\b`) == "" {
		t.Fatal("empty")
	}
}

func TestWriteListRoundTrip(t *testing.T) {
	root := t.TempDir()
	stamp := NewStamp(time.Date(2026, 10, 21, 1, 2, 3, 0, time.Local))
	dir, err := CreateDir(root, stamp)
	if err != nil {
		t.Fatal(err)
	}
	item, err := WriteConfig(dir, "app.yaml", "DEFAULT_GROUP", "k: v\n")
	if err != nil {
		t.Fatal(err)
	}
	if item.Bytes == 0 || item.RelPath == "" {
		t.Fatalf("%+v", item)
	}
	m := Manifest{
		CreatedAt: stamp,
		Server:    "http://127.0.0.1:8848",
		Items:     []Item{item},
	}
	if err := WriteManifest(dir, m); err != nil {
		t.Fatal(err)
	}
	list, err := List(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 || list[0].Name != stamp || list[0].ItemCount != 1 {
		t.Fatalf("%+v", list)
	}
	got, err := ReadManifest(dir)
	if err != nil || len(got.Items) != 1 {
		t.Fatalf("%v %+v", err, got)
	}
}

func TestRootEnvOverride(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("NACOS_BACKUP_DIR", dir)
	if Root("") != dir {
		t.Fatal(Root(""))
	}
}

func TestEnsureWritable(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "nested")
	if err := EnsureWritable(dir); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(dir); err != nil {
		t.Fatal(err)
	}
}

func TestListMissingRoot(t *testing.T) {
	list, err := List(filepath.Join(t.TempDir(), "nope"))
	if err != nil || len(list) != 0 {
		t.Fatalf("%v %v", err, list)
	}
}
