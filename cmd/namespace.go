package cmd

import (
	"strings"

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
	var query string
	var limit int
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
			out = filterNamespaceList(out, query, limit)
			return a.ok(cmd, out)
		},
	}
	cmd.Flags().StringVar(&query, "query", "", "client-side substring filter on namespace id/name")
	cmd.Flags().IntVar(&limit, "limit", 0, "client-side max items (0 = no limit)")
	mark(cmd, risk.Read, "List namespaces")
	return cmd
}

func filterNamespaceList(out any, query string, limit int) any {
	q := strings.TrimSpace(strings.ToLower(query))
	arr, ok := out.([]any)
	if !ok {
		if m, ok := out.(map[string]any); ok {
			for _, k := range []string{"data", "namespaces", "pageItems"} {
				if raw, ok := m[k].([]any); ok {
					m[k] = filterNamespaceItems(raw, q, limit)
					return m
				}
			}
		}
		return out
	}
	return filterNamespaceItems(arr, q, limit)
}

func filterNamespaceItems(arr []any, q string, limit int) []any {
	kept := make([]any, 0, len(arr))
	for _, it := range arr {
		if q != "" {
			name := ""
			switch v := it.(type) {
			case string:
				name = v
			case map[string]any:
				for _, nk := range []string{"namespace", "namespaceShowName", "namespaceId", "showName", "name"} {
					if s, ok := v[nk].(string); ok && s != "" {
						name += " " + s
					}
				}
			}
			if !strings.Contains(strings.ToLower(name), q) {
				continue
			}
		}
		kept = append(kept, it)
		if limit > 0 && len(kept) >= limit {
			break
		}
	}
	return kept
}
