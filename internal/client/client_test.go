package client

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/xiaoxiaolin0918/nacos-cli/internal/config"
	"github.com/xiaoxiaolin0918/nacos-cli/internal/domain"
)

func TestLoginAuthDisabled(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/nacos/v1/auth/login" {
			t.Fatalf("path %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"accessToken": "AUTH_DISABLED"})
	}))
	defer srv.Close()
	c := New(config.Settings{Server: srv.URL, ContextPath: "/nacos"})
	res, err := c.Login(context.Background(), "nacos", "")
	if err != nil || res.AuthMode != "disabled" || res.HasToken {
		t.Fatalf("%+v %v", res, err)
	}
}

func TestConfigGetNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		_, _ = w.Write([]byte("config data not exist"))
	}))
	defer srv.Close()
	c := New(config.Settings{Server: srv.URL, ContextPath: "/nacos"})
	_, err := c.ConfigGet("missing.yaml", "DEFAULT_GROUP", "")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("%v", err)
	}
}
