package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/xiaoxiaolin0918/nacos-cli/internal/credstore"
)

func envTruthy(name string) bool {
	v := strings.TrimSpace(strings.ToLower(os.Getenv(name)))
	switch v {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

func (a *App) loginDir() string {
	return credstore.DirFromCredentialsPath(a.credPath)
}

// wireAutoRenew attaches a one-shot RenewOnce when auto-login is enabled.
func (a *App) wireAutoRenew() {
	if a.client == nil || !a.autoLogin || a.dryRun {
		return
	}
	a.client.RenewOnce = func(ctx context.Context) error {
		err := a.silentLogin(ctx)
		if err == nil {
			a.traceRenewed = true
		}
		return err
	}
}

// resolveSilentLoginCreds returns an atomic username/password pair.
// 1) Both NACOS_USERNAME and NACOS_PASSWORD -> env pair
// 2) Else whole login.cred
// Never mixes username from one source with password from another.
func (a *App) resolveSilentLoginCreds() (user, pass string, err error) {
	envUser := strings.TrimSpace(os.Getenv("NACOS_USERNAME"))
	envPass := os.Getenv("NACOS_PASSWORD")
	if envUser != "" && envPass != "" {
		return envUser, envPass, nil
	}
	if c, ok, loadErr := credstore.LoadLogin(a.loginDir()); loadErr == nil && ok {
		u := strings.TrimSpace(c.Username)
		if u != "" && c.Password != "" {
			return u, c.Password, nil
		}
	}
	return "", "", errors.New("auto-login requires NACOS_USERNAME+NACOS_PASSWORD or remembered login.cred (auth login --remember)")
}

func (a *App) silentLogin(ctx context.Context) error {
	user, pass, err := a.resolveSilentLoginCreds()
	if err != nil {
		return fmt.Errorf("auto-login failed: %w; run: nacos auth login --remember", err)
	}
	a.track(pass)
	res, err := a.client.Login(ctx, user, pass)
	if err != nil {
		// Bad login credentials only (not ACL 403 on business APIs): clear remembered password.
		low := strings.ToLower(err.Error())
		if strings.Contains(low, "http 401") ||
			strings.Contains(low, "user not found") || strings.Contains(low, "password incorrect") ||
			strings.Contains(low, "unknown user") || strings.Contains(low, "bad password") ||
			(strings.Contains(low, "http 403") && (strings.Contains(low, "password") || strings.Contains(low, "user"))) {
			a.disableRememberedAutoLogin()
			return fmt.Errorf("auto-login failed: invalid credentials (remembered login.cred cleared): %w; run: nacos auth login --remember", err)
		}
		return fmt.Errorf("auto-login failed: %w; run: nacos auth login --remember", err)
	}
	cred := credstore.Cred{
		Server: a.settings.Server, Username: user, AuthMode: res.AuthMode, Via: "auto-login",
	}
	if res.AuthMode == "token" {
		cred.AccessToken = res.AccessTok
		a.track(res.AccessTok)
		a.client.ApplyToken(res.AccessTok, "token")
	} else {
		a.client.ApplyToken("", res.AuthMode)
	}
	if err := credstore.Save(a.credPath, cred); err != nil {
		return fmt.Errorf("auto-login token save: %w", err)
	}
	return nil
}

func (a *App) disableRememberedAutoLogin() {
	a.autoLogin = false
	if a.client != nil {
		a.client.RenewOnce = nil
	}
	_ = credstore.ClearLogin(a.loginDir())
}

func (a *App) saveRemembered(user, pass string, remember bool) error {
	dir := a.loginDir()
	if !remember {
		// Align plumelog/xxl: unchecking remember clears prior login.cred.
		return credstore.ClearLogin(dir)
	}
	return credstore.SaveLogin(dir, credstore.LoginPassword{
		Username:  user,
		Password:  pass,
		AutoLogin: true,
	})
}

func (a *App) rememberedStatus() map[string]any {
	dir := a.loginDir()
	u, auto, ok := credstore.PeekLogin(dir)
	stored := ok || credstore.HasLogin(dir)
	envPair := strings.TrimSpace(os.Getenv("NACOS_USERNAME")) != "" && os.Getenv("NACOS_PASSWORD") != ""
	return map[string]any{
		"password_stored": stored,
		"auto_login_pref": auto,
		"username":        u,
		"auto_renew":      a.autoLogin,
		"can_auto_login":  a.autoLogin && (stored || envPair),
		"login_cred_path": credstore.LoginPath(dir),
	}
}
