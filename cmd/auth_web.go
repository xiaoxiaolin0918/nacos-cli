package cmd

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"html"
	"io"
	"net"
	"net/http"
	"net/url"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"

	"github.com/xiaoxiaolin0918/nacos-cli/internal/credstore"
	"github.com/xiaoxiaolin0918/nacos-cli/internal/risk"
)

const webLoginTimeout = 120 * time.Second

func (a *App) runWebLogin(cmd *cobra.Command) error {
	token := make([]byte, 16)
	if _, err := rand.Read(token); err != nil {
		return a.fail(cmd, risk.Write, err)
	}
	pathToken := hex.EncodeToString(token)
	formPath := "/cli-login/" + pathToken + "/"

	type outcome struct {
		user string
		res  map[string]any
		err  error
	}
	ch := make(chan outcome, 1)
	var once sync.Once

	mux := http.NewServeMux()
	mux.HandleFunc(formPath, func(w http.ResponseWriter, r *http.Request) {
		lang := webLang(r.Header.Get("Accept-Language"))
		switch r.Method {
		case http.MethodGet:
			body := formHTML(a.settings.BaseURL(), a.settings.Username, "", lang)
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Header().Set("Cache-Control", "no-store")
			_, _ = io.WriteString(w, body)
		case http.MethodPost:
			_ = r.ParseForm()
			user := strings.TrimSpace(r.Form.Get("username"))
			pass := r.Form.Get("password")
			a.track(pass)
			res, err := a.client.Login(r.Context(), user, pass)
			if err != nil {
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				_, _ = io.WriteString(w, failHTML(err.Error(), lang))
				once.Do(func() { ch <- outcome{err: err} })
				return
			}
			cred := credstore.Cred{
				Server: a.settings.Server, Username: user, AuthMode: res.AuthMode, Via: "web",
			}
			if res.AuthMode == "token" {
				cred.AccessToken = res.AccessTok
				a.track(res.AccessTok)
			}
			if err := credstore.Save(a.credPath, cred); err != nil {
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				_, _ = io.WriteString(w, failHTML(err.Error(), lang))
				once.Do(func() { ch <- outcome{err: err} })
				return
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = io.WriteString(w, okHTML(a.settings.Server, res.AuthMode, lang))
			once.Do(func() {
				ch <- outcome{user: user, res: map[string]any{"auth_mode": res.AuthMode, "has_token": res.HasToken}}
			})
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return a.fail(cmd, risk.Write, err)
	}
	host, port, _ := net.SplitHostPort(ln.Addr().String())
	if host != "127.0.0.1" && host != "::1" {
		_ = ln.Close()
		return a.fail(cmd, risk.Write, fmt.Errorf("refusing non-loopback bind %s", host))
	}
	formURL := fmt.Sprintf("http://127.0.0.1:%s%s", port, formPath)

	srv := &http.Server{Handler: mux}
	go func() { _ = srv.Serve(ln) }()
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = srv.Shutdown(ctx)
	}()

	_ = openBrowser(formURL)
	fmt.Fprintf(cmd.ErrOrStderr(), "若浏览器未打开请访问 / Open %s if the browser did not launch\n", formURL)

	select {
	case out := <-ch:
		if out.err != nil {
			return a.fail(cmd, risk.Write, out.err)
		}
		st := credstore.Status(a.credPath)
		st["ok"] = true
		st["via"] = "web"
		return a.ok(cmd, st)
	case <-time.After(webLoginTimeout):
		return a.fail(cmd, risk.Write, errorsTimeout())
	}
}

func errorsTimeout() error {
	return &strErr{s: "web login timed out; local listener closed"}
}

// webLang picks "zh" when Accept-Language mentions a Chinese tag; otherwise "en".
func webLang(acceptLanguage string) string {
	low := strings.ToLower(acceptLanguage)
	if strings.Contains(low, "zh") {
		return "zh"
	}
	return "en"
}

const webLoginCSS = `
*{box-sizing:border-box}
body{margin:0;min-height:100vh;display:flex;align-items:center;justify-content:center;
background:#f5f7fa;font-family:system-ui,-apple-system,"Segoe UI",sans-serif;color:#1f1f1f}
.card{background:#fff;border-radius:12px;box-shadow:0 8px 24px rgba(15,23,42,.08);
padding:28px 28px 24px;width:100%;max-width:420px;margin:24px}
h1{font-size:1.35rem;margin:0 0 12px;font-weight:600}
.hint{color:#64748b;font-size:13px;margin:0 0 8px;line-height:1.45}
label{display:block;margin:14px 0 6px;font-size:14px;font-weight:500}
input{width:100%;padding:10px 12px;border:1px solid #d9d9d9;border-radius:8px;font-size:14px}
input:focus{outline:none;border-color:#1677ff;box-shadow:0 0 0 2px rgba(22,119,255,.15)}
button.primary{margin-top:20px;width:100%;padding:11px 16px;border:0;border-radius:8px;
background:#1677ff;color:#fff;font-size:15px;font-weight:600;cursor:pointer}
button.primary:hover{background:#4096ff}
.err{color:#b00020;background:#fff2f0;border:1px solid #ffccc7;border-radius:8px;
padding:10px 12px;margin:12px 0;font-size:13px}
code{font-size:12px;word-break:break-all}
`

func formHTML(serverHint, username, errMsg, lang string) string {
	title, target, passHint, userLabel, passLabel, submit :=
		"nacos-cli Login", "Target", "Password stays in this form; the CLI never prints it.",
		"Username", "Password", "Sign in"
	if lang == "zh" {
		title, target, passHint, userLabel, passLabel, submit =
			"nacos-cli 登录", "目标", "密码仅留在本表单；CLI 不会打印密码。",
			"用户名", "密码", "登录"
	}
	errBlock := ""
	if errMsg != "" {
		errBlock = `<p class="err">` + html.EscapeString(errMsg) + `</p>`
	}
	return `<!doctype html>
<html lang="` + html.EscapeString(lang) + `"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<title>` + html.EscapeString(title) + `</title>
<style>` + webLoginCSS + `</style></head><body>
<div class="card">
<h1>` + html.EscapeString(title) + `</h1>
<p class="hint">` + html.EscapeString(target) + `: <code>` + html.EscapeString(serverHint) + `</code></p>
<p class="hint">` + html.EscapeString(passHint) + `</p>
` + errBlock + `
<form method="POST">
<label>` + html.EscapeString(userLabel) + `</label>
<input name="username" value="` + html.EscapeString(username) + `" autocomplete="username" required>
<label>` + html.EscapeString(passLabel) + `</label>
<input name="password" type="password" autocomplete="current-password">
<button class="primary" type="submit">` + html.EscapeString(submit) + `</button>
</form>
</div>
</body></html>`
}

func okHTML(server, authMode, lang string) string {
	title, serverL, modeL, closeHint :=
		"Login saved", "Server", "Auth mode", "You can close this tab."
	if lang == "zh" {
		title, serverL, modeL, closeHint =
			"登录已保存", "服务器", "认证模式", "可关闭此标签页。"
	}
	return `<!doctype html>
<html lang="` + html.EscapeString(lang) + `"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<title>` + html.EscapeString(title) + `</title>
<style>` + webLoginCSS + `</style></head><body>
<div class="card">
<h1>` + html.EscapeString(title) + `</h1>
<p class="hint">` + html.EscapeString(serverL) + `: <code>` + html.EscapeString(server) + `</code></p>
<p class="hint">` + html.EscapeString(modeL) + `: <code>` + html.EscapeString(authMode) + `</code></p>
<p class="hint">` + html.EscapeString(closeHint) + `</p>
</div>
</body></html>`
}

func failHTML(message, lang string) string {
	title := "Login failed"
	if lang == "zh" {
		title = "登录失败"
	}
	return `<!doctype html>
<html lang="` + html.EscapeString(lang) + `"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<title>` + html.EscapeString(title) + `</title>
<style>` + webLoginCSS + `</style></head><body>
<div class="card">
<h1>` + html.EscapeString(title) + `</h1>
<p class="err">` + html.EscapeString(message) + `</p>
</div>
</body></html>`
}

func openBrowser(u string) error {
	var c *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		c = exec.Command("rundll32", "url.dll,FileProtocolHandler", u)
	case "darwin":
		c = exec.Command("open", u)
	default:
		c = exec.Command("xdg-open", u)
	}
	return c.Start()
}

// silence unused url import helper if needed
var _ = url.Values{}