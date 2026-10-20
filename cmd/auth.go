package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/term"

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
	var web, browser bool
	cmd := &cobra.Command{
		Use: "login",
		Long: `POST /v1/auth/login. On AUTH_DISABLED servers, stores auth_mode=disabled without a token.
Password from --password, NACOS_PASSWORD, or a terminal prompt (never printed).
--web starts a 127.0.0.1-only form (120s timeout).`,
		Example: `  nacos auth login --username nacos
  nacos auth login --web
  nacos auth login --dry-run`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if a.dryRun {
				return a.write(cmd, output.Envelope{
					OK: true, DryRun: true, Risk: risk.Write,
					Would: map[string]any{"action": "auth.login", "web": web || browser, "server": a.settings.Server},
				})
			}
			if err := a.requireServer(cmd); err != nil {
				return err
			}
			if web || browser {
				return a.runWebLogin(cmd)
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
			if pass == "" {
				p, err := promptPassword(cmd)
				if err != nil {
					// allow empty password for AUTH_DISABLED
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
			st := credstore.Status(a.credPath)
			st["ok"] = true
			return a.ok(cmd, st)
		},
	}
	cmd.Flags().StringVar(&username, "username", "", "login username")
	cmd.Flags().StringVar(&password, "password", "", "login password (never printed)")
	cmd.Flags().BoolVar(&web, "web", false, "open a local browser form on 127.0.0.1")
	cmd.Flags().BoolVar(&browser, "browser", false, "alias of --web")
	mark(cmd, risk.Write, "Log in and persist access token (or AUTH_DISABLED)")
	return cmd
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
	cmd := &cobra.Command{
		Use: "logout",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cleared := true
			if err := credstore.Clear(a.credPath); err != nil {
				cleared = false
			}
			a.client.ApplyToken("", "none")
			return a.ok(cmd, map[string]any{
				"credentials_path": a.credPath,
				"cleared":          cleared,
			})
		},
	}
	mark(cmd, risk.Write, "Clear local credentials file")
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
