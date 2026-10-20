package cmd

import (
	"github.com/spf13/cobra"

	"github.com/xiaoxiaolin0918/nacos-cli/internal/risk"
)

func (a *App) newWhoamiCmd() *cobra.Command {
	var probe bool
	cmd := &cobra.Command{
		Use:   "whoami",
		Short: "Alias of auth status",
		Long:  "Thin alias of auth status (same fields). Cookie/token values never printed.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return a.runAuthStatus(cmd, probe)
		},
	}
	cmd.Flags().BoolVar(&probe, "probe", false, "GET namespaces to verify session")
	mark(cmd, risk.Read, "Show session identity metadata without secrets")
	return cmd
}
