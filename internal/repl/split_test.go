package repl

import (
	"reflect"
	"testing"
)

func TestSplitArgs(t *testing.T) {
	cases := []struct {
		in   string
		want []string
	}{
		{"", nil},
		{"  namespace list  ", []string{"namespace", "list"}},
		{`config get --data-id "a b.yaml"`, []string{"config", "get", "--data-id", "a b.yaml"}},
		{`config publish --content 'k: v' --dry-run`, []string{"config", "publish", "--content", "k: v", "--dry-run"}},
		{`echo "say \"hi\""`, []string{"echo", `say "hi"`}},
	}
	for _, c := range cases {
		got, err := SplitArgs(c.in)
		if err != nil {
			t.Fatalf("%q: %v", c.in, err)
		}
		if !reflect.DeepEqual(got, c.want) {
			t.Fatalf("%q: got %#v want %#v", c.in, got, c.want)
		}
	}
}

func TestSplitArgsErrors(t *testing.T) {
	if _, err := SplitArgs(`"unterminated`); err == nil {
		t.Fatal("expected error")
	}
	if _, err := SplitArgs(`trailing\`); err == nil {
		t.Fatal("expected escape error")
	}
}

func TestIsExitHelp(t *testing.T) {
	if !IsExit([]string{"exit"}) || !IsExit([]string{"QUIT"}) || !IsExit([]string{"q"}) {
		t.Fatal("exit")
	}
	if IsExit([]string{"namespace"}) {
		t.Fatal("not exit")
	}
	if !IsHelp([]string{"help"}) || !IsHelp([]string{"?"}) {
		t.Fatal("help")
	}
	if IsHelp([]string{"config", "--help"}) {
		t.Fatal("subcommand help is not bare help")
	}
}
