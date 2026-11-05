package cmd

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/spf13/cobra"

	"github.com/xiaoxiaolin0918/nacos-cli/internal/audit"
	"github.com/xiaoxiaolin0918/nacos-cli/internal/config"
	"github.com/xiaoxiaolin0918/nacos-cli/internal/risk"
)

func (a *App) newWebCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "web",
		Short: "Open local loopback console for profiles and audit log",
		Long: `Starts a browser console bound to 127.0.0.1 only.

Manage ~/.nacos-cli/*.conf connection profiles and *.meta.yaml notes,
and browse ~/.nacos-cli/audit.jsonl. Auth check uses a read-only login/probe against Nacos; no config/service writes.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return a.runWebConsole(cmd)
		},
	}
	mark(cmd, risk.Read, "Local profile/audit web console (127.0.0.1)")
	return cmd
}

func (a *App) runWebConsole(cmd *cobra.Command) error {
	dir, err := config.ProfileDir()
	if err != nil {
		return a.fail(cmd, risk.Read, err)
	}
	store, err := a.ensureAuditStore()
	if err != nil {
		return a.fail(cmd, risk.Read, err)
	}
	if store == nil {
		store, err = audit.Default()
		if err != nil {
			return a.fail(cmd, risk.Read, err)
		}
		a.auditStore = store
	}
	srv := newWebServer(dir, store)

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return a.fail(cmd, risk.Read, err)
	}
	host, port, _ := net.SplitHostPort(ln.Addr().String())
	if host != "127.0.0.1" && host != "::1" {
		_ = ln.Close()
		return a.fail(cmd, risk.Read, fmt.Errorf("refusing non-loopback bind %s", host))
	}
	url := fmt.Sprintf("http://127.0.0.1:%s/", port)

	httpSrv := &http.Server{Handler: srv.Handler()}
	go func() { _ = httpSrv.Serve(ln) }()

	_ = openBrowser(url)
	fmt.Fprintf(cmd.ErrOrStderr(), "nacos web listening on %s (Ctrl+C to stop)\n", url)
	fmt.Fprintf(cmd.ErrOrStderr(), "Open %s if the browser did not launch\n", url)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	<-sigCh
	signal.Stop(sigCh)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_ = httpSrv.Shutdown(ctx)
	return a.ok(cmd, map[string]any{"ok": true, "stopped": true, "url": url})
}

func isWebCommand(cmd *cobra.Command) bool {
	for c := cmd; c != nil; c = c.Parent() {
		if c.Name() == "web" {
			return true
		}
	}
	return false
}
