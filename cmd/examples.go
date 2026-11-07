package cmd

import (
	"strings"

	"github.com/spf13/cobra"

	"github.com/xiaoxiaolin0918/nacos-cli/internal/output"
	"github.com/xiaoxiaolin0918/nacos-cli/internal/risk"
	"github.com/xiaoxiaolin0918/nacos-cli/internal/version"
)

type exampleSample struct {
	Command string         `json:"command"`
	Title   string         `json:"title"`
	OK      map[string]any `json:"ok_example"`
	Fail    map[string]any `json:"fail_example"`
}

func (a *App) newExamplesCmd() *cobra.Command {
	var command string
	cmd := &cobra.Command{
		Use:   "examples",
		Short: "Emit success/failure JSON envelope samples for Agents",
		Long:  "Print baked-in ok-json success and failure samples.\n\nUse --command to filter (substring match on sample command path).\nSafe: no server call. Intended to feed Skills / Agents.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if a.format == "" {
				a.format = "json"
			}
			samples := exampleCatalog()
			q := strings.TrimSpace(strings.ToLower(command))
			if q != "" {
				filtered := make([]exampleSample, 0)
				for _, s := range samples {
					if strings.Contains(strings.ToLower(s.Command), q) {
						filtered = append(filtered, s)
					}
				}
				samples = filtered
			}
			return a.ok(cmd, map[string]any{
				"name":                version.Name,
				"version":             version.Version,
				"capabilities_schema": output.CapabilitiesSchemaVersion,
				"envelope":            "ok-json",
				"would":               "array",
				"redact_fields":       output.DefaultRedactFields,
				"samples":             samples,
			})
		},
	}
	cmd.Flags().StringVar(&command, "command", "", "filter samples by command path substring")
	mark(cmd, risk.Read, "JSON envelope examples for Agents")
	return cmd
}

func isExamplesCommand(cmd *cobra.Command) bool {
	for c := cmd; c != nil; c = c.Parent() {
		if c.Name() == "examples" {
			return true
		}
	}
	return false
}

func exampleCatalog() []exampleSample {
	return []exampleSample{
		{
			Command: "config publish --dry-run",
			Title:   "high-risk write dry-run with would[]",
			OK: map[string]any{
				"ok": true, "dry_run": true, "risk": "high-risk-write",
				"would": []map[string]any{{
					"action": "config.publish", "data_id": "app.yaml", "group": "DEFAULT_GROUP",
				}},
				"would_require": []string{"--yes"},
			},
			Fail: map[string]any{
				"ok": false, "risk": "high-risk-write",
				"error": map[string]any{
					"type": "confirmation", "subtype": "confirmation_required",
					"code": output.CodeConfirmationRequired,
					"message": "high-risk write requires --yes",
				},
			},
		},
		{
			Command: "doctor",
			Title:   "doctor success envelope",
			OK: map[string]any{
				"ok": true,
				"result": map[string]any{"healthy": true, "can_auto_login": false},
			},
			Fail: map[string]any{
				"ok": false,
				"error": map[string]any{
					"type": "cli", "code": output.CodeServerUnset,
					"message": "NACOS_SERVER not set",
				},
			},
		},
		{
			Command: "auth login",
			Title:   "non-TTY missing password",
			OK: map[string]any{
				"ok": true,
				"result": map[string]any{"auth_mode": "token", "username": "nacos"},
			},
			Fail: map[string]any{
				"ok": false, "risk": "write",
				"error": map[string]any{
					"type": "cli", "code": output.CodePasswordRequired,
					"message": "password required: pass --password or set NACOS_PASSWORD",
				},
			},
		},
		{
			Command: "config list --limit/--query",
			Title:   "list with Agent-friendly limit/query",
			OK: map[string]any{
				"ok": true,
				"result": map[string]any{
					"pageItems": []any{}, "namespace": "", "effective_namespace": "",
				},
			},
			Fail: map[string]any{
				"ok": false,
				"error": map[string]any{"type": "http", "code": "http_401", "message": "HTTP 401: token expired"},
			},
		},
		{
			Command: "commands / --trace",
			Title:   "capability table + optional meta",
			OK: map[string]any{
				"ok": true,
				"meta": map[string]any{"trace": true, "duration_ms": 12, "renewed": false},
				"result": map[string]any{
					"envelope":            "ok-json",
					"capabilities_schema": output.CapabilitiesSchemaVersion,
				},
			},
			Fail: map[string]any{
				"ok": false,
				"error": map[string]any{"type": "cli", "code": output.CodeCLIError, "message": "cli_error"},
			},
		},
	}
}
