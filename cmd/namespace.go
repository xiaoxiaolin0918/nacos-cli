package cmd

import (
	"github.com/spf13/cobra"

	"github.com/xiaoxiaolin0918/nacos-cli/internal/domain"
	"github.com/xiaoxiaolin0918/nacos-cli/internal/risk"
)

func (a *App) newNamespaceCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "namespace", Short: "Namespace: list", RunE: helpOnly}
	cmd.AddCommand(a.newNamespaceListCmd())
	return cmd
}

func (a *App) newNamespaceListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "list",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := a.requireServer(cmd); err != nil {
				return err
			}
			uc := domain.NamespaceUseCase{Port: a.client}
			out, err := uc.List()
			if err != nil {
				return a.fail(cmd, risk.Read, err)
			}
			return a.ok(cmd, out)
		},
	}
	mark(cmd, risk.Read, "List namespaces")
	return cmd
}
