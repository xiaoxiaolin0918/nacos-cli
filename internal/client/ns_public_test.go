package client

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/xiaoxiaolin0918/nacos-cli/internal/config"
	"github.com/xiaoxiaolin0918/nacos-cli/internal/domain"
)

func TestNSMapsPublicToEmptyTenant(t *testing.T) {
	var gotTenant string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotTenant = r.URL.Query().Get("tenant")
		w.WriteHeader(404)
		_, _ = w.Write([]byte("config data not exist"))
	}))
	defer srv.Close()
	c := New(config.Settings{Server: srv.URL, ContextPath: "/nacos", Namespace: "public"})
	_, _ = c.ConfigGet("a.yaml", "DEFAULT_GROUP", "")
	if gotTenant != "" {
		t.Fatalf("settings public -> tenant %q", gotTenant)
	}
	_, _ = c.ConfigGet("a.yaml", "DEFAULT_GROUP", "Public")
	if gotTenant != "" {
		t.Fatalf("explicit Public -> tenant %q", gotTenant)
	}
}

// I1: after cmd EffectiveNamespace("public", "prod") resolves to "", client must NOT
// fall back to Settings.Namespace=prod. dry-run and wire tenant must both stay empty.
func TestResolvedEmptyTenantNotFallenBackToSettings(t *testing.T) {
	var gotTenant string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotTenant = r.URL.Query().Get("tenant")
		w.WriteHeader(404)
		_, _ = w.Write([]byte("config data not exist"))
	}))
	defer srv.Close()

	settingsNS := "prod"
	c := New(config.Settings{Server: srv.URL, ContextPath: "/nacos", Namespace: settingsNS})

	resolved := domain.EffectiveNamespace("public", settingsNS)
	if resolved != "" {
		t.Fatalf("EffectiveNamespace(public,%s)=%q want empty", settingsNS, resolved)
	}

	uc := domain.ConfigUseCase{Port: c}
	_, _ = uc.Get("a.yaml", "DEFAULT_GROUP", resolved)
	if gotTenant != "" {
		t.Fatalf("usecase Get after public resolve: wire tenant=%q want empty (settings=%s)", gotTenant, settingsNS)
	}

	gotTenant = "SENTINEL"
	_, _ = c.ConfigGet("a.yaml", "DEFAULT_GROUP", "")
	if gotTenant != "" {
		t.Fatalf("client ConfigGet empty ns with Settings=%s: wire tenant=%q want empty", settingsNS, gotTenant)
	}

	gotTenant = "SENTINEL"
	_, _ = c.ConfigGet("a.yaml", "DEFAULT_GROUP", "public")
	if gotTenant != "" {
		t.Fatalf("client ConfigGet public with Settings=%s: wire tenant=%q want empty", settingsNS, gotTenant)
	}
}
