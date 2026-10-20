package cmd

import (
	"github.com/spf13/cobra"

	"github.com/xiaoxiaolin0918/nacos-cli/internal/domain"
	"github.com/xiaoxiaolin0918/nacos-cli/internal/risk"
)

func (a *App) newPermissionCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "permission", Short: "ACL permissions: list/grant/revoke", RunE: helpOnly}
	cmd.AddCommand(
		a.newPermissionListCmd(),
		a.newPermissionGrantCmd(),
		a.newPermissionRevokeCmd(),
	)
	return cmd
}

func (a *App) newPermissionListCmd() *cobra.Command {
	var role string
	var page, pageSize int
	cmd := &cobra.Command{
		Use: "list",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := a.requireServer(cmd); err != nil {
				return err
			}
			uc := domain.PermissionUseCase{Port: a.client}
			out, err := uc.List(role, page, pageSize)
			if err != nil {
				return a.fail(cmd, risk.Read, err)
			}
			return a.ok(cmd, out)
		},
	}
	cmd.Flags().StringVar(&role, "role", "", "role name filter")
	cmd.Flags().IntVar(&page, "page", 1, "pageNo")
	cmd.Flags().IntVar(&pageSize, "page-size", 100, "pageSize")
	mark(cmd, risk.Read, "List permissions for a role")
	return cmd
}

func (a *App) newPermissionGrantCmd() *cobra.Command {
	var role, resource, action string
	cmd := &cobra.Command{
		Use:  "grant",
		Long: "Grant permission to a role. Resource is usually namespaceId:group:dataId or *:*:* . Action: r|w|rw.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			would := map[string]any{
				"action": "permission.grant", "role": role, "resource": resource, "perm_action": action,
			}
			stop, err := a.gateHigh(cmd, "permission.grant", would)
			if stop || err != nil {
				return err
			}
			if err := a.requireServer(cmd); err != nil {
				return err
			}
			uc := domain.PermissionUseCase{Port: a.client}
			out, err := uc.Grant(role, resource, action)
			if err != nil {
				return a.fail(cmd, risk.High, err)
			}
			return a.ok(cmd, out)
		},
	}
	cmd.Flags().StringVar(&role, "role", "", "role (required)")
	cmd.Flags().StringVar(&resource, "resource", "", "resource pattern (required)")
	cmd.Flags().StringVar(&action, "action", "", "r|w|rw (required)")
	_ = cmd.MarkFlagRequired("role")
	_ = cmd.MarkFlagRequired("resource")
	_ = cmd.MarkFlagRequired("action")
	mark(cmd, risk.High, "Grant permission to role")
	return cmd
}

func (a *App) newPermissionRevokeCmd() *cobra.Command {
	var role, resource, action string
	cmd := &cobra.Command{
		Use: "revoke",
		RunE: func(cmd *cobra.Command, _ []string) error {
			would := map[string]any{
				"action": "permission.revoke", "role": role, "resource": resource, "perm_action": action,
			}
			stop, err := a.gateHigh(cmd, "permission.revoke", would)
			if stop || err != nil {
				return err
			}
			if err := a.requireServer(cmd); err != nil {
				return err
			}
			uc := domain.PermissionUseCase{Port: a.client}
			out, err := uc.Revoke(role, resource, action)
			if err != nil {
				return a.fail(cmd, risk.High, err)
			}
			return a.ok(cmd, out)
		},
	}
	cmd.Flags().StringVar(&role, "role", "", "role (required)")
	cmd.Flags().StringVar(&resource, "resource", "", "resource pattern (required)")
	cmd.Flags().StringVar(&action, "action", "", "r|w|rw (required)")
	_ = cmd.MarkFlagRequired("role")
	_ = cmd.MarkFlagRequired("resource")
	_ = cmd.MarkFlagRequired("action")
	mark(cmd, risk.High, "Revoke permission from role")
	return cmd
}
