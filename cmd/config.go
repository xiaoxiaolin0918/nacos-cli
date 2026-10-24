package cmd

import (
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/xiaoxiaolin0918/nacos-cli/internal/domain"
	"github.com/xiaoxiaolin0918/nacos-cli/internal/risk"
)

func (a *App) newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "config", Short: "Config center: get/list/publish/history/rollback/export/delete", RunE: helpOnly}
	cmd.AddCommand(
		a.newConfigGetCmd(),
		a.newConfigListCmd(),
		a.newConfigPublishCmd(),
		a.newConfigHistoryCmd(),
		a.newConfigRollbackCmd(),
		a.newConfigExportCmd(),
		a.newConfigDeleteCmd(),
	)
	return cmd
}

func (a *App) effectiveNS(explicit string) string {
	return domain.EffectiveNamespace(explicit, a.settings.Namespace)
}

func (a *App) newConfigGetCmd() *cobra.Command {
	var dataID, group, ns string
	cmd := &cobra.Command{
		Use: "get",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := a.requireServer(cmd); err != nil {
				return err
			}
			eff := a.effectiveNS(ns)
			uc := domain.ConfigUseCase{Port: a.client}
			out, err := uc.Get(dataID, group, eff)
			if err != nil {
				return a.fail(cmd, risk.Read, err)
			}
			return a.ok(cmd, map[string]any{
				"data_id":             dataID,
				"group":               group,
				"namespace":           eff,
				"effective_namespace": eff,
				"content":             out,
			})
		},
	}
	cmd.Flags().StringVar(&dataID, "data-id", "", "dataId (required)")
	cmd.Flags().StringVar(&group, "group", "DEFAULT_GROUP", "group")
	cmd.Flags().StringVar(&ns, "namespace", "", "tenant/namespace id (overrides default; display name public maps to built-in empty id)")
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
			eff := a.effectiveNS(ns)
			uc := domain.ConfigUseCase{Port: a.client}
			out, err := uc.List(domain.BlurPattern(dataID), domain.BlurPattern(group), eff, page, pageSize)
			if err != nil {
				return a.fail(cmd, risk.Read, err)
			}
			if m, ok := out.(map[string]any); ok {
					m["namespace"] = eff
					m["effective_namespace"] = eff
					return a.ok(cmd, m)
				}
				return a.ok(cmd, map[string]any{"result": out, "namespace": eff, "effective_namespace": eff})
		},
	}
	cmd.Flags().StringVar(&dataID, "data-id", "", "blur dataId filter (auto-wrapped as *x* unless it contains *)")
	cmd.Flags().StringVar(&group, "group", "", "blur group filter (auto-wrapped as *x* unless it contains *)")
	cmd.Flags().StringVar(&ns, "namespace", "", "tenant/namespace id")
	cmd.Flags().IntVar(&page, "page", 1, "pageNo")
	cmd.Flags().IntVar(&pageSize, "page-size", 100, "pageSize")
	mark(cmd, risk.Read, "List configs")
	return cmd
}

func (a *App) newConfigPublishCmd() *cobra.Command {
	var dataID, group, ns, typ, content, file string
	var overwrite, createOnly, changeType bool
	cmd := &cobra.Command{
		Use: "publish",
		Long: "Publish a config. Default is create-only (refuse if dataId exists).\n" +
			"Use --overwrite to update an existing config. --create-only is an affirmative alias of the default.\n" +
			"Type is inferred from dataId extension (.yaml/.yml/.json/.properties/.toml/...); unknown extensions require explicit --type.\n" +
			"Overwriting with a different type requires --change-type.",
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
			typeSet := cmd.Flags().Changed("type")
			resolvedType, err := domain.ResolvePublishType(dataID, typ, typeSet)
			if err != nil {
				return a.fail(cmd, risk.High, err)
			}
			eff := a.effectiveNS(ns)
			create := !overwrite
			would := map[string]any{
				"action":              "config.publish",
				"data_id":             dataID,
				"group":               group,
				"namespace":           eff,
				"effective_namespace": eff,
				"type":                resolvedType,
				"create_only":         create,
				"overwrite":           overwrite,
				"change_type":         changeType,
				"content_bytes":       len([]byte(body)),
			}
			if overwrite && a.client != nil && strings.TrimSpace(a.settings.Server) != "" {
				if existing, err := a.client.ConfigType(dataID, group, eff); err == nil {
					existing = strings.TrimSpace(existing)
					if existing != "" && !strings.EqualFold(existing, resolvedType) {
						would["type_change"] = map[string]any{"from": existing, "to": resolvedType}
						would["warning"] = "config type would change; actual publish requires --change-type"
					}
				}
			}
			stop, err := a.gateHigh(cmd, "config.publish", would)
			if stop || err != nil {
				return err
			}
			if err := a.requireServer(cmd); err != nil {
				return err
			}
			uc := domain.ConfigUseCase{Port: a.client}
			out, err := uc.Publish(dataID, group, body, eff, resolvedType, create, changeType)
			if err != nil {
				return a.fail(cmd, risk.High, err)
			}
			return a.ok(cmd, map[string]any{
				"result":              out,
				"data_id":             dataID,
				"group":               group,
				"namespace":           eff,
				"effective_namespace": eff,
				"type":                resolvedType,
			})
		},
	}
	cmd.Flags().StringVar(&dataID, "data-id", "", "dataId (required)")
	cmd.Flags().StringVar(&group, "group", "DEFAULT_GROUP", "group")
	cmd.Flags().StringVar(&ns, "namespace", "", "tenant/namespace id")
	cmd.Flags().StringVar(&typ, "type", "", "config type (inferred from dataId when omitted)")
	cmd.Flags().StringVar(&content, "content", "", "inline content")
	cmd.Flags().StringVar(&file, "file", "", "read content from file")
	cmd.Flags().BoolVar(&overwrite, "overwrite", false, "allow update of existing dataId")
	cmd.Flags().BoolVar(&createOnly, "create-only", false, "affirmative alias of default create-only")
	cmd.Flags().BoolVar(&changeType, "change-type", false, "allow overwrite to change the stored config type")
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
			eff := a.effectiveNS(ns)
			uc := domain.ConfigUseCase{Port: a.client}
			out, err := uc.History(dataID, group, eff, page, pageSize)
			if err != nil {
				return a.fail(cmd, risk.Read, err)
			}
			if m, ok := out.(map[string]any); ok {
					m["namespace"] = eff
					m["effective_namespace"] = eff
					return a.ok(cmd, m)
				}
				return a.ok(cmd, map[string]any{"result": out, "namespace": eff, "effective_namespace": eff})
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
			eff := a.effectiveNS(ns)
			would := map[string]any{
				"action": "config.rollback", "data_id": dataID, "group": group,
				"namespace": eff, "effective_namespace": eff, "nid": nid,
			}
			stop, err := a.gateHigh(cmd, "config.rollback", would)
			if stop || err != nil {
				return err
			}
			if err := a.requireServer(cmd); err != nil {
				return err
			}
			uc := domain.ConfigUseCase{Port: a.client}
			out, err := uc.Rollback(dataID, group, eff, nid)
			if err != nil {
				return a.fail(cmd, risk.High, err)
			}
			return a.ok(cmd, map[string]any{
				"result": out, "namespace": eff, "effective_namespace": eff,
			})
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

func (a *App) newConfigDeleteCmd() *cobra.Command {
	var dataID, group, ns string
	cmd := &cobra.Command{
		Use:  "delete",
		Long: "Delete one existing config (refuses when it does not exist). Back up first: nacos backup create --data-id <id>.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			eff := a.effectiveNS(ns)
			would := map[string]any{
				"action": "config.delete", "data_id": dataID, "group": group,
				"namespace": eff, "effective_namespace": eff,
			}
			stop, err := a.gateHigh(cmd, "config.delete", would)
			if stop || err != nil {
				return err
			}
			if err := a.requireServer(cmd); err != nil {
				return err
			}
			uc := domain.ConfigUseCase{Port: a.client}
			out, err := uc.Delete(dataID, group, eff)
			if err != nil {
				return a.fail(cmd, risk.High, err)
			}
			return a.ok(cmd, map[string]any{
				"result": out, "namespace": eff, "effective_namespace": eff,
			})
		},
	}
	cmd.Flags().StringVar(&dataID, "data-id", "", "dataId (required)")
	cmd.Flags().StringVar(&group, "group", "DEFAULT_GROUP", "group")
	cmd.Flags().StringVar(&ns, "namespace", "", "tenant/namespace id")
	_ = cmd.MarkFlagRequired("data-id")
	mark(cmd, risk.High, "Delete one config")
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
