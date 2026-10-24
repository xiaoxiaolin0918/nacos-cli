package client

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/xiaoxiaolin0918/nacos-cli/internal/config"
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
