package cmd

import (
	"github.com/spf13/cobra"

	"github.com/xiaoxiaolin0918/nacos-cli/internal/domain"
	"github.com/xiaoxiaolin0918/nacos-cli/internal/risk"
)

func (a *App) newRoleCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "role", Short: "ACL roles: list/bind/unbind", RunE: helpOnly}
	cmd.AddCommand(
		a.newRoleListCmd(),
		a.newRoleBindCmd(),
		a.newRoleUnbindCmd(),
	)
	return cmd
}

func (a *App) newRoleListCmd() *cobra.Command {
	var username, role string
	var page, pageSize int
	cmd := &cobra.Command{
		Use: "list",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := a.requireServer(cmd); err != nil {
				return err
			}
			uc := domain.RoleUseCase{Port: a.client}
			out, err := uc.List(username, role, page, pageSize)
			if err != nil {
				return a.fail(cmd, risk.Read, err)
			}
			return a.ok(cmd, out)
		},
	}
	cmd.Flags().StringVar(&username, "username", "", "blur username filter")
	cmd.Flags().StringVar(&role, "role", "", "role filter (optional)")
	cmd.Flags().IntVar(&page, "page", 1, "pageNo")
	cmd.Flags().IntVar(&pageSize, "page-size", 100, "pageSize")
	mark(cmd, risk.Read, "List roles / user-role bindings")
	return cmd
}

func (a *App) newRoleBindCmd() *cobra.Command {
	var role, username string
	cmd := &cobra.Command{
		Use: "bind",
		RunE: func(cmd *cobra.Command, _ []string) error {
			would := map[string]any{"action": "role.bind", "role": role, "username": username}
			stop, err := a.gateHigh(cmd, "role.bind", would)
			if stop || err != nil {
				return err
			}
			if err := a.requireServer(cmd); err != nil {
				return err
			}
			uc := domain.RoleUseCase{Port: a.client}
			out, err := uc.Bind(role, username)
			if err != nil {
				return a.fail(cmd, risk.High, err)
			}
			return a.ok(cmd, out)
		},
	}
	cmd.Flags().StringVar(&role, "role", "", "role name (required)")
	cmd.Flags().StringVar(&username, "username", "", "username (required)")
	_ = cmd.MarkFlagRequired("role")
	_ = cmd.MarkFlagRequired("username")
	mark(cmd, risk.High, "Bind role to user (creates role if needed)")
	return cmd
}

func (a *App) newRoleUnbindCmd() *cobra.Command {
	var role, username string
	cmd := &cobra.Command{
		Use: "unbind",
		RunE: func(cmd *cobra.Command, _ []string) error {
			would := map[string]any{"action": "role.unbind", "role": role, "username": username}
			stop, err := a.gateHigh(cmd, "role.unbind", would)
			if stop || err != nil {
				return err
			}
			if err := a.requireServer(cmd); err != nil {
				return err
			}
			uc := domain.RoleUseCase{Port: a.client}
			out, err := uc.Unbind(role, username)
			if err != nil {
				return a.fail(cmd, risk.High, err)
			}
			return a.ok(cmd, out)
		},
	}
	cmd.Flags().StringVar(&role, "role", "", "role name (required)")
	cmd.Flags().StringVar(&username, "username", "", "username (required)")
	_ = cmd.MarkFlagRequired("role")
	_ = cmd.MarkFlagRequired("username")
	mark(cmd, risk.High, "Unbind role from user")
	return cmd
}
