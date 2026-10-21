package cmd

import (
	"github.com/spf13/cobra"

	"github.com/xiaoxiaolin0918/nacos-cli/internal/domain"
	"github.com/xiaoxiaolin0918/nacos-cli/internal/risk"
)

func (a *App) newUserCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "user", Short: "ACL users: list/create/delete/passwd", RunE: helpOnly}
	cmd.AddCommand(
		a.newUserListCmd(),
		a.newUserCreateCmd(),
		a.newUserDeleteCmd(),
		a.newUserPasswdCmd(),
	)
	return cmd
}

func (a *App) newUserListCmd() *cobra.Command {
	var username string
	var page, pageSize int
	cmd := &cobra.Command{
		Use: "list",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := a.requireServer(cmd); err != nil {
				return err
			}
			uc := domain.UserUseCase{Port: a.client}
			out, err := uc.List(username, page, pageSize)
			if err != nil {
				return a.fail(cmd, risk.Read, err)
			}
			return a.ok(cmd, domain.RedactUserPasswords(out))
		},
	}
	cmd.Flags().StringVar(&username, "username", "", "blur username filter")
	cmd.Flags().IntVar(&page, "page", 1, "pageNo")
	cmd.Flags().IntVar(&pageSize, "page-size", 100, "pageSize")
	mark(cmd, risk.Read, "List users")
	return cmd
}

func (a *App) newUserCreateCmd() *cobra.Command {
	var username, password string
	cmd := &cobra.Command{
		Use:  "create",
		Long: "Create a Nacos local user. Password is never echoed in dry-run output.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			would := map[string]any{
				"action":          "user.create",
				"username":        username,
				"password_set":    password != "",
				"password_length": len([]byte(password)),
			}
			stop, err := a.gateHigh(cmd, "user.create", would)
			if stop || err != nil {
				return err
			}
			if err := a.requireServer(cmd); err != nil {
				return err
			}
			uc := domain.UserUseCase{Port: a.client}
			out, err := uc.Create(username, password)
			if err != nil {
				return a.fail(cmd, risk.High, err)
			}
			return a.ok(cmd, out)
		},
	}
	cmd.Flags().StringVar(&username, "username", "", "username (required)")
	cmd.Flags().StringVar(&password, "password", "", "password (required; never logged)")
	_ = cmd.MarkFlagRequired("username")
	_ = cmd.MarkFlagRequired("password")
	mark(cmd, risk.High, "Create user")
	return cmd
}

func (a *App) newUserDeleteCmd() *cobra.Command {
	var username string
	cmd := &cobra.Command{
		Use: "delete",
		RunE: func(cmd *cobra.Command, _ []string) error {
			would := map[string]any{"action": "user.delete", "username": username}
			stop, err := a.gateHigh(cmd, "user.delete", would)
			if stop || err != nil {
				return err
			}
			if err := a.requireServer(cmd); err != nil {
				return err
			}
			uc := domain.UserUseCase{Port: a.client}
			out, err := uc.Delete(username)
			if err != nil {
				return a.fail(cmd, risk.High, err)
			}
			return a.ok(cmd, out)
		},
	}
	cmd.Flags().StringVar(&username, "username", "", "username (required)")
	_ = cmd.MarkFlagRequired("username")
	mark(cmd, risk.High, "Delete user")
	return cmd
}

func (a *App) newUserPasswdCmd() *cobra.Command {
	var username, password string
	cmd := &cobra.Command{
		Use:  "passwd",
		Long: "Update user password. New password is never echoed in dry-run output.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			would := map[string]any{
				"action":          "user.passwd",
				"username":        username,
				"password_set":    password != "",
				"password_length": len([]byte(password)),
			}
			stop, err := a.gateHigh(cmd, "user.passwd", would)
			if stop || err != nil {
				return err
			}
			if err := a.requireServer(cmd); err != nil {
				return err
			}
			uc := domain.UserUseCase{Port: a.client}
			out, err := uc.UpdatePassword(username, password)
			if err != nil {
				return a.fail(cmd, risk.High, err)
			}
			return a.ok(cmd, out)
		},
	}
	cmd.Flags().StringVar(&username, "username", "", "username (required)")
	cmd.Flags().StringVar(&password, "password", "", "new password (required; never logged)")
	_ = cmd.MarkFlagRequired("username")
	_ = cmd.MarkFlagRequired("password")
	mark(cmd, risk.High, "Update user password")
	return cmd
}
