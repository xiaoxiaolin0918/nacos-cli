package cmd

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/xiaoxiaolin0918/nacos-cli/internal/credstore"
)

func runCLI(t *testing.T, env map[string]string, args ...string) (string, error) {
	t.Helper()
	for k, v := range env {
		t.Setenv(k, v)
	}
	root := NewRoot()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs(args)
	err := root.Execute()
	return buf.String(), err
}

func TestAutoLoginOn401ThenReplay(t *testing.T) {
	var hits atomic.Int32
	var loginHits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/v1/auth/login"):
			loginHits.Add(1)
			_ = json.NewEncoder(w).Encode(map[string]any{"accessToken": "fresh-token", "tokenTtl": 18000})
		case strings.HasSuffix(r.URL.Path, "/v1/console/namespaces") || strings.Contains(r.URL.Path, "/namespaces"):
			n := hits.Add(1)
			tok := r.URL.Query().Get("accessToken")
			if n == 1 {
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = w.Write([]byte(`{"message":"token expired"}`))
				return
			}
			if tok != "fresh-token" {
				t.Errorf("replay token=%q", tok)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"code": 200, "data": []any{}})
		default:
			// Probe/Ping/ServerState may hit other paths; answer ok.
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{}`))
		}
	}))
	defer srv.Close()

	credDir := t.TempDir()
	credPath := filepath.Join(credDir, "credentials.json")
	if err := credstore.Save(credPath, credstore.Cred{AccessToken: "stale", AuthMode: "token", Username: "alice"}); err != nil {
		t.Fatal(err)
	}

	out, err := runCLI(t, map[string]string{
		"NACOS_USERNAME":         "alice",
		"NACOS_PASSWORD":         "s3cret-auto-login",
		"NACOS_AUTO_LOGIN":       "1",
		"NACOS_CREDENTIALS_PATH": credPath,
		"NACOS_CONTEXT_PATH":     "/",
	}, "--server", srv.URL, "namespace", "list")
	if err != nil {
		t.Fatal(err, out)
	}
	if strings.Contains(out, "s3cret-auto-login") {
		t.Fatalf("password leaked: %s", out)
	}
	if hits.Load() < 1 || loginHits.Load() != 1 {
		t.Fatalf("hits=%d login=%d out=%s", hits.Load(), loginHits.Load(), out)
	}
	c, ok, _ := credstore.Load(credPath)
	if !ok || c.AccessToken != "fresh-token" {
		t.Fatalf("token not written back: %+v", c)
	}
}

func TestAutoLoginOnlyOnce(t *testing.T) {
	var loginHits atomic.Int32
	var nsHits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/v1/auth/login"):
			loginHits.Add(1)
			_ = json.NewEncoder(w).Encode(map[string]any{"accessToken": "t", "tokenTtl": 1})
		case strings.Contains(r.URL.Path, "namespace"):
			nsHits.Add(1)
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"message":"still bad"}`))
		default:
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{}`))
		}
	}))
	defer srv.Close()

	credDir := t.TempDir()
	credPath := filepath.Join(credDir, "credentials.json")
	_ = credstore.Save(credPath, credstore.Cred{AccessToken: "stale", AuthMode: "token"})

	out, err := runCLI(t, map[string]string{
		"NACOS_USERNAME":         "alice",
		"NACOS_PASSWORD":         "s3cret-once",
		"NACOS_CREDENTIALS_PATH": credPath,
		"NACOS_CONTEXT_PATH":     "/",
	}, "--server", srv.URL, "--auto-login", "namespace", "list")
	if err == nil {
		t.Fatal("expected persistent 401", out)
	}
	if loginHits.Load() != 1 || nsHits.Load() != 2 {
		t.Fatalf("login=%d ns=%d out=%s", loginHits.Load(), nsHits.Load(), out)
	}
	if strings.Contains(out, "s3cret-once") {
		t.Fatalf("password leaked: %s", out)
	}
}

func TestAutoLoginFromEncryptedCredStore(t *testing.T) {
	var loginHits atomic.Int32
	var nsHits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/v1/auth/login"):
			loginHits.Add(1)
			b, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(b), "store-user") || !strings.Contains(string(b), "store-pass-secret") {
				t.Errorf("login body=%s", b)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"accessToken": "from-store", "tokenTtl": 9})
		case strings.Contains(r.URL.Path, "namespace"):
			n := nsHits.Add(1)
			if n == 1 {
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = w.Write([]byte(`no`))
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"code": 200, "data": []any{}})
		default:
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{}`))
		}
	}))
	defer srv.Close()

	credDir := t.TempDir()
	credPath := filepath.Join(credDir, "credentials.json")
	_ = credstore.Save(credPath, credstore.Cred{AccessToken: "stale", AuthMode: "token"})
	if err := credstore.SaveLogin(credDir, credstore.LoginPassword{Username: "store-user", Password: "store-pass-secret", AutoLogin: true}); err != nil {
		t.Fatal(err)
	}

	// Clear env pair so login.cred is used atomically.
	out, err := runCLI(t, map[string]string{
		"NACOS_USERNAME":         "",
		"NACOS_PASSWORD":         "",
		"NACOS_CREDENTIALS_PATH": credPath,
		"NACOS_CONTEXT_PATH":     "/",
	}, "--server", srv.URL, "namespace", "list")
	if err != nil {
		t.Fatal(err, out)
	}
	if strings.Contains(out, "store-pass-secret") {
		t.Fatalf("password leaked: %s", out)
	}
	if loginHits.Load() != 1 || nsHits.Load() != 2 {
		t.Fatalf("login=%d ns=%d out=%s", loginHits.Load(), nsHits.Load(), out)
	}
}

func TestLogoutKeepsLoginCredUnlessForget(t *testing.T) {
	credDir := t.TempDir()
	credPath := filepath.Join(credDir, "credentials.json")
	_ = credstore.Save(credPath, credstore.Cred{AccessToken: "tok", AuthMode: "token", Username: "u"})
	if err := credstore.SaveLogin(credDir, credstore.LoginPassword{Username: "u", Password: "keep-me", AutoLogin: true}); err != nil {
		t.Fatal(err)
	}

	out, err := runCLI(t, map[string]string{"NACOS_CREDENTIALS_PATH": credPath}, "auth", "logout")
	if err != nil {
		t.Fatal(err, out)
	}
	if _, ok, _ := credstore.Load(credPath); ok {
		t.Fatal("token should be cleared")
	}
	if !credstore.HasLogin(credDir) {
		t.Fatal("login.cred should remain after logout")
	}

	out, err = runCLI(t, map[string]string{"NACOS_CREDENTIALS_PATH": credPath}, "auth", "logout", "--forget")
	if err != nil {
		t.Fatal(err, out)
	}
	if credstore.HasLogin(credDir) {
		t.Fatal("login.cred should be cleared with --forget")
	}
}

func TestLoginRememberWritesEncryptedCred(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"accessToken": "tok-rem", "tokenTtl": 1})
	}))
	defer srv.Close()

	credDir := t.TempDir()
	credPath := filepath.Join(credDir, "credentials.json")
	out, err := runCLI(t, map[string]string{
		"NACOS_CREDENTIALS_PATH": credPath,
		"NACOS_CONTEXT_PATH":     "/",
	}, "--server", srv.URL, "auth", "login", "--username", "alice", "--password", "pw-secret-rem", "--remember")
	if err != nil {
		t.Fatal(err, out)
	}
	if strings.Contains(out, "pw-secret-rem") {
		t.Fatalf("password leaked: %s", out)
	}
	raw, _ := os.ReadFile(credstore.LoginPath(credDir))
	if strings.Contains(string(raw), "pw-secret-rem") {
		t.Fatal("plaintext in login.cred")
	}
	got, ok, err := credstore.LoadLogin(credDir)
	if err != nil || !ok || got.Password != "pw-secret-rem" || !got.AutoLogin {
		t.Fatalf("%+v ok=%v err=%v", got, ok, err)
	}
}

func TestProfileSetRejectsPassword(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("USERPROFILE", dir) // windows ProfileDir uses home via other path; use --config file instead
	out, err := runCLI(t, nil, "profile", "set", "tmp-no-pw", "--server", "http://127.0.0.1:8848", "--password", "nope-secret")
	if err == nil {
		t.Fatal("expected reject", out)
	}
	if !strings.Contains(out, "remember") && !strings.Contains(out, "password") {
		t.Fatalf("%s", out)
	}
	if strings.Contains(out, "nope-secret") {
		t.Fatalf("password leaked: %s", out)
	}
}

func TestLoginRememberFalseClearsCred(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"accessToken": "tok2", "tokenTtl": 1})
	}))
	defer srv.Close()

	credDir := t.TempDir()
	credPath := filepath.Join(credDir, "credentials.json")
	if err := credstore.SaveLogin(credDir, credstore.LoginPassword{Username: "alice", Password: "old-secret", AutoLogin: true}); err != nil {
		t.Fatal(err)
	}
	out, err := runCLI(t, map[string]string{
		"NACOS_CREDENTIALS_PATH": credPath,
		"NACOS_CONTEXT_PATH":     "/",
	}, "--server", srv.URL, "auth", "login", "--username", "alice", "--password", "new-secret")
	if err != nil {
		t.Fatal(err, out)
	}
	if credstore.HasLogin(credDir) {
		t.Fatal("login without --remember must Clear prior login.cred")
	}
}
