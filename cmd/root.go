package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/xiaoxiaolin0918/nacos-cli/internal/client"
	"github.com/xiaoxiaolin0918/nacos-cli/internal/config"
	"github.com/xiaoxiaolin0918/nacos-cli/internal/credstore"
	"github.com/xiaoxiaolin0918/nacos-cli/internal/output"
	"github.com/xiaoxiaolin0918/nacos-cli/internal/risk"
	"github.com/xiaoxiaolin0918/nacos-cli/internal/version"
)

type App struct {
	flagServer   string
	flagUsername string
	flagFormat   string
	flagOutput   string
	flagDryRun   bool
	flagYes      bool
	flagTimeout  string
	flagConfig   string
	flagProfile  string
	flagNS       string

	settings    config.Settings
	format      string
	dryRun      bool
	yes         bool
	client      *client.Client
	credPath    string
	secrets     []string
	profileName string
	profile     config.Profile
	configFile  string
}

func Main() int {
	root := NewRoot()
	err := root.Execute()
	if err == nil {
		return 0
	}
	var pe *PrintedError
	if errors.As(err, &pe) {
		return pe.ExitCode()
	}
	fmt.Fprintln(os.Stderr, "Error:", err)
	return 1
}

func NewRoot() *cobra.Command {
	a := &App{}
	root := &cobra.Command{
		Use:   "nacos",
		Short: "Nacos ops CLI (Open API: config / service / namespace / auth / ACL)",
		Long: `nacos-cli talks to a Nacos server over Open API.

Auth uses accessToken from POST /v1/auth/login (or AUTH_DISABLED).
Credentials (token only, never password) live under the OS config dir.

Profiles: ~/.nacos-cli/<name>.conf (YAML). Select with --profile (default: default)
or an explicit file via --config. Priority: flags > --config > --profile > env/.env > defaults.

Output: --format json|pretty (default json). --output is an alias for official CLI parity.

Risk:
  read             runs immediately
  high-risk-write  requires --yes; --dry-run previews without calling Nacos

Config: flags > --config > --profile > env > .env > defaults.
Skills: nacos skills list|read|path

Bare nacos on a TTY starts the interactive REPL (same as: nacos repl).`,
		SilenceUsage:  true,
		SilenceErrors: true,
		Version:       version.Version,
		RunE: func(cmd *cobra.Command, args []string) error {
			return maybeStartReplOnRoot(a, cmd, args)
		},
	}
	root.SetVersionTemplate("{\"ok\":true,\"result\":{\"name\":\"nacos-cli\",\"version\":\"{{.Version}}\"}}\n")
	root.CompletionOptions.DisableDefaultCmd = true
	root.PersistentPreRunE = a.resolve

	p := root.PersistentFlags()
	p.StringVar(&a.flagServer, "server", "", "Nacos server root, e.g. http://127.0.0.1:8848 (env NACOS_SERVER)")
	p.StringVar(&a.flagUsername, "username", "", "login username (env NACOS_USERNAME)")
	p.StringVar(&a.flagNS, "namespace", "", "default namespace id (env NACOS_NAMESPACE)")
	p.StringVar(&a.flagProfile, "profile", "default", "profile name under ~/.nacos-cli/<name>.conf")
	p.StringVarP(&a.flagConfig, "config", "c", "", "explicit profile file path (overrides --profile)")
	p.StringVar(&a.flagFormat, "format", "json", "output format: json or pretty")
	p.StringVar(&a.flagOutput, "output", "", "alias of --format (json|pretty); official CLI parity")
	p.BoolVar(&a.flagDryRun, "dry-run", false, "print the would-be action and do not call Nacos")
	p.BoolVarP(&a.flagYes, "yes", "y", false, "confirm a high-risk write")

	root.AddCommand(
		a.newConfigCmd(),
		a.newServiceCmd(),
		a.newNamespaceCmd(),
		a.newAuthCmd(),
		a.newWhoamiCmd(),
		a.newDoctorCmd(),
		a.newSkillsCmd(),
		a.newUserCmd(),
		a.newRoleCmd(),
		a.newPermissionCmd(),
		a.newProfileCmd(),
		a.newBackupCmd(),
		a.newReplCmd(),
	)
	silence(root)
	return root
}

func silence(cmd *cobra.Command) {
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true
	for _, c := range cmd.Commands() {
		silence(c)
	}
}

func helpOnly(cmd *cobra.Command, _ []string) error {
	return cmd.Help()
}

func mark(cmd *cobra.Command, level, short string) {
	cmd.Short = short + " (Risk: " + level + ")"
	if cmd.Annotations == nil {
		cmd.Annotations = map[string]string{}
	}
	cmd.Annotations["risk"] = level
	note := "Risk: " + level + "."
	switch level {
	case risk.High:
		note += "\nExecution requires --yes. --dry-run prints the action and does not send it, even when --yes is set."
	case risk.Write:
		note += "\n--dry-run prints the action and does not send it."
	}
	if strings.TrimSpace(cmd.Long) == "" {
		cmd.Long = note
	} else {
		cmd.Long = strings.TrimRight(cmd.Long, "\n") + "\n\n" + note
	}
}

func (a *App) resolve(cmd *cobra.Command, _ []string) error {
	a.dryRun = a.flagDryRun
	a.yes = a.flagYes
	a.profileName = strings.TrimSpace(a.flagProfile)
	if a.profileName == "" {
		a.profileName = "default"
	}
	a.configFile = strings.TrimSpace(a.flagConfig)

	if isSkillsCommand(cmd) || isProfileCommand(cmd) || cmd.Name() == "help" {
		// Still resolve format for JSON envelope on profile commands.
		if err := a.resolveFormat(cmd, config.Profile{}); err != nil {
			return a.fail(cmd, "", err)
		}
		return nil
	}

	_ = config.LoadDotEnv(config.FindDotEnv())
	s := config.FromEnv()

	var prof config.Profile
	var profOK bool
	if a.configFile != "" {
		p, ok, err := config.LoadProfileFile(a.configFile)
		if err != nil {
			return a.fail(cmd, "", err)
		}
		if !ok {
			return a.fail(cmd, "", fmt.Errorf("--config file not found: %s", a.configFile))
		}
		prof, profOK = p, true
	} else {
		p, ok, err := config.LoadProfile(a.profileName)
		if err != nil {
			return a.fail(cmd, "", err)
		}
		// Missing default/named profile is fine; fall through to env/.env.
		prof, profOK = p, ok
	}
	if profOK {
		s = config.MergeSettings(s, prof)
		a.profile = prof
	}

	if changed(cmd, "server") && strings.TrimSpace(a.flagServer) != "" {
		s.Server = a.flagServer
	}
	if changed(cmd, "username") && strings.TrimSpace(a.flagUsername) != "" {
		s.Username = a.flagUsername
	}
	if changed(cmd, "namespace") {
		s.Namespace = config.NormalizeNamespaceID(a.flagNS)
	}
	a.settings = s

	if err := a.resolveFormat(cmd, prof); err != nil {
		return a.fail(cmd, "", err)
	}

	credPath, err := config.CredentialsPath()
	if err != nil {
		return a.fail(cmd, "", err)
	}
	a.credPath = credPath

	c := client.New(s)
	if cred, ok, _ := credstore.Load(credPath); ok {
		if cred.AuthMode == "disabled" {
			c.ApplyToken("", "disabled")
		} else if cred.AccessToken != "" {
			c.ApplyToken(cred.AccessToken, firstNonEmpty(cred.AuthMode, "token"))
			a.track(cred.AccessToken)
		}
	}
	a.client = c
	return nil
}

func (a *App) resolveFormat(cmd *cobra.Command, prof config.Profile) error {
	format := a.flagFormat
	if changed(cmd, "output") && strings.TrimSpace(a.flagOutput) != "" {
		format = a.flagOutput
	} else if !changed(cmd, "format") && strings.TrimSpace(prof.Format) != "" {
		format = prof.Format
	}
	norm, err := config.NormalizeFormat(format)
	if err != nil {
		return err
	}
	a.format = norm
	return nil
}

func changed(cmd *cobra.Command, name string) bool {
	if cmd.Flags().Changed(name) {
		return true
	}
	if cmd.Root() != nil && cmd.Root().PersistentFlags().Changed(name) {
		return true
	}
	return false
}

func (a *App) track(secret string) {
	if secret != "" {
		a.secrets = append(a.secrets, secret)
	}
}

func (a *App) write(cmd *cobra.Command, env output.Envelope) error {
	w := cmd.OutOrStdout()
	if !env.OK {
		w = cmd.ErrOrStderr()
	}
	return output.Write(w, a.format, env, a.secrets)
}

func (a *App) ok(cmd *cobra.Command, v any) error {
	return a.write(cmd, output.Envelope{OK: true, Result: output.MustResult(v)})
}

func (a *App) fail(cmd *cobra.Command, level string, err error) error {
	body := output.ErrorBody{Type: "cli", Message: err.Error()}
	if level != "" {
		body.Risk = level
	}
	env := output.Envelope{OK: false, Error: body, Risk: level}
	if werr := a.write(cmd, env); werr != nil {
		return werr
	}
	return &PrintedError{Msg: body.Message}
}

func (a *App) failConfirmation(cmd *cobra.Command, g risk.GateResult) error {
	body := output.ErrorBody{
		Type:    "confirmation",
		Subtype: "confirmation_required",
		Message: g.Error(),
		Hint:    g.Hint,
		Risk:    g.Risk,
		Action:  g.Action,
	}
	env := output.Envelope{OK: false, Error: body, Risk: g.Risk, Hint: g.Hint}
	if werr := a.write(cmd, env); werr != nil {
		return werr
	}
	return &PrintedError{Msg: body.Message, Code: risk.ExitConfirmationRequired}
}

func (a *App) gateHigh(cmd *cobra.Command, action string, would map[string]any) (stop bool, err error) {
	if a.dryRun {
		env := output.Envelope{
			OK:     true,
			DryRun: true,
			Risk:   risk.High,
			Would:  would,
		}
		if would == nil {
			env.Would = map[string]any{"action": action}
		} else if _, ok := would["action"]; !ok {
			would["action"] = action
			env.Would = would
		}
		env.WouldRequire = []string{"--yes"}
		if werr := a.write(cmd, env); werr != nil {
			return true, werr
		}
		return true, nil
	}
	if err := risk.HighRisk().Check(a.yes, action); err != nil {
		if g, ok := err.(risk.GateResult); ok {
			return true, a.failConfirmation(cmd, g)
		}
		return true, a.fail(cmd, risk.High, err)
	}
	return false, nil
}

func (a *App) requireServer(cmd *cobra.Command) error {
	if strings.TrimSpace(a.settings.Server) == "" {
		return a.fail(cmd, "", errors.New("NACOS_SERVER not set; copy .env.example to .env or use --profile / --server"))
	}
	return nil
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func isSkillsCommand(cmd *cobra.Command) bool {
	for c := cmd; c != nil; c = c.Parent() {
		if c.Name() == "skills" {
			return true
		}
	}
	return false
}