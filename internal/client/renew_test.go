package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/xiaoxiaolin0918/nacos-cli/internal/config"
)

func TestDoWithRenewOn401Once(t *testing.T) {
	var hits atomic.Int32
	var renewHits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := hits.Add(1)
		tok := r.URL.Query().Get("accessToken")
		if n == 1 {
			if tok != "stale" {
				t.Errorf("first token=%q", tok)
			}
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"message":"token expired"}`))
			return
		}
		if tok != "fresh" {
			t.Errorf("replay token=%q", tok)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"count": 1})
	}))
	defer srv.Close()

	c := New(config.Settings{Server: srv.URL, ContextPath: ""})
	c.ApplyToken("stale", "token")
	c.RenewOnce = func(ctx context.Context) error {
		renewHits.Add(1)
		c.ApplyToken("fresh", "token")
		return nil
	}
	v, err := c.getJSON(context.Background(), "/v1/cs/configs", nil)
	if err != nil {
		t.Fatal(err)
	}
	if hits.Load() != 2 || renewHits.Load() != 1 {
		t.Fatalf("hits=%d renew=%d v=%v", hits.Load(), renewHits.Load(), v)
	}
	// second call must NOT renew again (RenewOnce already nilled)
	hits.Store(0)
	c.ApplyToken("stale", "token")
	_, err = c.getJSON(context.Background(), "/v1/cs/configs", nil)
	if err == nil {
		t.Fatal("expected 401 without renew")
	}
	if hits.Load() != 1 || renewHits.Load() != 1 {
		t.Fatalf("after: hits=%d renew=%d", hits.Load(), renewHits.Load())
	}
}

func TestDoWithRenewSkipsWhenRenewFails(t *testing.T) {
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`no`))
	}))
	defer srv.Close()
	c := New(config.Settings{Server: srv.URL, ContextPath: ""})
	c.ApplyToken("x", "token")
	c.RenewOnce = func(ctx context.Context) error {
		return context.Canceled
	}
	_, err := c.getJSON(context.Background(), "/x", nil)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "context canceled") {
		t.Fatalf("renew error must surface, got %v", err)
	}
	if hits.Load() != 1 {
		t.Fatalf("hits=%d", hits.Load())
	}
}

func TestDoWithRenewSkipsACL403(t *testing.T) {
	var hits atomic.Int32
	var renewHits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"message":"permission denied"}`))
	}))
	defer srv.Close()
	c := New(config.Settings{Server: srv.URL, ContextPath: ""})
	c.ApplyToken("tok", "token")
	c.RenewOnce = func(ctx context.Context) error {
		renewHits.Add(1)
		return nil
	}
	_, err := c.getJSON(context.Background(), "/x", nil)
	if err == nil {
		t.Fatal("expected ACL error")
	}
	if renewHits.Load() != 0 || hits.Load() != 1 {
		t.Fatalf("ACL 403 must not renew: hits=%d renew=%d", hits.Load(), renewHits.Load())
	}
}

func TestConfigPublishRenewsOn401(t *testing.T) {
	var hits atomic.Int32
	var renewHits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := hits.Add(1)
		if n == 1 {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`token expired`))
			return
		}
		_, _ = w.Write([]byte(`true`))
	}))
	defer srv.Close()
	c := New(config.Settings{Server: srv.URL, ContextPath: ""})
	c.ApplyToken("stale", "token")
	c.RenewOnce = func(ctx context.Context) error {
		renewHits.Add(1)
		c.ApplyToken("fresh", "token")
		return nil
	}
	v, err := c.ConfigPublish("a", "DEFAULT_GROUP", "x", "", "text")
	if err != nil {
		t.Fatal(err)
	}
	if hits.Load() != 2 || renewHits.Load() != 1 {
		t.Fatalf("hits=%d renew=%d v=%v", hits.Load(), renewHits.Load(), v)
	}
}
