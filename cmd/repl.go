package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/xiaoxiaolin0918/nacos-cli/internal/repl"
	"github.com/xiaoxiaolin0918/nacos-cli/internal/risk"
)

func (a *App) newReplCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "repl",
		Aliases: []string{"interactive", "shell"},
		Long: `Start a lightweight interactive shell (nacos>).

Type the same flags/subcommands as the CLI. Built-ins: help, exit|quit|q.
High-risk writes still require --yes or --dry-run on each line.
Outer --server/--profile/--format/--output are prepended when not overridden.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return a.runREPL(cmd)
		},
	}
	mark(cmd, risk.Read, "Interactive shell (reuses CLI commands)")
	return cmd
}

func (a *App) runREPL(cmd *cobra.Command) error {
	in := cmd.InOrStdin()
	out := cmd.OutOrStdout()
	errOut := cmd.ErrOrStderr()
	fmt.Fprintln(out, `nacos-cli interactive mode. Type "help" or "exit".`)
	br := bufio.NewReader(in)
	for {
		if isTerminalReader(in) {
			fmt.Fprint(out, "nacos> ")
		}
		line, err := br.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				fmt.Fprintln(out)
				return nil
			}
			return err
		}
		line = strings.TrimRight(line, "\r\n")
		if strings.TrimSpace(line) == "" {
			continue
		}
		args, err := repl.SplitArgs(line)
		if err != nil {
			fmt.Fprintln(errOut, "parse error:", err)
			continue
		}
		if repl.IsExit(args) {
			return nil
		}
		if repl.IsHelp(args) {
			printReplHelp(out)
			continue
		}
		if err := a.execReplLine(cmd, args); err != nil {
			// PrintedError already wrote JSON; still show a short stderr hint for humans.
			if _, ok := err.(*PrintedError); !ok {
				fmt.Fprintln(errOut, "Error:", err)
			}
		}
	}
}

func printReplHelp(w io.Writer) {
	fmt.Fprintln(w, `Commands: same as CLI (config, service, namespace, auth, user, role, permission,`)
	fmt.Fprintln(w, `profile, doctor, whoami, skills, backup, ...). Built-ins: help, exit|quit|q.`)
	fmt.Fprintln(w, `Example: namespace list --output pretty`)
	fmt.Fprintln(w, `Writes still need --dry-run or --yes on the line.`)
}

func (a *App) execReplLine(parent *cobra.Command, args []string) error {
	full := append(a.persistentArgs(), args...)
	root := NewRoot()
	root.SetOut(parent.OutOrStdout())
	root.SetErr(parent.ErrOrStderr())
	root.SetIn(parent.InOrStdin())
	root.SetArgs(full)
	return root.Execute()
}

// persistentArgs rebuilds selected root flags so REPL lines inherit session context.
func (a *App) persistentArgs() []string {
	var out []string
	if strings.TrimSpace(a.flagServer) != "" {
		out = append(out, "--server", a.flagServer)
	}
	if strings.TrimSpace(a.flagUsername) != "" {
		out = append(out, "--username", a.flagUsername)
	}
	if strings.TrimSpace(a.flagNS) != "" {
		out = append(out, "--namespace", a.flagNS)
	}
	if strings.TrimSpace(a.configFile) != "" {
		out = append(out, "--config", a.configFile)
	} else if strings.TrimSpace(a.profileName) != "" && a.profileName != "default" {
		out = append(out, "--profile", a.profileName)
	} else if strings.TrimSpace(a.flagProfile) != "" && a.flagProfile != "default" {
		out = append(out, "--profile", a.flagProfile)
	}
	if strings.TrimSpace(a.flagOutput) != "" {
		out = append(out, "--output", a.flagOutput)
	} else if strings.TrimSpace(a.format) != "" && a.format != "json" {
		out = append(out, "--format", a.format)
	} else if strings.TrimSpace(a.flagFormat) != "" && a.flagFormat != "json" {
		out = append(out, "--format", a.flagFormat)
	}
	return out
}

func isTerminalReader(r io.Reader) bool {
	f, ok := r.(*os.File)
	if !ok {
		return false
	}
	return term.IsTerminal(int(f.Fd()))
}

func maybeStartReplOnRoot(a *App, cmd *cobra.Command, args []string) error {
	if len(args) > 0 {
		return fmt.Errorf("unknown command %q for %q", args[0], cmd.CommandPath())
	}
	if isTerminalReader(cmd.InOrStdin()) {
		return a.runREPL(cmd)
	}
	return cmd.Help()
}
