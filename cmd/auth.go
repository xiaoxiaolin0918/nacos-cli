package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/xiaoxiaolin0918/nacos-cli/internal/config"
	"github.com/xiaoxiaolin0918/nacos-cli/internal/credstore"
	"github.com/xiaoxiaolin0918/nacos-cli/internal/output"
	"github.com/xiaoxiaolin0918/nacos-cli/internal/risk"
)

func (a *App) newAuthCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "auth", Short: "Login, logout, and local token status", RunE: helpOnly}
	cmd.AddCommand(a.newLoginCmd(), a.newLogoutCmd(), a.newStatusCmd())
	return cmd
}

func (a *App) newLoginCmd() *cobra.Command {
	var username, password string
	var web, browser, remember, storePassword bool
	cmd := &cobra.Command{
		Use: "login",
		Long: `POST /v1/auth/login. On AUTH_DISABLED servers, stores auth_mode=disabled without a token.
Password from --password, NACOS_PASSWORD, or a terminal prompt (never printed).
--web starts a 127.0.0.1-only form (120s timeout).
--remember (alias --store-password) encrypts the password into login.cred beside credentials.json
(Win=DPAPI; macOS/Linux=AES-GCM+login.key) for later --auto-login. Password never enters conf/audit.`,
		Example: `  nacos auth login --username nacos
  nacos auth login --remember
  nacos auth login --web
  nacos auth login --dry-run`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			store := remember || storePassword
			if a.dryRun {
				return a.write(cmd, output.Envelope{
					OK: true, DryRun: true, Risk: risk.Write,
					Would: output.AsWould(map[string]any{"action": "auth.login", "web": web || browser, "server": a.settings.Server, "remember": store}),
				})
			}
			if err := a.requireServer(cmd); err != nil {
				return err
			}
			if web || browser {
				return a.runWebLogin(cmd, store)
			}
			user := username
			if user == "" {
				user = a.settings.Username
			}
			if user == "" {
				return a.fail(cmd, risk.Write, errors.New("username required (--username or NACOS_USERNAME)"))
			}
			pass := password
			if pass == "" {
				pass = a.settings.Password
			}
			if pass == "" && !cmd.Flags().Changed("password") {
				p, err := promptPassword(cmd)
				if err != nil {
					// Non-TTY: never prompt. If username is set, require --password / NACOS_PASSWORD.
					// Empty user+pass may still work for AUTH_DISABLED.
					if !isInteractiveStdin(cmd) && strings.TrimSpace(user) != "" {
						return a.failCoded(cmd, risk.Write, output.CodePasswordRequired, err)
					}
					pass = ""
				} else {
					pass = p
				}
			}
			a.track(pass)
			res, err := a.client.Login(cmd.Context(), user, pass)
			if err != nil {
				return a.fail(cmd, risk.Write, err)
			}
			cred := credstore.Cred{
				Server: a.settings.Server, Username: user, AuthMode: res.AuthMode, Via: "password",
			}
			if res.AuthMode == "token" {
				cred.AccessToken = res.AccessTok
				a.track(res.AccessTok)
			}
			if err := credstore.Save(a.credPath, cred); err != nil {
				return a.fail(cmd, risk.Write, err)
			}
			if err := a.saveRemembered(user, pass, store); err != nil {
				return a.fail(cmd, risk.Write, fmt.Errorf("token saved but remember failed: %w", err))
			}
			st := credstore.Status(a.credPath)
			st["ok"] = true
			st["password_stored"] = store && pass != ""
			st["remembered"] = a.rememberedStatus()
			return a.ok(cmd, st)
		},
	}
	cmd.Flags().StringVar(&username, "username", "", "login username")
	cmd.Flags().StringVar(&password, "password", "", "login password (never printed)")
	cmd.Flags().BoolVar(&web, "web", false, "open a local browser form on 127.0.0.1")
	cmd.Flags().BoolVar(&browser, "browser", false, "alias of --web")
	cmd.Flags().BoolVar(&remember, "remember", false, "encrypt password into login.cred for --auto-login (opt-in)")
	cmd.Flags().BoolVar(&storePassword, "store-password", false, "alias of --remember")
	mark(cmd, risk.Write, "Log in and persist access token (or AUTH_DISABLED)")
	return cmd
}


func isInteractiveStdin(cmd *cobra.Command) bool {
	in := cmd.InOrStdin()
	f, ok := in.(*os.File)
	if !ok {
		return false
	}
	return term.IsTerminal(int(f.Fd()))
}

func promptPassword(cmd *cobra.Command) (string, error) {
	in := cmd.InOrStdin()
	f, ok := in.(*os.File)
	if !ok || !term.IsTerminal(int(f.Fd())) {
		return "", errors.New("password required: pass --password or set NACOS_PASSWORD")
	}
	fmt.Fprint(cmd.ErrOrStderr(), "Nacos password (empty if auth disabled): ")
	b, err := term.ReadPassword(int(f.Fd()))
	fmt.Fprintln(cmd.ErrOrStderr())
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (a *App) newLogoutCmd() *cobra.Command {
	var forget, clearPassword bool
	cmd := &cobra.Command{
		Use: "logout",
		Long: `Clear the local access token (credentials.json). Encrypted login.cred is kept by default
so --auto-login can mint a new token later. Pass --forget (alias --clear-password) to also
remove credentials-dir login.cred/login.key AND the current --profile encrypted store under
~/.nacos-cli/.login/<profile>/ (supervisory-web remembered passwords). Default logout never
double-clears; only --forget does.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			doForget := forget || clearPassword
			would := map[string]any{
				"action":                 "auth.logout",
				"clear_token":            true,
				"clear_login_cred":       doForget,
				"clear_profile_login":    doForget,
				"credentials_path":       a.credPath,
				"login_cred_path":        credstore.LoginPath(a.loginDir()),
				"profile":                a.profileName,
			}
			if a.dryRun {
				return a.write(cmd, output.Envelope{OK: true, DryRun: true, Risk: risk.Write, Would: output.AsWould(would)})
			}
			clearedToken := true
			if err := credstore.Clear(a.credPath); err != nil {
				clearedToken = false
			}
			a.client.ApplyToken("", "none")
			clearedLogin := false
			clearedProfileLogin := false
			profileLoginPath := ""
			if doForget {
				if err := credstore.ClearLogin(a.loginDir()); err != nil {
					return a.fail(cmd, risk.Write, err)
				}
				clearedLogin = true
				if pdir, err := config.ProfileDir(); err == nil && strings.TrimSpace(a.profileName) != "" {
					pldir := credstore.ProfileLoginDir(pdir, a.profileName)
					profileLoginPath = credstore.LoginPath(pldir)
					if err := credstore.ClearLogin(pldir); err != nil {
						return a.fail(cmd, risk.Write, err)
					}
					clearedProfileLogin = true
				}
				a.autoLogin = false
				if a.client != nil {
					a.client.RenewOnce = nil
				}
			}
			return a.ok(cmd, map[string]any{
				"credentials_path":      a.credPath,
				"cleared":               clearedToken,
				"cleared_token":         clearedToken,
				"cleared_login":         clearedLogin,
				"cleared_profile_login": clearedProfileLogin,
				"profile_login_path":    profileLoginPath,
				"forgot_password":       doForget,
			})
		},
	}
	cmd.Flags().BoolVar(&forget, "forget", false, "also delete encrypted login.cred (+login.key) and current profile ~/.nacos-cli/.login/<name>")
	cmd.Flags().BoolVar(&clearPassword, "clear-password", false, "alias of --forget")
	mark(cmd, risk.Write, "Clear local access token; --forget also clears remembered passwords")
	return cmd
}

func (a *App) newStatusCmd() *cobra.Command {
	var probe bool
	cmd := &cobra.Command{
		Use: "status",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return a.runAuthStatus(cmd, probe)
		},
	}
	cmd.Flags().BoolVar(&probe, "probe", false, "GET namespaces to verify session")
	mark(cmd, risk.Read, "Show token presence without printing secrets")
	return cmd
}

func (a *App) runAuthStatus(cmd *cobra.Command, probe bool) error {
	st := credstore.Status(a.credPath)
	st["server_env"] = a.settings.Server
	st["ok"] = true
	st["remembered"] = a.rememberedStatus()
	st["auto_renew"] = a.autoLogin
	if !probe {
		return a.ok(cmd, st)
	}
	if strings.TrimSpace(a.settings.Server) == "" {
		st["probe"] = map[string]any{"ok": false, "error": "NACOS_SERVER not set"}
		st["ok"] = false
		env := output.Envelope{OK: false, Result: output.MustResult(st), Error: output.ErrorBody{Type: "cli", Message: "NACOS_SERVER not set"}}
		if err := a.write(cmd, env); err != nil {
			return err
		}
		return &PrintedError{Msg: "NACOS_SERVER not set"}
	}
	st["probe"] = a.client.Probe(cmd.Context())
	return a.ok(cmd, st)
}
