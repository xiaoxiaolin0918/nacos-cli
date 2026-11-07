package cmd

import (
	"github.com/spf13/cobra"

	"github.com/xiaoxiaolin0918/nacos-cli/internal/credstore"
	"github.com/xiaoxiaolin0918/nacos-cli/internal/output"
	"github.com/xiaoxiaolin0918/nacos-cli/internal/risk"
	"github.com/xiaoxiaolin0918/nacos-cli/internal/version"
)

func (a *App) newDoctorCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "doctor",
		Long: `Check NACOS_SERVER, credentials file, reachability, and a namespace list probe.
Never prints passwords or raw tokens.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			type check struct {
				Name   string `json:"name"`
				OK     bool   `json:"ok"`
				Detail string `json:"detail,omitempty"`
			}
			checks := []check{}
			healthy := true
			rem := a.rememberedStatus()
			report := map[string]any{
				"version":          version.Version,
				"server":           a.settings.Server,
				"context_path":     a.settings.ContextPath,
				"username_env":     a.settings.Username,
				"has_password_env": a.settings.Password != "",
				"credentials":      credstore.Status(a.credPath),
				"remembered":       rem,
				"auto_renew":       a.autoLogin,
				"can_auto_login":   rem["can_auto_login"],
				"check_update":     a.checkUpdateEnabled,
				"auto_update":      a.autoUpdateEnabled,
				"base_url":         a.settings.BaseURL(),
			}
			if a.settings.Server == "" {
				healthy = false
				checks = append(checks, check{Name: "server", OK: false, Detail: "NACOS_SERVER not set"})
				report["checks"] = checks
				report["ok"] = false
				env := output.Envelope{OK: false, Result: output.MustResult(report), Error: output.ErrorBody{Type: "cli", Code: output.CodeServerUnset, Message: "NACOS_SERVER not set"}}
				if err := a.write(cmd, env); err != nil {
					return err
				}
				return &PrintedError{Msg: "NACOS_SERVER not set", Code: 2}
			}
			checks = append(checks, check{Name: "server", OK: true, Detail: a.settings.Server})

			status, err := a.client.Ping(cmd.Context())
			if err != nil {
				healthy = false
				checks = append(checks, check{Name: "reachable", OK: false, Detail: "connection failed"})
			} else {
				checks = append(checks, check{Name: "reachable", OK: true, Detail: "HTTP " + itoa(status)})
			}

			if healthy {
				probe := a.client.Probe(cmd.Context())
				ok, _ := probe["ok"].(bool)
				detail := "ok"
				if !ok {
					healthy = false
					if e, has := probe["error"].(string); has {
						detail = e
					} else {
						detail = "probe failed"
					}
				} else if mode, _ := probe["auth_mode"].(string); mode != "" {
					detail = "auth_mode=" + mode
				}
				checks = append(checks, check{Name: "namespaces", OK: ok, Detail: detail})
				report["probe"] = probe
				state := a.client.ServerState(cmd.Context())
				report["server_state"] = state
				if ver, _ := state["version"].(string); ver != "" {
					checks = append(checks, check{Name: "server_version", OK: true, Detail: ver})
				} else if sok, _ := state["ok"].(bool); sok {
					checks = append(checks, check{Name: "server_version", OK: true, Detail: "state ok (no version field)"})
				} else {
					// Best-effort: do not fail doctor solely on missing console state.
					det := "unavailable"
					if e, has := state["error"].(string); has && e != "" {
						det = e
					}
					checks = append(checks, check{Name: "server_version", OK: true, Detail: det})
				}
			}

			report["checks"] = checks
			report["ok"] = healthy
			env := output.Envelope{OK: healthy, Result: output.MustResult(report)}
			if !healthy {
				env.Error = output.ErrorBody{Type: "cli", Code: output.CodeDoctorFailed, Message: "doctor found a failing check"}
				if err := a.write(cmd, env); err != nil {
					return err
				}
				return &PrintedError{Msg: "doctor found a failing check"}
			}
			return a.write(cmd, env)
		},
	}
	var checkUpdate bool
	cmd.Flags().BoolVar(&checkUpdate, "check-update", false, "also query GitHub Releases for a newer binary (best-effort)")
	mark(cmd, risk.Read, "Check connectivity and credentials")
	inner := cmd.RunE
	cmd.RunE = func(c *cobra.Command, args []string) error {
		err := inner(c, args)
		if !checkUpdate && !a.checkUpdateEnabled {
			return err
		}
		res, uerr := a.checkUpdate(c)
		if uerr != nil {
			c.PrintErrf("nacos: update check failed: %v\n", uerr)
			return err
		}
		if res != nil && res.Status == "outdated" {
			c.PrintErrf("nacos: newer version %s available (current %s); run: nacos update --yes\n", res.Latest, res.Current)
		}
		return err
	}
	return cmd
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [16]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
