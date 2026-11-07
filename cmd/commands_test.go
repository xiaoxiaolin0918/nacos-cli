package cmd

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestCommandsHelpJSON(t *testing.T) {
	root := NewRoot()
	stdout := &strings.Builder{}
	stderr := &strings.Builder{}
	root.SetOut(stdout)
	root.SetErr(stderr)
	root.SetArgs([]string{"commands"})
	if err := root.Execute(); err != nil {
		t.Fatalf("%v stderr=%s", err, stderr.String())
	}
	var env map[string]any
	if err := json.Unmarshal([]byte(stdout.String()), &env); err != nil {
		t.Fatal(err, stdout.String())
	}
	if env["ok"] != true {
		t.Fatalf("%v", env)
	}
	res, _ := env["result"].(map[string]any)
	cmds, _ := res["commands"].([]any)
	if len(cmds) < 5 {
		t.Fatalf("too few commands: %v", cmds)
	}
	foundUpdate := false
	for _, c := range cmds {
		m, _ := c.(map[string]any)
		if m["path"] == "nacos update" || strings.HasSuffix(str(m["path"]), "update") {
			foundUpdate = true
			if m["needs_yes"] != true {
				t.Fatalf("update should need yes: %v", m)
			}
		}
	}
	if !foundUpdate {
		t.Fatalf("update missing: %v", cmds)
	}
}

func TestHelpJSONFlag(t *testing.T) {
	root := NewRoot()
	stdout := &strings.Builder{}
	stderr := &strings.Builder{}
	root.SetOut(stdout)
	root.SetErr(stderr)
	root.SetArgs([]string{"--help-json"})
	err := root.Execute()
	out := stdout.String()
	if err != nil {
		if _, ok := err.(SilentSuccess); !ok {
			if pe, ok2 := err.(*PrintedError); !ok2 || pe.ExitCode() != 0 {
				t.Fatalf("err=%v (%T) out=%s stderr=%s", err, err, out, stderr.String())
			}
		}
	}
	if !strings.Contains(out, `"envelope":"ok-json"`) && !strings.Contains(out, `"ok":true`) {
		t.Fatalf("out=%s stderr=%s err=%v", out, stderr.String(), err)
	}
}

func str(v any) string {
	s, _ := v.(string)
	return s
}
