package domain

import "testing"

func TestNormalizeNamespace(t *testing.T) {
	cases := map[string]string{
		"":         "",
		"  ":       "",
		"public":   "",
		"Public":   "",
		"PUBLIC":   "",
		" public ": "",
		"ns1":      "ns1",
		"prod":     "prod",
		"public-x": "public-x",
	}
	for in, want := range cases {
		if got := NormalizeNamespace(in); got != want {
			t.Fatalf("NormalizeNamespace(%q)=%q want %q", in, got, want)
		}
	}
}

func TestEffectiveNamespace(t *testing.T) {
	if got := EffectiveNamespace("", "public"); got != "" {
		t.Fatalf("fallback public -> empty, got %q", got)
	}
	if got := EffectiveNamespace("public", "ns1"); got != "" {
		t.Fatalf("explicit public wins and normalizes, got %q", got)
	}
	if got := EffectiveNamespace("ns1", "public"); got != "ns1" {
		t.Fatalf("got %q", got)
	}
	if got := EffectiveNamespace("", "ns1"); got != "ns1" {
		t.Fatalf("got %q", got)
	}
	if got := EffectiveNamespace("", ""); got != "" {
		t.Fatalf("got %q", got)
	}
}