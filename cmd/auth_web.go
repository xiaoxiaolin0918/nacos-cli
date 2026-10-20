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
		switch r.Method {
		case http.MethodGet:
			body := formHTML(a.settings.BaseURL(), a.settings.Username, "")
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
				_, _ = io.WriteString(w, failHTML(err.Error()))
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
				_, _ = io.WriteString(w, failHTML(err.Error()))
				once.Do(func() { ch <- outcome{err: err} })
				return
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = io.WriteString(w, okHTML(a.settings.Server, res.AuthMode))
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
	fmt.Fprintf(cmd.ErrOrStderr(), "Open %s if the browser did not launch\n", formURL)

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

func formHTML(serverHint, username, errMsg string) string {
	errBlock := ""
	if errMsg != "" {
		errBlock = `<p style="color:#b00020">` + html.EscapeString(errMsg) + `</p>`
	}
	return `<!doctype html>
<html><head><meta charset="utf-8"><title>nacos-cli login</title>
<style>
body{font-family:system-ui,sans-serif;max-width:420px;margin:40px auto;padding:0 16px}
label{display:block;margin:12px 0 4px}
input{width:100%;padding:8px;box-sizing:border-box}
button{margin-top:16px;padding:10px 16px}
.hint{color:#555;font-size:13px}
</style></head><body>
<h1>nacos-cli web login</h1>
<p class="hint">Target: <code>` + html.EscapeString(serverHint) + `</code></p>
<p class="hint">Password stays in this form; the CLI never prints it.</p>
` + errBlock + `
<form method="POST">
<label>Username</label>
<input name="username" value="` + html.EscapeString(username) + `" autocomplete="username" required>
<label>Password</label>
<input name="password" type="password" autocomplete="current-password">
<button type="submit">Sign in</button>
</form>
</body></html>`
}

func okHTML(server, authMode string) string {
	return `<!doctype html><html><head><meta charset="utf-8"><title>nacos-cli login ok</title></head>
<body style="font-family:system-ui;margin:40px">
<h1>Login saved</h1>
<p>Server: <code>` + html.EscapeString(server) + `</code></p>
<p>Auth mode: <code>` + html.EscapeString(authMode) + `</code></p>
<p>You can close this tab.</p>
</body></html>`
}

func failHTML(message string) string {
	return `<!doctype html><html><head><meta charset="utf-8"><title>nacos-cli login failed</title></head>
<body style="font-family:system-ui;margin:40px">
<h1>Login failed</h1>
<p style="color:#b00020">` + html.EscapeString(message) + `</p>
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
