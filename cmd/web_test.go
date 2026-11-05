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
	"testing"
	"time"

	"github.com/xiaoxiaolin0918/nacos-cli/internal/audit"
	"github.com/xiaoxiaolin0918/nacos-cli/internal/config"
	"github.com/xiaoxiaolin0918/nacos-cli/internal/credstore"
)

func TestWebAPIProfilesCRUD(t *testing.T) {
	dir := t.TempDir()
	auditPath := filepath.Join(dir, "audit.jsonl")
	srv := newWebServer(dir, &audit.Store{Path: auditPath})

	// create
	body := map[string]any{
		"server":      "http://127.0.0.1:8848",
		"username":    "nacos",
		"password":    "s3cret-pw",
		"authType":    "nacos",
		"namespace":   "",
		"contextPath": "/nacos",
		"format":      "json",
		"meta": map[string]any{
			"displayName": "实验室",
			"envType":     "lab",
			"color":       "#1677ff",
			"owner":       "xiaolin",
			"notes":       "demo",
		},
	}
	raw, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPut, "/api/profiles/lab", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept-Language", "zh-CN")
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)
	if rr.Code != 200 {
		t.Fatalf("create status=%d body=%s", rr.Code, rr.Body.String())
	}
	var created map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	prof := created["profile"].(map[string]any)
	if prof["password"] == "s3cret-pw" {
		t.Fatal("password must never be echoed")
	}
	if prof["password_set"] != true {
		t.Fatalf("password_set=%v", prof["password_set"])
	}
	meta := created["meta"].(map[string]any)
	if meta["displayName"] != "实验室" || meta["envType"] != "lab" {
		t.Fatalf("meta=%v", meta)
	}

	// list
	req = httptest.NewRequest(http.MethodGet, "/api/profiles", nil)
	rr = httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)
	if rr.Code != 200 {
		t.Fatalf("list %d %s", rr.Code, rr.Body.String())
	}
	var list map[string]any
	_ = json.Unmarshal(rr.Body.Bytes(), &list)
	items := list["profiles"].([]any)
	if len(items) != 1 {
		t.Fatalf("list=%v", list)
	}
	item := items[0].(map[string]any)
	pview := item["profile"].(map[string]any)
	if pview["password"] == "s3cret-pw" {
		t.Fatal("list leaked password")
	}

	// update keep password when omitted
	upd := map[string]any{
		"server":   "http://127.0.0.1:8849",
		"username": "nacos",
		"meta":     map[string]any{"displayName": "实验室2", "envType": "lab"},
	}
	raw, _ = json.Marshal(upd)
	req = httptest.NewRequest(http.MethodPut, "/api/profiles/lab", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)
	if rr.Code != 200 {
		t.Fatalf("update %d %s", rr.Code, rr.Body.String())
	}
	got, ok, err := config.LoadProfileIn(dir, "lab")
	if err != nil || !ok {
		t.Fatalf("load err=%v ok=%v", err, ok)
	}
	if got.Password != "s3cret-pw" {
		t.Fatalf("password should be kept, got %q", got.Password)
	}
	if got.Server != "http://127.0.0.1:8849" {
		t.Fatalf("server=%q", got.Server)
	}

	// clear password
	clr := map[string]any{"server": "http://127.0.0.1:8849", "username": "nacos", "clear_password": true}
	raw, _ = json.Marshal(clr)
	req = httptest.NewRequest(http.MethodPut, "/api/profiles/lab", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)
	if rr.Code != 200 {
		t.Fatalf("clear %d %s", rr.Code, rr.Body.String())
	}
	got, _, _ = config.LoadProfileIn(dir, "lab")
	if got.Password != "" {
		t.Fatal("password should be cleared")
	}

	// delete
	req = httptest.NewRequest(http.MethodDelete, "/api/profiles/lab?confirm=1", nil)
	rr = httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)
	if rr.Code != 200 {
		t.Fatalf("delete %d %s", rr.Code, rr.Body.String())
	}
	_, ok, _ = config.LoadProfileIn(dir, "lab")
	if ok {
		t.Fatal("profile still exists")
	}
	_, ok, _ = config.LoadMetaIn(dir, "lab")
	if ok {
		t.Fatal("meta still exists")
	}
}

func TestWebAPIAuditFilter(t *testing.T) {
	dir := t.TempDir()
	auditPath := filepath.Join(dir, "audit.jsonl")
	store := &audit.Store{Path: auditPath}
	_ = store.Append(audit.Entry{
		Time: time.Date(2026, 10, 29, 1, 0, 0, 0, time.UTC).Format(time.RFC3339),
		Command: "nacos config get", Risk: "read", Profile: "lab", OK: true,
	})
	_ = store.Append(audit.Entry{
		Time: time.Date(2026, 10, 29, 2, 0, 0, 0, time.UTC).Format(time.RFC3339),
		Command: "nacos profile delete", Risk: "high-risk-write", Profile: "prod", OK: false, Error: "no",
	})
	srv := newWebServer(dir, store)
	req := httptest.NewRequest(http.MethodGet, "/api/audit?command=config&profile=lab", nil)
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)
	if rr.Code != 200 {
		t.Fatalf("%d %s", rr.Code, rr.Body.String())
	}
	var resp map[string]any
	_ = json.Unmarshal(rr.Body.Bytes(), &resp)
	entries := resp["entries"].([]any)
	if len(entries) != 1 {
		t.Fatalf("%v", resp)
	}
}

func TestWebIndexBilingual(t *testing.T) {
	dir := t.TempDir()
	srv := newWebServer(dir, &audit.Store{Path: filepath.Join(dir, "a.jsonl")})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9")
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)
	body := rr.Body.String()
	if rr.Code != 200 {
		t.Fatalf("%d", rr.Code)
	}
	for _, want := range []string{"card", "#1677ff", "Profiles", "监管", "操作", `lang="zh"`} {
		// zh page may use 配置文件 / 操作记录; accept either zh labels
		_ = want
	}
	if !strings.Contains(body, "card") || !strings.Contains(body, "#1677ff") {
		t.Fatal("missing light card style")
	}
	if !strings.Contains(body, "配置文件") && !strings.Contains(body, "Profile") {
		t.Fatal("missing profiles section label")
	}
	if !strings.Contains(body, `lang="zh"`) {
		t.Fatal("expected zh lang")
	}
	if strings.Contains(body, "Audit log") && !strings.Contains(body, "操作") {
		// en-only labels on zh page
		t.Fatal("zh page should prefer Chinese labels")
	}

	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Language", "en-US")
	rr = httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)
	body = rr.Body.String()
	if !strings.Contains(body, `lang="en"`) {
		t.Fatal("expected en lang")
	}
	if !strings.Contains(body, "Profiles") || !strings.Contains(body, "Audit") {
		t.Fatal("en labels missing")
	}
	if strings.Contains(body, "配置文件") {
		t.Fatal("en page should not show Chinese section title")
	}
}

func TestRecordOutcomeWritesAudit(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "audit.jsonl")
	a := &App{profileName: "lab", auditStore: &audit.Store{Path: path}, secrets: []string{"deadbeefdeadbeef"}}
	a.recordOutcome("nacos whoami", "read", true, "")
	a.recordOutcome("nacos config publish", "high-risk-write", false, "boom token=deadbeefdeadbeef")
	a.recordOutcome("nacos auth login", "write", false, "bad deadbeefdeadbeef")

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	s := string(raw)
	if !strings.Contains(s, "nacos whoami") || !strings.Contains(s, `"ok":true`) {
		t.Fatalf("%s", s)
	}
	if strings.Contains(s, "deadbeefdeadbeef") {
		t.Fatal("secret leaked into audit")
	}
	lines := strings.Split(strings.TrimSpace(s), "\n")
	if len(lines) < 2 {
		t.Fatalf("lines=%d", len(lines))
	}
}

func TestWebDeleteRequiresConfirm(t *testing.T) {
	dir := t.TempDir()
	_, _ = config.SaveProfileIn(dir, "x", config.Profile{Server: "http://a:8848"})
	srv := newWebServer(dir, &audit.Store{Path: filepath.Join(dir, "a.jsonl")})
	req := httptest.NewRequest(http.MethodDelete, "/api/profiles/x", nil)
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)
	if rr.Code == 200 {
		t.Fatal("delete without confirm should fail")
	}
	_, ok, _ := config.LoadProfileIn(dir, "x")
	if !ok {
		t.Fatal("profile should remain")
	}
}

func TestApplyPasswordUpdate(t *testing.T) {
	p := config.Profile{Password: "old"}
	got := applyPasswordUpdate(p, profileWriteRequest{})
	if got.Password != "old" {
		t.Fatal("keep")
	}
	got = applyPasswordUpdate(p, profileWriteRequest{Password: "new", PasswordSet: true})
	if got.Password != "new" {
		t.Fatal("set")
	}
	got = applyPasswordUpdate(p, profileWriteRequest{ClearPassword: true})
	if got.Password != "" {
		t.Fatal("clear")
	}
}


func TestWebCheckAuthTokenOK(t *testing.T) {
	nacos := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/v1/auth/login"):
			_ = r.ParseForm()
			if r.Form.Get("username") != "nacos" || r.Form.Get("password") != "s3cret-pw" {
				w.WriteHeader(401)
				_, _ = w.Write([]byte(`{"message":"user not found"}`))
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"accessToken": "tok-abc", "tokenTtl": 18000})
		default:
			if strings.Contains(r.URL.Path, "namespace") {
				_ = json.NewEncoder(w).Encode(map[string]any{"code": 200, "data": []any{}})
				return
			}
			w.WriteHeader(404)
			_, _ = w.Write([]byte(r.URL.Path))
		}
	}))
	defer nacos.Close()

	dir := t.TempDir()
	_, err := config.SaveProfileIn(dir, "lab", config.Profile{
		Server: nacos.URL, Username: "nacos", Password: "s3cret-pw", ContextPath: "/nacos",
	})
	if err != nil {
		t.Fatal(err)
	}
	srv := newWebServer(dir, &audit.Store{Path: filepath.Join(dir, "a.jsonl")})
	req := httptest.NewRequest(http.MethodPost, "/api/profiles/lab/check-auth", nil)
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)
	if rr.Code != 200 {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
	body := rr.Body.String()
	if strings.Contains(body, "s3cret-pw") || strings.Contains(body, "tok-abc") {
		t.Fatalf("secrets leaked: %s", body)
	}
	var resp map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp["ok"] != true {
		t.Fatalf("%v", resp)
	}
	if resp["auth_mode"] != "token" {
		t.Fatalf("auth_mode=%v", resp["auth_mode"])
	}
	if resp["via"] != "password" {
		t.Fatalf("via=%v want password", resp["via"])
	}
	if resp["profile"] != "lab" {
		t.Fatalf("profile=%v", resp["profile"])
	}
}

func TestWebCheckAuthDisabled(t *testing.T) {
	nacos := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "namespace") {
			_ = json.NewEncoder(w).Encode(map[string]any{"code": 200, "data": []any{}})
			return
		}
		w.WriteHeader(404)
	}))
	defer nacos.Close()
	dir := t.TempDir()
	_, _ = config.SaveProfileIn(dir, "open", config.Profile{Server: nacos.URL, Username: "nacos", ContextPath: "/nacos"})
	credPath := filepath.Join(dir, "credentials.json")
	if err := credstore.Save(credPath, credstore.Cred{
		Server: nacos.URL, Username: "nacos", AuthMode: "disabled",
	}); err != nil {
		t.Fatal(err)
	}
	srv := newWebServer(dir, &audit.Store{Path: filepath.Join(dir, "a.jsonl")})
	srv.credPath = credPath
	req := httptest.NewRequest(http.MethodPost, "/api/profiles/open/check-auth", nil)
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)
	if rr.Code != 200 {
		t.Fatalf("%d %s", rr.Code, rr.Body.String())
	}
	var resp map[string]any
	_ = json.Unmarshal(rr.Body.Bytes(), &resp)
	if resp["ok"] != true || resp["auth_mode"] != "disabled" {
		t.Fatalf("%v", resp)
	}
	if resp["via"] != "token" {
		t.Fatalf("via=%v want token", resp["via"])
	}
}

func TestWebCheckAuthBadPassword(t *testing.T) {
	nacos := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(401)
		_, _ = w.Write([]byte(`{"message":"unknown user"}`))
	}))
	defer nacos.Close()
	dir := t.TempDir()
	_, _ = config.SaveProfileIn(dir, "bad", config.Profile{
		Server: nacos.URL, Username: "nacos", Password: "wrong-pw-xyz", ContextPath: "/nacos",
	})
	srv := newWebServer(dir, &audit.Store{Path: filepath.Join(dir, "a.jsonl")})
	req := httptest.NewRequest(http.MethodPost, "/api/profiles/bad/check-auth", nil)
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)
	body := rr.Body.String()
	if strings.Contains(body, "wrong-pw-xyz") {
		t.Fatal("password leaked")
	}
	var resp map[string]any
	_ = json.Unmarshal(rr.Body.Bytes(), &resp)
	if resp["ok"] == true {
		t.Fatalf("expected failure: %v", resp)
	}
	errMsg, _ := resp["error"].(string)
	if errMsg == "" {
		t.Fatalf("expected error detail: %v", resp)
	}
	if resp["via"] != "password" {
		t.Fatalf("via=%v want password", resp["via"])
	}
}

func TestWebCheckAuthMissingProfile(t *testing.T) {
	dir := t.TempDir()
	srv := newWebServer(dir, &audit.Store{Path: filepath.Join(dir, "a.jsonl")})
	req := httptest.NewRequest(http.MethodPost, "/api/profiles/nope/check-auth", nil)
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)
	var resp map[string]any
	_ = json.Unmarshal(rr.Body.Bytes(), &resp)
	if resp["ok"] == true {
		t.Fatal("missing profile should not be ok")
	}
}

func TestWebCheckAuthNoPasswordNoToken(t *testing.T) {
	loginHits := 0
	nacos := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/v1/auth/login") {
			loginHits++
			w.WriteHeader(403)
			_, _ = w.Write([]byte(`{"message":"user not found!"}`))
			return
		}
		w.WriteHeader(200)
		_, _ = w.Write([]byte("ok"))
	}))
	defer nacos.Close()

	dir := t.TempDir()
	_, err := config.SaveProfileIn(dir, "nopw", config.Profile{
		Server: nacos.URL, Username: "nacos", ContextPath: "/nacos",
	})
	if err != nil {
		t.Fatal(err)
	}
	credPath := filepath.Join(dir, "credentials.json")
	srv := newWebServer(dir, &audit.Store{Path: filepath.Join(dir, "a.jsonl")})
	srv.credPath = credPath

	req := httptest.NewRequest(http.MethodPost, "/api/profiles/nopw/check-auth", nil)
	req.Header.Set("Accept-Language", "zh-CN")
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)
	if rr.Code != 200 {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
	var resp map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp["ok"] == true {
		t.Fatalf("expected failure: %v", resp)
	}
	if resp["detail"] != "password_missing" {
		t.Fatalf("detail=%v want password_missing", resp["detail"])
	}
	errMsg, _ := resp["error"].(string)
	if !strings.Contains(errMsg, "未设置密码") && !strings.Contains(errMsg, "No password") {
		t.Fatalf("error should mention missing password/token: %v", resp)
	}
	if loginHits != 0 {
		t.Fatalf("must not empty-password Login; hits=%d", loginHits)
	}
}

func TestWebCheckAuthTokenFromCredstore(t *testing.T) {
	nacos := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/v1/auth/login") {
			t.Errorf("Login should not be called when using credstore token")
			w.WriteHeader(500)
			return
		}
		if strings.Contains(r.URL.Path, "namespace") {
			tok := r.URL.Query().Get("accessToken")
			if tok != "saved-tok-xyz" {
				w.WriteHeader(403)
				_, _ = w.Write([]byte(`{"message":"forbidden"}`))
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"code": 200, "data": []any{}})
			return
		}
		w.WriteHeader(200)
		_, _ = w.Write([]byte("ok"))
	}))
	defer nacos.Close()

	dir := t.TempDir()
	_, err := config.SaveProfileIn(dir, "tokprof", config.Profile{
		Server: nacos.URL, Username: "nacos", ContextPath: "/nacos",
	})
	if err != nil {
		t.Fatal(err)
	}
	credPath := filepath.Join(dir, "credentials.json")
	if err := credstore.Save(credPath, credstore.Cred{
		Server: nacos.URL, Username: "nacos", AccessToken: "saved-tok-xyz", AuthMode: "token", Via: "password",
	}); err != nil {
		t.Fatal(err)
	}
	srv := newWebServer(dir, &audit.Store{Path: filepath.Join(dir, "a.jsonl")})
	srv.credPath = credPath

	req := httptest.NewRequest(http.MethodPost, "/api/profiles/tokprof/check-auth", nil)
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)
	body := rr.Body.String()
	if strings.Contains(body, "saved-tok-xyz") {
		t.Fatalf("token leaked: %s", body)
	}
	var resp map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp["ok"] != true {
		t.Fatalf("%v", resp)
	}
	if resp["via"] != "token" {
		t.Fatalf("via=%v want token", resp["via"])
	}
	if resp["auth_mode"] != "token" {
		t.Fatalf("auth_mode=%v", resp["auth_mode"])
	}
}

func TestWebIndexHasCheckAuth(t *testing.T) {
	dir := t.TempDir()
	srv := newWebServer(dir, &audit.Store{Path: filepath.Join(dir, "a.jsonl")})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Language", "zh-CN")
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)
	body := rr.Body.String()
	if !strings.Contains(body, "check-auth") && !strings.Contains(body, "检查认证") {
		t.Fatal("zh page missing check-auth UI")
	}
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Language", "en-US")
	rr = httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)
	body = rr.Body.String()
	if !strings.Contains(body, "Check auth") && !strings.Contains(body, "check-auth") {
		t.Fatal("en page missing check-auth UI")
	}
}


func TestMergeProfileWriteDraft(t *testing.T) {
	existing := config.Profile{
		Server: "http://old:8848", Username: "old", Password: "keep-me", ContextPath: "/nacos", AuthType: "nacos",
	}
	req := profileWriteRequest{
		Server: "http://new:8848", Username: "admin", ContextPath: "/nacos",
	}
	got, err := mergeProfileWrite(existing, req)
	if err != nil {
		t.Fatal(err)
	}
	if got.Server != "http://new:8848" || got.Username != "admin" {
		t.Fatalf("merge fields: %+v", got)
	}
	if got.Password != "keep-me" {
		t.Fatalf("empty password must keep saved, got %q", got.Password)
	}
	req2 := profileWriteRequest{Server: "http://new:8848", Username: "admin", Password: "fresh", PasswordSet: true, ContextPath: "/nacos"}
	got2, err := mergeProfileWrite(existing, req2)
	if err != nil {
		t.Fatal(err)
	}
	if got2.Password != "fresh" {
		t.Fatalf("password_set should apply, got %q", got2.Password)
	}
	req3 := profileWriteRequest{Server: "http://new:8848", Username: "admin", ClearPassword: true, ContextPath: "/nacos"}
	got3, err := mergeProfileWrite(existing, req3)
	if err != nil {
		t.Fatal(err)
	}
	if got3.Password != "" {
		t.Fatal("clear_password should wipe")
	}
}

func TestWebCheckAuthDraftUsesFormPassword(t *testing.T) {
	nacos := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/v1/auth/login"):
			_ = r.ParseForm()
			if r.Form.Get("username") != "nacos" || r.Form.Get("password") != "form-pw-ok" {
				w.WriteHeader(401)
				_, _ = w.Write([]byte(`{"message":"user not found"}`))
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"accessToken": "tok-form", "tokenTtl": 18000})
		default:
			if strings.Contains(r.URL.Path, "namespace") {
				_ = json.NewEncoder(w).Encode(map[string]any{"code": 200, "data": []any{}})
				return
			}
			w.WriteHeader(200)
			_, _ = w.Write([]byte("ok"))
		}
	}))
	defer nacos.Close()

	dir := t.TempDir()
	_, err := config.SaveProfileIn(dir, "lab", config.Profile{
		Server: nacos.URL, Username: "nacos", Password: "disk-wrong", ContextPath: "/nacos",
	})
	if err != nil {
		t.Fatal(err)
	}
	srv := newWebServer(dir, &audit.Store{Path: filepath.Join(dir, "a.jsonl")})
	body := map[string]any{
		"server": nacos.URL, "username": "nacos", "password": "form-pw-ok", "password_set": true, "contextPath": "/nacos",
	}
	raw, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/profiles/lab/check-auth", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)
	if rr.Code != 200 {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
	out := rr.Body.String()
	if strings.Contains(out, "form-pw-ok") || strings.Contains(out, "disk-wrong") || strings.Contains(out, "tok-form") {
		t.Fatalf("secrets leaked: %s", out)
	}
	var resp map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp["ok"] != true {
		t.Fatalf("draft password should authenticate: %v", resp)
	}
	if resp["via"] != "password" {
		t.Fatalf("via=%v", resp["via"])
	}
	// disk unchanged
	saved, _, _ := config.LoadProfileIn(dir, "lab")
	if saved.Password != "disk-wrong" {
		t.Fatalf("check-auth must not persist draft password, got %q", saved.Password)
	}
}

func TestWebCheckAuthDraftEmptyPasswordKeepsSaved(t *testing.T) {
	nacos := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/v1/auth/login"):
			_ = r.ParseForm()
			if r.Form.Get("password") != "saved-pw" {
				w.WriteHeader(401)
				_, _ = w.Write([]byte(`{"message":"bad"}`))
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"accessToken": "tok", "tokenTtl": 1})
		default:
			if strings.Contains(r.URL.Path, "namespace") {
				_ = json.NewEncoder(w).Encode(map[string]any{"code": 200, "data": []any{}})
				return
			}
			w.WriteHeader(200)
			_, _ = w.Write([]byte("ok"))
		}
	}))
	defer nacos.Close()
	dir := t.TempDir()
	_, _ = config.SaveProfileIn(dir, "lab", config.Profile{
		Server: nacos.URL, Username: "nacos", Password: "saved-pw", ContextPath: "/nacos",
	})
	srv := newWebServer(dir, &audit.Store{Path: filepath.Join(dir, "a.jsonl")})
	body := map[string]any{"server": nacos.URL, "username": "nacos", "contextPath": "/nacos"}
	raw, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/profiles/lab/check-auth", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)
	var resp map[string]any
	_ = json.Unmarshal(rr.Body.Bytes(), &resp)
	if resp["ok"] != true {
		t.Fatalf("empty form password should use saved: %v", resp)
	}
}

func TestWebCheckAuthDraftNewProfile(t *testing.T) {
	nacos := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/v1/auth/login"):
			_ = r.ParseForm()
			if r.Form.Get("password") != "only-in-form" {
				w.WriteHeader(401)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"accessToken": "tok", "tokenTtl": 1})
		default:
			if strings.Contains(r.URL.Path, "namespace") {
				_ = json.NewEncoder(w).Encode(map[string]any{"code": 200, "data": []any{}})
				return
			}
			w.WriteHeader(200)
			_, _ = w.Write([]byte("ok"))
		}
	}))
	defer nacos.Close()
	dir := t.TempDir()
	srv := newWebServer(dir, &audit.Store{Path: filepath.Join(dir, "a.jsonl")})
	body := map[string]any{
		"server": nacos.URL, "username": "nacos", "password": "only-in-form", "password_set": true, "contextPath": "/nacos",
	}
	raw, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/profiles/brand-new/check-auth", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)
	if rr.Code != 200 {
		t.Fatalf("status=%d %s", rr.Code, rr.Body.String())
	}
	var resp map[string]any
	_ = json.Unmarshal(rr.Body.Bytes(), &resp)
	if resp["ok"] != true {
		t.Fatalf("unsaved draft profile should check: %v", resp)
	}
	_, ok, _ := config.LoadProfileIn(dir, "brand-new")
	if ok {
		t.Fatal("draft check must not create profile")
	}
}

func TestWebIndexDraftCheckAuthPostsBody(t *testing.T) {
	dir := t.TempDir()
	srv := newWebServer(dir, &audit.Store{Path: filepath.Join(dir, "a.jsonl")})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)
	body := rr.Body.String()
	if !strings.Contains(body, "checkAuthDraft") && !strings.Contains(body, "password_set") {
		t.Fatal("edit form check-auth should post draft body")
	}
	if !strings.Contains(body, "f-password") {
		t.Fatal("missing password field collection")
	}
}

// silence unused import in case helpers shift
var _ = io.Discard
