package cmd

import (
	"os"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/xiaoxiaolin0918/nacos-cli/internal/domain"
	"github.com/xiaoxiaolin0918/nacos-cli/internal/risk"
)

func (a *App) newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "config", Short: "Config center: get/list/publish/history/rollback/export", RunE: helpOnly}
	cmd.AddCommand(
		a.newConfigGetCmd(),
		a.newConfigListCmd(),
		a.newConfigPublishCmd(),
		a.newConfigHistoryCmd(),
		a.newConfigRollbackCmd(),
		a.newConfigExportCmd(),
	)
	return cmd
}

func (a *App) newConfigGetCmd() *cobra.Command {
	var dataID, group, ns string
	cmd := &cobra.Command{
		Use: "get",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := a.requireServer(cmd); err != nil {
				return err
			}
			uc := domain.ConfigUseCase{Port: a.client}
			out, err := uc.Get(dataID, group, ns)
			if err != nil {
				return a.fail(cmd, risk.Read, err)
			}
			return a.ok(cmd, map[string]any{"data_id": dataID, "group": group, "content": out})
		},
	}
	cmd.Flags().StringVar(&dataID, "data-id", "", "dataId (required)")
	cmd.Flags().StringVar(&group, "group", "DEFAULT_GROUP", "group")
	cmd.Flags().StringVar(&ns, "namespace", "", "tenant/namespace id (overrides default)")
	_ = cmd.MarkFlagRequired("data-id")
	mark(cmd, risk.Read, "Get one config")
	return cmd
}

func (a *App) newConfigListCmd() *cobra.Command {
	var dataID, group, ns string
	var page, pageSize int
	cmd := &cobra.Command{
		Use: "list",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := a.requireServer(cmd); err != nil {
				return err
			}
			uc := domain.ConfigUseCase{Port: a.client}
			out, err := uc.List(dataID, group, ns, page, pageSize)
			if err != nil {
				return a.fail(cmd, risk.Read, err)
			}
			return a.ok(cmd, out)
		},
	}
	cmd.Flags().StringVar(&dataID, "data-id", "", "blur dataId filter (may be empty)")
	cmd.Flags().StringVar(&group, "group", "", "blur group filter")
	cmd.Flags().StringVar(&ns, "namespace", "", "tenant/namespace id")
	cmd.Flags().IntVar(&page, "page", 1, "pageNo")
	cmd.Flags().IntVar(&pageSize, "page-size", 100, "pageSize")
	mark(cmd, risk.Read, "List configs")
	return cmd
}

func (a *App) newConfigPublishCmd() *cobra.Command {
	var dataID, group, ns, typ, content, file string
	var overwrite, createOnly bool
	cmd := &cobra.Command{
		Use: "publish",
		Long: "Publish a config. Default is create-only (refuse if dataId exists).\nUse --overwrite to update an existing config. --create-only is an affirmative alias of the default.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if overwrite && createOnly {
				return a.fail(cmd, risk.High, errBothCreateOverwrite())
			}
			body := content
			if file != "" {
				b, err := os.ReadFile(file)
				if err != nil {
					return a.fail(cmd, risk.High, err)
				}
				body = string(b)
			}
			if body == "" && file == "" && content == "" {
				return a.fail(cmd, risk.High, errNeedContent())
			}
			create := !overwrite
			would := map[string]any{
				"action":        "config.publish",
				"data_id":       dataID,
				"group":         group,
				"namespace":     ns,
				"type":          typ,
				"create_only":   create,
				"overwrite":     overwrite,
				"content_bytes": len([]byte(body)),
			}
			stop, err := a.gateHigh(cmd, "config.publish", would)
			if stop || err != nil {
				return err
			}
			if err := a.requireServer(cmd); err != nil {
				return err
			}
			uc := domain.ConfigUseCase{Port: a.client}
			out, err := uc.Publish(dataID, group, body, ns, typ, create)
			if err != nil {
				return a.fail(cmd, risk.High, err)
			}
			return a.ok(cmd, out)
		},
	}
	cmd.Flags().StringVar(&dataID, "data-id", "", "dataId (required)")
	cmd.Flags().StringVar(&group, "group", "DEFAULT_GROUP", "group")
	cmd.Flags().StringVar(&ns, "namespace", "", "tenant/namespace id")
	cmd.Flags().StringVar(&typ, "type", "text", "config type")
	cmd.Flags().StringVar(&content, "content", "", "inline content")
	cmd.Flags().StringVar(&file, "file", "", "read content from file")
	cmd.Flags().BoolVar(&overwrite, "overwrite", false, "allow update of existing dataId")
	cmd.Flags().BoolVar(&createOnly, "create-only", false, "affirmative alias of default create-only")
	_ = cmd.MarkFlagRequired("data-id")
	mark(cmd, risk.High, "Publish config (create-only by default)")
	return cmd
}

func (a *App) newConfigHistoryCmd() *cobra.Command {
	var dataID, group, ns string
	var page, pageSize int
	cmd := &cobra.Command{
		Use: "history",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := a.requireServer(cmd); err != nil {
				return err
			}
			uc := domain.ConfigUseCase{Port: a.client}
			out, err := uc.History(dataID, group, ns, page, pageSize)
			if err != nil {
				return a.fail(cmd, risk.Read, err)
			}
			return a.ok(cmd, out)
		},
	}
	cmd.Flags().StringVar(&dataID, "data-id", "", "dataId (required)")
	cmd.Flags().StringVar(&group, "group", "DEFAULT_GROUP", "group")
	cmd.Flags().StringVar(&ns, "namespace", "", "tenant/namespace id")
	cmd.Flags().IntVar(&page, "page", 1, "pageNo")
	cmd.Flags().IntVar(&pageSize, "page-size", 20, "pageSize")
	_ = cmd.MarkFlagRequired("data-id")
	mark(cmd, risk.Read, "List config history")
	return cmd
}

func (a *App) newConfigRollbackCmd() *cobra.Command {
	var dataID, group, ns, nid string
	cmd := &cobra.Command{
		Use: "rollback",
		RunE: func(cmd *cobra.Command, _ []string) error {
			would := map[string]any{
				"action": "config.rollback", "data_id": dataID, "group": group, "namespace": ns, "nid": nid,
			}
			stop, err := a.gateHigh(cmd, "config.rollback", would)
			if stop || err != nil {
				return err
			}
			if err := a.requireServer(cmd); err != nil {
				return err
			}
			uc := domain.ConfigUseCase{Port: a.client}
			out, err := uc.Rollback(dataID, group, ns, nid)
			if err != nil {
				return a.fail(cmd, risk.High, err)
			}
			return a.ok(cmd, out)
		},
	}
	cmd.Flags().StringVar(&dataID, "data-id", "", "dataId (required)")
	cmd.Flags().StringVar(&group, "group", "DEFAULT_GROUP", "group")
	cmd.Flags().StringVar(&ns, "namespace", "", "tenant/namespace id")
	cmd.Flags().StringVar(&nid, "nid", "", "history id (required)")
	_ = cmd.MarkFlagRequired("data-id")
	_ = cmd.MarkFlagRequired("nid")
	mark(cmd, risk.High, "Rollback config to a history nid")
	return cmd
}

func errBothCreateOverwrite() error {
	return errString("use either --create-only or --overwrite, not both")
}
func errNeedContent() error {
	return errString("publish needs --file or --content")
}
func errString(s string) error { return &strErr{s} }

type strErr struct{ s string }

func (e *strErr) Error() string { return e.s }

// silence unused strconv in case of future use
var _ = strconv.Itoa
