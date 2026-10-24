package cmd

import (
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/xiaoxiaolin0918/nacos-cli/internal/backup"
	"github.com/xiaoxiaolin0918/nacos-cli/internal/domain"
	"github.com/xiaoxiaolin0918/nacos-cli/internal/output"
	"github.com/xiaoxiaolin0918/nacos-cli/internal/risk"
)

func (a *App) newBackupCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "backup", Short: "Local config backups under backups/", RunE: helpOnly}
	cmd.AddCommand(
		a.newBackupPathCmd(),
		a.newBackupListCmd(),
		a.newBackupCreateCmd(),
	)
	return cmd
}

func (a *App) newBackupPathCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "path",
		RunE: func(cmd *cobra.Command, _ []string) error {
			wd, _ := os.Getwd()
			rootDir := backup.Root(wd)
			abs, _ := filepath.Abs(rootDir)
			return a.ok(cmd, map[string]any{
				"backup_root": abs,
				"env":         "NACOS_BACKUP_DIR",
			})
		},
	}
	mark(cmd, risk.Read, "Show local backup root path")
	return cmd
}

func (a *App) newBackupListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "list",
		RunE: func(cmd *cobra.Command, _ []string) error {
			wd, _ := os.Getwd()
			rootDir := backup.Root(wd)
			list, err := backup.List(rootDir)
			if err != nil {
				return a.fail(cmd, risk.Read, err)
			}
			abs, _ := filepath.Abs(rootDir)
			return a.ok(cmd, map[string]any{"backup_root": abs, "backups": list})
		},
	}
	mark(cmd, risk.Read, "List local backup folders")
	return cmd
}

func (a *App) newBackupCreateCmd() *cobra.Command {
	var dataID, group, ns string
	var page, pageSize int
	var all bool
	cmd := &cobra.Command{
		Use: "create",
		Long: `Snapshot configs from Nacos into backups/<stamp>/ (gitignored).

Reads Open API only; writes local files. Use --dry-run to preview without writing.
--all snapshots every page across all groups of one namespace (narrow with --group / --page).
Without --all, --data-id is required. Same-second runs get a -2, -3 suffix (never merged).`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := domain.ValidateExportArgs(dataID, all); err != nil {
				return a.fail(cmd, risk.Read, err)
			}
			filterID := dataID
			if all {
				filterID = ""
				if !cmd.Flags().Changed("group") {
					group = ""
				}
			}
			allPages := all && !cmd.Flags().Changed("page")
			ns = a.effectiveNS(ns)
			if err := a.requireServer(cmd); err != nil {
				return err
			}
			uc := domain.ConfigUseCase{Port: a.client}
			var entries []domain.ConfigEntry
			var err error
			if allPages {
				entries, err = uc.CollectAllPages(filterID, group, ns, pageSize)
			} else {
				entries, err = uc.CollectForBackup(filterID, group, ns, page, pageSize)
			}
			if err != nil {
				return a.fail(cmd, risk.Read, err)
			}
			names := make([]string, 0, len(entries))
			for _, e := range entries {
				names = append(names, e.Group+"/"+e.DataID)
			}
			would := map[string]any{
				"action":    "backup.create",
				"data_id":   filterID,
				"group":     group,
				"namespace": ns,
				"effective_namespace": ns,
				"page":      page,
				"page_size": pageSize,
				"all":       all,
				"all_pages": allPages,
				"count":     len(entries),
				"items":     names,
			}
			if a.dryRun {
				return a.write(cmd, output.Envelope{
					OK:     true,
					DryRun: true,
					Risk:   risk.Read,
					Would:  would,
				})
			}
			wd, _ := os.Getwd()
			rootDir := backup.Root(wd)
			if err := backup.EnsureWritable(rootDir); err != nil {
				return a.fail(cmd, risk.Read, err)
			}
			stamp := backup.NewStamp(time.Now())
			dir, err := backup.CreateDir(rootDir, stamp)
			if err != nil {
				return a.fail(cmd, risk.Read, err)
			}
			items := make([]backup.Item, 0, len(entries))
			for _, e := range entries {
				it, err := backup.WriteConfig(dir, e.DataID, e.Group, e.Content)
				if err != nil {
					return a.fail(cmd, risk.Read, err)
				}
				it.Namespace = ns
				items = append(items, it)
			}
			man := backup.Manifest{
				CreatedAt: stamp,
				Server:    a.settings.Server,
				Namespace: ns,
				Group:     group,
				DataID:    filterID,
				Items:     items,
			}
			if err := backup.WriteManifest(dir, man); err != nil {
				return a.fail(cmd, risk.Read, err)
			}
			return a.ok(cmd, map[string]any{
				"backup_dir": dir,
				"stamp":      stamp,
				"count":      len(items),
				"manifest":   "manifest.json",
			})
		},
	}
	cmd.Flags().StringVar(&dataID, "data-id", "", "dataId (required unless --all)")
	cmd.Flags().StringVar(&group, "group", "DEFAULT_GROUP", "group filter (with --all: default all groups)")
	cmd.Flags().StringVar(&ns, "namespace", "", "tenant/namespace id")
	cmd.Flags().BoolVar(&all, "all", false, "backup all pages / all groups of the namespace")
	cmd.Flags().IntVar(&page, "page", 1, "only this pageNo with --all (default: all pages)")
	cmd.Flags().IntVar(&pageSize, "page-size", 100, "pageSize when listing")
	mark(cmd, risk.Read, "Create a local config backup snapshot")
	return cmd
}
