package cmd

import (
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/xiaoxiaolin0918/nacos-cli/internal/output"
	"github.com/xiaoxiaolin0918/nacos-cli/internal/risk"
	"github.com/xiaoxiaolin0918/nacos-cli/internal/version"
)

type commandInfo struct {
	Path           string   `json:"path"`
	Use            string   `json:"use"`
	Short          string   `json:"short,omitempty"`
	Risk           string   `json:"risk"`
	NeedsYes       bool     `json:"needs_yes"`
	SupportsDryRun bool     `json:"supports_dry_run"`
	Flags          []string `json:"flags,omitempty"`
}

func (a *App) newCommandsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "commands",
		Short: "List commands with risk / gate metadata (Agent capability table)",
		Long: `Emit a machine-readable capability table for every command.

Equivalent to root flag --help-json. Safe for Agents: no server call.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if a.format == "" {
				a.format = "json"
			}
			return a.ok(cmd, a.commandsPayload(cmd.Root()))
		},
	}
	mark(cmd, risk.Read, "List command capabilities for Agents")
	return cmd
}

func (a *App) commandsPayload(root *cobra.Command) map[string]any {
	list := []commandInfo{}
	var walk func(c *cobra.Command, prefix string)
	walk = func(c *cobra.Command, prefix string) {
		if c == nil || !c.IsAvailableCommand() || c.Hidden {
			return
		}
		path := strings.TrimSpace(prefix + " " + c.Name())
		if c == root {
			path = c.Name()
		}
		if c != root {
			r := cmdRisk(c)
			if r == "" {
				r = risk.Read
			}
			info := commandInfo{
				Path:           strings.TrimSpace(path),
				Use:            c.Use,
				Short:          c.Short,
				Risk:           r,
				NeedsYes:       r == risk.High,
				SupportsDryRun: r == risk.High || r == risk.Write,
				Flags:          localFlagNames(c),
			}
			list = append(list, info)
		}
		for _, ch := range c.Commands() {
			walk(ch, path)
		}
	}
	walk(root, "")
	return map[string]any{
		"name":     version.Name,
		"version":  version.Version,
		"envelope": "ok-json",
		"commands": list,
		"global_flags": []string{
			"--server", "--username", "--namespace", "--profile", "--config",
			"--format", "--output", "--dry-run", "--yes",
			"--auto-login", "--auto-renew",
			"--check-update", "--auto-update", "--help-json",
		},
		"error_codes": []string{
			output.CodeConfirmationRequired,
			output.CodeServerUnset,
			output.CodeCLIError,
			output.CodeDoctorFailed,
			output.CodeUpdateFailed,
			"http_401", "http_403", "http_N",
		},
		"update": map[string]any{
			"check_update": map[string]any{
				"default": false,
				"flags":   []string{"--check-update"},
				"env":     "NACOS_CHECK_UPDATE",
				"profile": "check_update",
				"hint":    "opportunistic stderr notice when newer release exists (24h cache)",
			},
			"auto_update": map[string]any{
				"default": false,
				"flags":   []string{"--auto-update"},
				"env":     "NACOS_AUTO_UPDATE",
				"profile": "auto_update",
				"hint":    "when check finds newer + this on: silent download/replace (rollback on failure)",
			},
			"command": "nacos update [--check|--yes]",
		},
	}
}

func localFlagNames(c *cobra.Command) []string {
	out := []string{}
	if c.Flags() == nil {
		return out
	}
	c.Flags().VisitAll(func(f *pflag.Flag) {
		if f.Name == "help" {
			return
		}
		out = append(out, "--"+f.Name)
	})
	return out
}

func isCommandsCommand(cmd *cobra.Command) bool {
	for c := cmd; c != nil; c = c.Parent() {
		if c.Name() == "commands" {
			return true
		}
	}
	return false
}

func (a *App) emitHelpJSON(cmd *cobra.Command) error {
	if a.format == "" {
		a.format = "json"
	}
	root := cmd.Root()
	if root == nil {
		root = cmd
	}
	env := output.Envelope{OK: true, Result: output.MustResult(a.commandsPayload(root))}
	if err := a.write(cmd, env); err != nil {
		return err
	}
	return SilentSuccess{}
}
