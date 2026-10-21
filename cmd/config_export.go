package cmd

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/xiaoxiaolin0918/nacos-cli/internal/backup"
	"github.com/xiaoxiaolin0918/nacos-cli/internal/domain"
	"github.com/xiaoxiaolin0918/nacos-cli/internal/output"
	"github.com/xiaoxiaolin0918/nacos-cli/internal/risk"
)

func (a *App) newConfigExportCmd() *cobra.Command {
	var dataID, group, ns, outFile, dir string
	var all bool
	var page, pageSize int
	cmd := &cobra.Command{
		Use: "export",
		Long: `Export config content to a file or directory (read-only against Nacos).

--out writes one file. --dir writes group/dataId files (safe names).
--all exports every page across all groups (unless --group / --page is given).
Without --all, --data-id is required. --dry-run writes nothing and does not call Nacos.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := domain.ValidateExportArgs(dataID, all); err != nil {
				return a.fail(cmd, risk.Read, err)
			}
			if outFile == "" && dir == "" {
				return a.fail(cmd, risk.Read, errString("export needs --out or --dir"))
			}
			if outFile != "" && dir != "" {
				return a.fail(cmd, risk.Read, errString("use either --out or --dir, not both"))
			}
			if outFile != "" && all {
				return a.fail(cmd, risk.Read, errString("--out is for a single data-id; use --dir with --all"))
			}
			filterID := dataID
			if all {
				filterID = ""
				if !cmd.Flags().Changed("group") {
					group = ""
				}
			}
			allPages := all && !cmd.Flags().Changed("page")
			if a.dryRun {
				return a.write(cmd, output.Envelope{OK: true, DryRun: true, Risk: risk.Read, Would: map[string]any{
					"action": "config.export", "data_id": filterID, "group": group, "namespace": ns,
					"out": outFile, "dir": dir, "all": all, "all_pages": allPages,
				}})
			}
			if err := a.requireServer(cmd); err != nil {
				return err
			}
			uc := domain.ConfigUseCase{Port: a.client}
			if outFile != "" {
				body, err := uc.ExportOne(dataID, group, ns)
				if err != nil {
					return a.fail(cmd, risk.Read, err)
				}
				if d := filepath.Dir(outFile); d != "." && d != "" {
					if err := os.MkdirAll(d, 0o755); err != nil {
						return a.fail(cmd, risk.Read, err)
					}
				}
				if err := os.WriteFile(outFile, []byte(body), 0o644); err != nil {
					return a.fail(cmd, risk.Read, err)
				}
				abs, _ := filepath.Abs(outFile)
				return a.ok(cmd, map[string]any{
					"data_id": dataID,
					"group":   group,
					"path":    abs,
					"bytes":   len(body),
				})
			}
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
			if err := os.MkdirAll(dir, 0o755); err != nil {
				return a.fail(cmd, risk.Read, err)
			}
			written := make([]map[string]any, 0, len(entries))
			for _, e := range entries {
				it, err := backup.WriteConfig(dir, e.DataID, e.Group, e.Content)
				if err != nil {
					return a.fail(cmd, risk.Read, err)
				}
				written = append(written, map[string]any{
					"data_id":  it.DataID,
					"group":    it.Group,
					"rel_path": it.RelPath,
					"bytes":    it.Bytes,
				})
			}
			abs, _ := filepath.Abs(dir)
			return a.ok(cmd, map[string]any{"dir": abs, "count": len(written), "files": written})
		},
	}
	cmd.Flags().StringVar(&dataID, "data-id", "", "dataId (required unless --all)")
	cmd.Flags().StringVar(&group, "group", "DEFAULT_GROUP", "group (with --all: default all groups)")
	cmd.Flags().StringVar(&ns, "namespace", "", "tenant/namespace id")
	cmd.Flags().StringVar(&outFile, "out", "", "write one config to this file")
	cmd.Flags().StringVar(&dir, "dir", "", "write one or more configs under this directory")
	cmd.Flags().BoolVar(&all, "all", false, "export all pages / all groups")
	cmd.Flags().IntVar(&page, "page", 1, "only this pageNo when --all (default: all pages)")
	cmd.Flags().IntVar(&pageSize, "page-size", 100, "pageSize when --all")
	mark(cmd, risk.Read, "Export config(s) to local file/dir")
	return cmd
}
