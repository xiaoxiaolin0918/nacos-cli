package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/xiaoxiaolin0918/nacos-cli/internal/client"
	"github.com/xiaoxiaolin0918/nacos-cli/internal/config"
	"github.com/xiaoxiaolin0918/nacos-cli/internal/credstore"
)

// authCheckResult is a redacted auth probe outcome for the local web console.
type authCheckResult struct {
	OK        bool   `json:"ok"`
	Profile   string `json:"profile"`
	Server    string `json:"server,omitempty"`
	Username  string `json:"username,omitempty"`
	AuthMode  string `json:"auth_mode,omitempty"`
	Reachable bool   `json:"reachable"`
	Via       string `json:"via,omitempty"` // password | token | ""
	Detail    string `json:"detail,omitempty"`
	Error     string `json:"error,omitempty"`
}

func redactSecrets(msg string, secrets ...string) string {
	out := msg
	for _, s := range secrets {
		s = strings.TrimSpace(s)
		if s == "" || len(s) < 3 {
			continue
		}
		out = strings.ReplaceAll(out, s, "***")
	}
	return out
}

func settingsFromProfile(p config.Profile) config.Settings {
	server := strings.TrimRight(strings.TrimSpace(p.Server), "/")
	if server == "" && strings.TrimSpace(p.Host) != "" {
		port := p.Port
		if port <= 0 {
			port = 8848
		}
		server = "http://" + strings.TrimSpace(p.Host) + ":" + itoa(port)
	}
	ctx := strings.TrimSpace(p.ContextPath)
	if ctx == "" {
		ctx = "/nacos"
	}
	return config.Settings{
		Server:      server,
		Username:    strings.TrimSpace(p.Username),
		Password:    p.Password,
		Namespace:   p.Namespace,
		ContextPath: ctx,
	}
}

func normalizeServerURL(s string) string {
	return strings.TrimRight(strings.TrimSpace(s), "/")
}

func credMatchesProfile(cred credstore.Cred, settings config.Settings) bool {
	credServer := normalizeServerURL(cred.Server)
	profServer := normalizeServerURL(settings.Server)
	if credServer == "" || profServer == "" {
		return false
	}
	if !strings.EqualFold(credServer, profServer) {
		return false
	}
	cu := strings.TrimSpace(cred.Username)
	pu := strings.TrimSpace(settings.Username)
	if cu != "" && pu != "" && !strings.EqualFold(cu, pu) {
		return false
	}
	return true
}

func passwordMissingMsg(lang string) string {
	if lang == "zh" {
		return "未设置密码且无本地登录 token，请先执行 nacos auth login 或在 Profile 中填写密码"
	}
	return "No password set and no local login token; run nacos auth login or set a password on the profile"
}

func resolveCredPath(explicit string) string {
	if strings.TrimSpace(explicit) != "" {
		return explicit
	}
	p, err := config.CredentialsPath()
	if err != nil {
		return ""
	}
	return p
}

// checkProfileAuth verifies profile auth without leaking secrets.
// Prefer profile password Login; otherwise reuse a matching local credstore token and Probe.
// Never posts an empty-password Login (Nacos often replies with a misleading "user not found").
func checkProfileAuth(ctx context.Context, name string, p config.Profile, httpClient *http.Client, credPath, lang string) authCheckResult {
	settings := settingsFromProfile(p)
	res := authCheckResult{Profile: name, Username: settings.Username, Server: settings.Server}
	if strings.TrimSpace(settings.Server) == "" {
		res.Error = "server not set"
		return res
	}
	c := client.New(settings)
	if httpClient != nil {
		c.HTTP = httpClient
	} else {
		c.HTTP = &http.Client{Timeout: 12 * time.Second}
	}

	if _, err := c.Ping(ctx); err != nil {
		res.Reachable = false
		res.Error = redactSecrets("unreachable: "+err.Error(), p.Password)
		return res
	}
	res.Reachable = true

	pass := p.Password
	if strings.TrimSpace(pass) != "" {
		user := settings.Username
		if user == "" {
			user = "nacos"
			res.Username = user
		}
		login, err := c.Login(ctx, user, pass)
		if err != nil {
			res.Via = "password"
			res.Error = redactSecrets(err.Error(), p.Password, login.AccessTok)
			return res
		}
		res.Via = "password"
		res.AuthMode = login.AuthMode
		if login.AuthMode == "disabled" {
			res.OK = true
			res.Detail = "AUTH_DISABLED"
			return res
		}
		return finishProbe(ctx, c, &res, p.Password, login.AccessTok)
	}

	// No password in profile: try local credstore token (same path as CLI).
	path := resolveCredPath(credPath)
	if path != "" {
		cred, ok, err := credstore.Load(path)
		if err == nil && ok && credMatchesProfile(cred, settings) {
			if cred.AuthMode == "disabled" || cred.AccessToken == "AUTH_DISABLED" {
				c.ApplyToken("", "disabled")
				res.Via = "token"
				res.AuthMode = "disabled"
				res.OK = true
				res.Detail = "AUTH_DISABLED"
				return res
			}
			if strings.TrimSpace(cred.AccessToken) != "" {
				mode := cred.AuthMode
				if mode == "" {
					mode = "token"
				}
				c.ApplyToken(cred.AccessToken, mode)
				res.Via = "token"
				res.AuthMode = mode
				if settings.Username == "" && cred.Username != "" {
					res.Username = cred.Username
				}
				return finishProbe(ctx, c, &res, cred.AccessToken)
			}
		}
	}

	res.Detail = "password_missing"
	res.Error = passwordMissingMsg(lang)
	return res
}

func finishProbe(ctx context.Context, c *client.Client, res *authCheckResult, secrets ...string) authCheckResult {
	probe := c.Probe(ctx)
	ok, _ := probe["ok"].(bool)
	if !ok {
		detail := "token accepted but namespace probe failed"
		if e, has := probe["error"].(string); has && e != "" {
			detail = redactSecrets(e, append(secrets, c.Token)...)
		}
		res.Error = detail
		res.Detail = "login_ok_probe_failed"
		return *res
	}
	res.OK = true
	if res.Detail == "" {
		res.Detail = "token valid"
	}
	return *res
}

func (s *webServer) handleCheckAuth(w http.ResponseWriter, r *http.Request, name string) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	existing, ok, err := config.LoadProfileIn(s.dir, name)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "profile": name, "error": err.Error()})
		return
	}

	p := existing
	usedDraft := false
	if r.Method == http.MethodPost && r.Body != nil {
		raw, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "profile": name, "error": "read body: " + err.Error()})
			return
		}
		if len(bytes.TrimSpace(raw)) > 0 {
			var req profileWriteRequest
			if err := json.Unmarshal(raw, &req); err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "profile": name, "error": "invalid json: " + err.Error()})
				return
			}
			base := existing
			if !ok {
				base = config.Profile{}
			}
			merged, err := mergeProfileWrite(base, req)
			if err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "profile": name, "error": err.Error()})
				return
			}
			p = merged
			usedDraft = true
		}
	}
	if !usedDraft && !ok {
		writeJSON(w, http.StatusNotFound, map[string]any{"ok": false, "profile": name, "error": "not found"})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()
	lang := webLang(r.Header.Get("Accept-Language"))
	result := checkProfileAuth(ctx, name, p, nil, s.credPath, lang)
	cmdName := "web.profile.check-auth"
	if usedDraft {
		cmdName = "web.profile.check-auth-draft"
	}
	s.note(cmdName, "read", name, result.OK, result.Error)
	writeJSON(w, http.StatusOK, result)
}
