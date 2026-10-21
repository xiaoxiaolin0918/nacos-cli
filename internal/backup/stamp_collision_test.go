package backup

import (
	"path/filepath"
	"testing"
)

func TestCreateDirUniqueOnCollision(t *testing.T) {
	root := t.TempDir()
	a, err := CreateDir(root, "20261021-010203")
	if err != nil {
		t.Fatal(err)
	}
	b, err := CreateDir(root, "20261021-010203")
	if err != nil {
		t.Fatal(err)
	}
	if a == b {
		t.Fatalf("same-second backups must not share a dir: %s", a)
	}
	if filepath.Base(b) != "20261021-010203-2" {
		t.Fatal(b)
	}
}
