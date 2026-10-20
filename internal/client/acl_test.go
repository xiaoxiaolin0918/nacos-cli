package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/xiaoxiaolin0918/nacos-cli/internal/config"
)

func TestUserCreateAndList(t *testing.T) {
	var sawCreate, sawList bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/nacos/v1/auth/users":
			sawCreate = true
			if r.URL.Query().Get("username") != "alice" {
				t.Fatalf("username %s", r.URL.Query().Get("username"))
			}
			if r.URL.Query().Get("password") == "" {
				t.Fatal("password missing")
			}
			_, _ = w.Write([]byte(`"ok"`))
		case r.Method == http.MethodGet && r.URL.Path == "/nacos/v1/auth/users":
			sawList = true
			_ = json.NewEncoder(w).Encode(map[string]any{
				"pageItems": []map[string]any{{"username": "alice"}},
				"totalCount": 1,
			})
		default:
			t.Fatalf("%s %s", r.Method, r.URL.Path)
		}
	}))
	defer srv.Close()
	c := New(config.Settings{Server: srv.URL, ContextPath: "/nacos"})
	c.ApplyToken("tok", "token")
	if _, err := c.UserCreate("alice", "secret"); err != nil {
		t.Fatal(err)
	}
	out, err := c.UserList("", 1, 10)
	if err != nil {
		t.Fatal(err)
	}
	if !sawCreate || !sawList {
		t.Fatalf("create=%v list=%v out=%v", sawCreate, sawList, out)
	}
}

func TestRoleBindPermissionGrant(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/nacos/v1/auth/roles":
			if q.Get("role") != "dev" || q.Get("username") != "alice" {
				t.Fatalf("%v", q)
			}
			_, _ = w.Write([]byte(`true`))
		case r.Method == http.MethodPost && r.URL.Path == "/nacos/v1/auth/permissions":
			if q.Get("role") != "dev" || q.Get("resource") == "" || q.Get("action") != "r" {
				t.Fatalf("%v", q)
			}
			_, _ = w.Write([]byte(`true`))
		case r.Method == http.MethodDelete && r.URL.Path == "/nacos/v1/auth/permissions":
			_, _ = w.Write([]byte(`true`))
		default:
			t.Fatalf("%s %s", r.Method, r.URL.Path)
		}
	}))
	defer srv.Close()
	c := New(config.Settings{Server: srv.URL, ContextPath: "/nacos"})
	if _, err := c.RoleBind("dev", "alice"); err != nil {
		t.Fatal(err)
	}
	if _, err := c.PermissionGrant("dev", "*:DEFAULT_GROUP:*", "r"); err != nil {
		t.Fatal(err)
	}
	if _, err := c.PermissionRevoke("dev", "*:DEFAULT_GROUP:*", "r"); err != nil {
		t.Fatal(err)
	}
}
