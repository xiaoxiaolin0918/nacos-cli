package cmd

import (
	"strings"

	"github.com/spf13/cobra"
)

// applyLimitAlias: if --limit was set, override pageSize (Agent-friendly alias of --page-size).
func applyLimitAlias(cmd *cobra.Command, pageSize *int, limit int) {
	if cmd.Flags().Changed("limit") && limit > 0 {
		*pageSize = limit
	}
}

// addListPageFlags adds --page / --page-size / --limit to a list command.
func addListPageFlags(cmd *cobra.Command, page, pageSize *int, limit *int) {
	cmd.Flags().IntVar(page, "page", 1, "pageNo")
	cmd.Flags().IntVar(pageSize, "page-size", 100, "pageSize")
	cmd.Flags().IntVar(limit, "limit", 0, "alias of --page-size (Agent-friendly; overrides when set)")
}

// applyQueryFallback sets *dst from --query when dst is empty.
func applyQueryFallback(dst *string, query string) {
	if strings.TrimSpace(*dst) == "" && strings.TrimSpace(query) != "" {
		*dst = query
	}
}
