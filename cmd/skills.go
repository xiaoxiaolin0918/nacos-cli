package cmd

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"

	"github.com/xiaoxiaolin0918/nacos-cli/internal/risk"
)

var skillVersionRE = regexp.MustCompile(`(?m)^version:\s*["']?([^\s"']+)["']?\s*$`)

func findDiskSkills() string {
	candidates := []string{}
	if exe, err := os.Executable(); err == nil {
		candidates = append(candidates, filepath.Join(filepath.Dir(exe), "skills"))
	}
	if wd, err := os.Getwd(); err == nil {
		candidates = append(candidates, filepath.Join(wd, "skills"))
		candidates = append(candidates, filepath.Join(wd, "..", "skills"))
	}
	for _, c := range candidates {
		if hasNacosSkill(c) {
			abs, err := filepath.Abs(c)
			if err == nil {
				return abs
			}
			return c
		}
	}
	return ""
}

func hasNacosSkill(root string) bool {
	entries, err := os.ReadDir(root)
	if err != nil {
		return false
	}
	for _, e := range entries {
		if e.IsDir() && strings.HasPrefix(e.Name(), "nacos-") {
			if st, err := os.Stat(filepath.Join(root, e.Name(), "SKILL.md")); err == nil && !st.IsDir() {
				return true
			}
		}
	}
	return false
}

func skillsRoot() (root, source string) {
	if disk := findDiskSkills(); disk != "" {
		return disk, "disk"
	}
	return "skills", "missing"
}

func (a *App) newSkillsCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "skills", Short: "List, read, or path bundled agent skills", RunE: helpOnly}

	list := &cobra.Command{
		Use: "list",
		RunE: func(cmd *cobra.Command, _ []string) error {
			a.format = firstNonEmpty(a.format, "json")
			root, source := skillsRoot()
			entries, err := os.ReadDir(root)
			if err != nil {
				return a.fail(cmd, risk.Read, err)
			}
			var items []map[string]any
			for _, e := range entries {
				if !e.IsDir() || !strings.HasPrefix(e.Name(), "nacos-") {
					continue
				}
				md := filepath.Join(root, e.Name(), "SKILL.md")
				if st, err := os.Stat(md); err != nil || st.IsDir() {
					continue
				}
				b, _ := os.ReadFile(md)
				ver := ""
				if m := skillVersionRE.FindSubmatch(b); m != nil {
					ver = string(m[1])
				}
				items = append(items, map[string]any{
					"name": e.Name(), "dir": e.Name(), "version": ver, "path": md,
				})
			}
			return a.ok(cmd, map[string]any{"skills": items, "skills_dir": root, "source": source})
		},
	}
	mark(list, risk.Read, "List available agent skills")

	read := &cobra.Command{
		Use:  "read NAME",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a.format = firstNonEmpty(a.format, "json")
			root, _ := skillsRoot()
			name := args[0]
			p := filepath.Join(root, name, "SKILL.md")
			b, err := os.ReadFile(p)
			if err != nil {
				return a.fail(cmd, risk.Read, err)
			}
			ver := ""
			if m := skillVersionRE.FindSubmatch(b); m != nil {
				ver = string(m[1])
			}
			return a.ok(cmd, map[string]any{"name": name, "path": p, "content": string(b), "version": ver})
		},
	}
	mark(read, risk.Read, "Read a skill SKILL.md")

	pathCmd := &cobra.Command{
		Use:  "path [NAME]",
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a.format = firstNonEmpty(a.format, "json")
			root, source := skillsRoot()
			if len(args) == 0 {
				return a.ok(cmd, map[string]any{"path": root, "source": source})
			}
			name := args[0]
			p := filepath.Join(root, name)
			if st, err := os.Stat(p); err != nil || !st.IsDir() {
				return a.fail(cmd, risk.Read, errString("skill not found: "+name))
			}
			return a.ok(cmd, map[string]any{"name": name, "path": p, "skill_md": filepath.Join(p, "SKILL.md"), "source": source})
		},
	}
	mark(pathCmd, risk.Read, "Print skills directory or one skill path")

	cmd.AddCommand(list, read, pathCmd)
	return cmd
}
