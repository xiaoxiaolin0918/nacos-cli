package cmd

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestReplExitAndHelpAndProfilePath(t *testing.T) {
	root := NewRoot()
	in := bytes.NewBufferString("help\nprofile path default\nexit\n")
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	root.SetIn(in)
	root.SetOut(out)
	root.SetErr(errOut)
	root.SetArgs([]string{"repl"})
	if err := root.Execute(); err != nil {
		t.Fatal(err, out.String(), errOut.String())
	}
	raw := out.String()
	if !strings.Contains(raw, "Built-ins") && !strings.Contains(raw, "help") {
		t.Fatalf("help missing: %s", raw)
	}
	if !strings.Contains(raw, `"ok"`) {
		t.Fatalf("expected profile path JSON: %s", raw)
	}
	// Last successful command should be parseable JSON somewhere
	if !strings.Contains(raw, "default.conf") && !strings.Contains(raw, "nacos-cli") {
		// path may vary; at least ok envelope
		idx := strings.LastIndex(raw, "{")
		if idx < 0 {
			t.Fatalf("%s", raw)
		}
		var env map[string]any
		if err := json.Unmarshal([]byte(raw[idx:]), &env); err != nil {
			t.Fatal(err, raw)
		}
	}
}

func TestReplQuitAlias(t *testing.T) {
	root := NewRoot()
	in := bytes.NewBufferString("quit\n")
	out := &bytes.Buffer{}
	root.SetIn(in)
	root.SetOut(out)
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"repl"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
}

func TestPersistentArgsInheritProfile(t *testing.T) {
	a := &App{flagProfile: "lab", profileName: "lab", format: "pretty"}
	args := a.persistentArgs()
	joined := strings.Join(args, " ")
	if !strings.Contains(joined, "--profile lab") {
		t.Fatalf("%v", args)
	}
	if !strings.Contains(joined, "--format pretty") {
		t.Fatalf("%v", args)
	}
}
