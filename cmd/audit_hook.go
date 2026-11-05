package cmd

import (
	"testing"
	"time"

	"github.com/spf13/cobra"

	"github.com/xiaoxiaolin0918/nacos-cli/internal/audit"
)

func (a *App) ensureAuditStore() (*audit.Store, error) {
	if a.auditStore != nil {
		return a.auditStore, nil
	}
	// Avoid writing into the developer's ~/.nacos-cli during go test.
	if testing.Testing() {
		return nil, nil
	}
	s, err := audit.Default()
	if err != nil {
		return nil, err
	}
	a.auditStore = s
	return s, nil
}

func (a *App) recordOutcome(command, riskLevel string, ok bool, errMsg string) {
	store, err := a.ensureAuditStore()
	if err != nil || store == nil {
		return
	}
	msg := audit.RedactError(errMsg, a.secrets)
	_ = store.Append(audit.Entry{
		Time:    time.Now().UTC().Format(time.RFC3339),
		Command: command,
		Risk:    riskLevel,
		Profile: a.profileName,
		OK:      ok,
		Error:   msg,
	})
}

func cmdRisk(cmd *cobra.Command) string {
	if cmd == nil {
		return ""
	}
	if cmd.Annotations != nil {
		if r := cmd.Annotations["risk"]; r != "" {
			return r
		}
	}
	return ""
}
