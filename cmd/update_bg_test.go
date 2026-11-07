package cmd

import (
	"testing"
	"time"

	"github.com/xiaoxiaolin0918/nacos-cli/internal/config"
)

func TestShouldSkipBackgroundUpdate(t *testing.T) {
	root := NewRoot()
	upd, _, err := root.Find([]string{"update"})
	if err != nil {
		t.Fatal(err)
	}
	if !shouldSkipBackgroundUpdate(upd) {
		t.Fatal("update should skip")
	}
	doc, _, err := root.Find([]string{"doctor"})
	if err != nil {
		t.Fatal(err)
	}
	if shouldSkipBackgroundUpdate(doc) {
		t.Fatal("doctor should not skip")
	}
}

func TestUpdateCheckCacheRoundTrip(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("APPDATA", dir)
	t.Setenv("XDG_CONFIG_HOME", dir)
	c := updateCheckCache{CheckedAt: time.Now().UTC(), Latest: "9.9.9", Status: "outdated"}
	saveUpdateCheckCache(c)
	got, ok := loadUpdateCheckCache()
	if !ok || got.Latest != "9.9.9" || got.Status != "outdated" {
		t.Fatalf("%v ok=%v", got, ok)
	}
}

func TestResolveUpdateFlagsDefaultsAndEnv(t *testing.T) {
	t.Setenv("NACOS_CHECK_UPDATE", "")
	t.Setenv("NACOS_AUTO_UPDATE", "")
	a := &App{}
	a.resolveUpdateFlags(config.Profile{})
	if a.checkUpdateEnabled || a.autoUpdateEnabled {
		t.Fatal("default off")
	}
	t.Setenv("NACOS_CHECK_UPDATE", "1")
	a = &App{}
	a.resolveUpdateFlags(config.Profile{})
	if !a.checkUpdateEnabled || a.autoUpdateEnabled {
		t.Fatal("check only")
	}
	t.Setenv("NACOS_AUTO_UPDATE", "true")
	a = &App{}
	a.resolveUpdateFlags(config.Profile{})
	if !a.checkUpdateEnabled || !a.autoUpdateEnabled {
		t.Fatal("auto implies check")
	}
	t.Setenv("NACOS_CHECK_UPDATE", "")
	t.Setenv("NACOS_AUTO_UPDATE", "")
	a = &App{}
	a.resolveUpdateFlags(config.Profile{CheckUpdate: true})
	if !a.checkUpdateEnabled || a.autoUpdateEnabled {
		t.Fatal("profile check_update")
	}
}
