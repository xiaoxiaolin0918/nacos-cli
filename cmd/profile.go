package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/xiaoxiaolin0918/nacos-cli/internal/config"
	"github.com/xiaoxiaolin0918/nacos-cli/internal/risk"
)

func (a *App) newProfileCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "profile",
		Short: "Manage ~/.nacos-cli/<name>.conf connection profiles",
		Long: `Named profiles (official nacos-cli layout under ~/.nacos-cli/).

Priority (high to low): flags > --config file > --profile > env/.env > defaults.
Passwords in profile files are local-only; never commit them. profile show masks password.`,
		RunE: helpOnly,
	}
	cmd.AddCommand(
		a.newProfileListCmd(),
		a.newProfileShowCmd(),
		a.newProfilePathCmd(),
		a.newProfileSetCmd(),
		a.newProfileDeleteCmd(),
	)
	return cmd
}

func (a *App) newProfileListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "list",
		RunE: func(cmd *cobra.Command, _ []string) error {
			dir, err := config.ProfileDir()
			if err != nil {
				return a.fail(cmd, risk.Read, err)
			}
			names, err := config.ListProfiles()
			if err != nil {
				return a.fail(cmd, risk.Read, err)
			}
			return a.ok(cmd, map[string]any{
				"profile_dir": dir,
				"profiles":    names,
				"active":      a.profileName,
			})
		},
	}
	mark(cmd, risk.Read, "List profile names under ~/.nacos-cli")
	return cmd
}

func (a *App) newProfileShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "show [name]",
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := a.profileName
			if len(args) == 1 {
				name = args[0]
			}
			if name == "" {
				name = "default"
			}
			path, err := config.ProfilePath(name)
			if err != nil {
				return a.fail(cmd, risk.Read, err)
			}
			p, ok, err := config.LoadProfile(name)
			if err != nil {
				return a.fail(cmd, risk.Read, err)
			}
			if !ok {
				return a.fail(cmd, risk.Read, fmt.Errorf("profile %q not found at %s", name, path))
			}
			return a.ok(cmd, map[string]any{
				"name":         name,
				"config_file":  path,
				"profile":      p.RedactedView(),
				"password_set": p.Password != "",
			})
		},
	}
	mark(cmd, risk.Read, "Show one profile with password masked")
	return cmd
}

func (a *App) newProfilePathCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "path [name]",
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := a.profileName
			if len(args) == 1 {
				name = args[0]
			}
			if name == "" {
				name = "default"
			}
			dir, err := config.ProfileDir()
			if err != nil {
				return a.fail(cmd, risk.Read, err)
			}
			path, err := config.ProfilePath(name)
			if err != nil {
				return a.fail(cmd, risk.Read, err)
			}
			return a.ok(cmd, map[string]any{
				"profile_dir": dir,
				"name":        name,
				"path":        path,
			})
		},
	}
	mark(cmd, risk.Read, "Print profile file path")
	return cmd
}

func (a *App) newProfileSetCmd() *cobra.Command {
	var (
		password, contextPath, defaultFormat string
		host                                 string
		port                                 int
	)
	cmd := &cobra.Command{
		Use:  "set <name>",
		Args: cobra.ExactArgs(1),
		Long: "Create or update a profile YAML under ~/.nacos-cli/<name>.conf (mode 0600).\n" +
			"Uses root flags --server / --username / --namespace when provided.\n" +
			"Prefer NACOS_PASSWORD env for secrets; --password is optional and local-only.",
		Example: `  nacos profile set lab --server http://127.0.0.1:8848 --username nacos
  nacos profile set lab --host 127.0.0.1 --port 8848 --username nacos --default-format json
  nacos --profile lab namespace list`,
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			existing, ok, err := config.LoadProfile(name)
			if err != nil {
				return a.fail(cmd, risk.Write, err)
			}
			p := existing
			if !ok {
				p = config.Profile{}
			}
			if changed(cmd, "server") && strings.TrimSpace(a.flagServer) != "" {
				p.Server = strings.TrimRight(strings.TrimSpace(a.flagServer), "/")
				p.Host = ""
				p.Port = 0
			}
			if cmd.Flags().Changed("host") {
				p.Host = strings.TrimSpace(host)
				if !changed(cmd, "server") {
					p.Server = ""
				}
			}
			if cmd.Flags().Changed("port") {
				p.Port = port
			}
			if changed(cmd, "username") && strings.TrimSpace(a.flagUsername) != "" {
				p.Username = strings.TrimSpace(a.flagUsername)
			}
			if cmd.Flags().Changed("password") {
				p.Password = password
			}
			if changed(cmd, "namespace") {
				p.Namespace = config.NormalizeNamespaceID(a.flagNS)
			}
			if cmd.Flags().Changed("context-path") {
				if err := config.ValidateContextPath(contextPath); err != nil {
					return a.fail(cmd, risk.Write, err)
				}
				p.ContextPath = strings.TrimSpace(contextPath)
			}
			if cmd.Flags().Changed("default-format") {
				norm, err := config.NormalizeFormat(defaultFormat)
				if err != nil {
					return a.fail(cmd, risk.Write, err)
				}
				p.Format = norm
			}
			if p.AuthType == "" {
				p.AuthType = "nacos"
			}
			if strings.TrimSpace(p.Server) == "" && strings.TrimSpace(p.Host) == "" {
				return a.fail(cmd, risk.Write, fmt.Errorf("set --server or --host"))
			}
			path, err := config.SaveProfile(name, p)
			if err != nil {
				return a.fail(cmd, risk.Write, err)
			}
			return a.ok(cmd, map[string]any{
				"name":         name,
				"path":         path,
				"password_set": p.Password != "",
				"profile":      p.RedactedView(),
			})
		},
	}
	cmd.Flags().StringVar(&host, "host", "", "host (official-style; composed with --port into server)")
	cmd.Flags().IntVar(&port, "port", 8848, "port used with --host")
	cmd.Flags().StringVar(&password, "password", "", "optional local password (prefer env)")
	cmd.Flags().StringVar(&contextPath, "context-path", "", "API context path (default /nacos). Under Git Bash use MSYS_NO_PATHCONV=1 so /nacos is not rewritten")
	cmd.Flags().StringVar(&defaultFormat, "default-format", "", "default output format stored in profile: json|pretty")
	mark(cmd, risk.Write, "Create or update a local profile file")
	return cmd
}

func (a *App) newProfileDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "delete <name>",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			path, err := config.ProfilePath(name)
			if err != nil {
				return a.fail(cmd, risk.High, err)
			}
			would := map[string]any{"action": "profile.delete", "name": name, "path": path}
			if stop, err := a.gateHigh(cmd, "profile.delete", would); stop {
				return err
			}
			if err := config.DeleteProfile(name); err != nil {
				return a.fail(cmd, risk.High, err)
			}
			return a.ok(cmd, map[string]any{"deleted": name, "path": path})
		},
	}
	mark(cmd, risk.High, "Delete a local profile file")
	return cmd
}

func isProfileCommand(cmd *cobra.Command) bool {
	for c := cmd; c != nil; c = c.Parent() {
		if c.Name() == "profile" {
			return true
		}
	}
	return false
}