package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/xiaoxiaolin0918/nacos-cli/internal/config"
	"github.com/xiaoxiaolin0918/nacos-cli/internal/version"
)

const updateCheckCacheTTL = 24 * time.Hour

// Short client for opportunistic background checks (never blocks business long).
var updateCheckHTTPClient = &http.Client{Timeout: 3 * time.Second}

type updateCheckCache struct {
	CheckedAt time.Time `json:"checked_at"`
	Latest    string    `json:"latest"`
	Status    string    `json:"status"`
	Notified  string    `json:"notified_version,omitempty"`
}

func (a *App) resolveUpdateFlags(prof config.Profile) {
	a.checkUpdateEnabled = a.flagCheckUpdate || envTruthy("NACOS_CHECK_UPDATE") || prof.CheckUpdate
	a.autoUpdateEnabled = a.flagAutoUpdate || envTruthy("NACOS_AUTO_UPDATE") || prof.AutoUpdate
	// Auto-update implies check.
	if a.autoUpdateEnabled {
		a.checkUpdateEnabled = true
	}
}

func updateCheckCachePath() (string, error) {
	dir, err := config.ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "update-check-cache.json"), nil
}

func loadUpdateCheckCache() (updateCheckCache, bool) {
	path, err := updateCheckCachePath()
	if err != nil {
		return updateCheckCache{}, false
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return updateCheckCache{}, false
	}
	var c updateCheckCache
	if json.Unmarshal(b, &c) != nil {
		return updateCheckCache{}, false
	}
	return c, true
}

func saveUpdateCheckCache(c updateCheckCache) {
	path, err := updateCheckCachePath()
	if err != nil {
		return
	}
	_ = os.MkdirAll(filepath.Dir(path), 0o700)
	b, err := json.Marshal(c)
	if err != nil {
		return
	}
	_ = os.WriteFile(path, b, 0o600)
}

func shouldSkipBackgroundUpdate(cmd *cobra.Command) bool {
	if cmd == nil {
		return true
	}
	name := cmd.Name()
	if name == "help" || name == "completion" || name == "commands" || name == "update" {
		return true
	}
	if isUpdateCommand(cmd) || isCommandsCommand(cmd) {
		return true
	}
	// Avoid recurse when --help-json short-circuited.
	if root := cmd.Root(); root != nil {
		if f := root.PersistentFlags().Lookup("help-json"); f != nil && f.Changed {
			v, _ := root.PersistentFlags().GetBool("help-json")
			if v {
				return true
			}
		}
	}
	return false
}

// maybeBackgroundUpdate runs after a successful command when check-update is enabled.
// Never fails the main command. Hints go to stderr only (never JSON stdout).
func (a *App) maybeBackgroundUpdate(cmd *cobra.Command) {
	if a == nil || !a.checkUpdateEnabled {
		return
	}
	if shouldSkipBackgroundUpdate(cmd) {
		return
	}
	cache, ok := loadUpdateCheckCache()
	if ok && time.Since(cache.CheckedAt) < updateCheckCacheTTL && cache.Latest != "" {
		if cache.Status == "outdated" && CompareSemver(version.Version, cache.Latest) == RelOutdated {
			a.emitUpdateHint(cmd, cache.Latest, false)
		}
		return
	}
	res, err := a.quickCheckLatest(cmd.Context())
	if err != nil || res == nil {
		return
	}
	cache = updateCheckCache{
		CheckedAt: time.Now().UTC(),
		Latest:    res.Latest,
		Status:    res.Status,
	}
	saveUpdateCheckCache(cache)
	if res.Status != "outdated" {
		return
	}
	applied := false
	if a.autoUpdateEnabled {
		applied = a.trySilentAutoUpdate(cmd)
	}
	a.emitUpdateHint(cmd, res.Latest, applied)
}

func (a *App) emitUpdateHint(cmd *cobra.Command, latest string, applied bool) {
	w := cmd.ErrOrStderr()
	if applied {
		fmt.Fprintf(w, "nacos: auto-updated to %s (was %s); re-run if needed\n", latest, version.Version)
		return
	}
	fmt.Fprintf(w, "nacos: newer version %s available (current %s); run: nacos update --yes\n", latest, version.Version)
}

func (a *App) quickCheckLatest(ctx context.Context) (*updateCheckResult, error) {
	old := updateHTTPClient
	updateHTTPClient = updateCheckHTTPClient
	defer func() { updateHTTPClient = old }()
	// Minimal cobra.Command for context only.
	cmd := &cobra.Command{}
	cmd.SetContext(ctx)
	return a.checkUpdate(cmd)
}

func (a *App) trySilentAutoUpdate(cmd *cobra.Command) bool {
	res, err := a.applyUpdate(cmd)
	if err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "nacos: auto-update failed: %v (rolled back if possible)\n", err)
		return false
	}
	return res != nil && res.Applied
}

// attachBackgroundUpdateHook wires PersistentPostRunE once on root.
func attachBackgroundUpdateHook(root *cobra.Command, a *App) {
	prev := root.PersistentPostRunE
	root.PersistentPostRunE = func(cmd *cobra.Command, args []string) error {
		if prev != nil {
			if err := prev(cmd, args); err != nil {
				return err
			}
		}
		a.maybeBackgroundUpdate(cmd)
		return nil
	}
}
