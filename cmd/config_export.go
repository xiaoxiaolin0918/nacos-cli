package cmd

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/xiaoxiaolin0918/nacos-cli/internal/backup"
	"github.com/xiaoxiaolin0918/nacos-cli/internal/domain"
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
--all exports the current list page. Without --all, --data-id is required.`,
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
			if err := a.requireServer(cmd); err != nil {
				return err
			}
			filterID := dataID
			if all {
				filterID = ""
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
			entries, err := uc.CollectForBackup(filterID, group, ns, page, pageSize)
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
	cmd.Flags().StringVar(&group, "group", "DEFAULT_GROUP", "group")
	cmd.Flags().StringVar(&ns, "namespace", "", "tenant/namespace id")
	cmd.Flags().StringVar(&outFile, "out", "", "write one config to this file")
	cmd.Flags().StringVar(&dir, "dir", "", "write one or more configs under this directory")
	cmd.Flags().BoolVar(&all, "all", false, "export current list page")
	cmd.Flags().IntVar(&page, "page", 1, "pageNo when --all")
	cmd.Flags().IntVar(&pageSize, "page-size", 100, "pageSize when --all")
	mark(cmd, risk.Read, "Export config(s) to local file/dir")
	return cmd
}
