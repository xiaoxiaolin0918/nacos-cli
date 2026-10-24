package domain

import "testing"

func TestInferConfigTypeFromDataID(t *testing.T) {
	cases := map[string]string{
		"app.yaml":         "yaml",
		"app.yml":          "yaml",
		"APP.YAML":         "yaml",
		"cfg.json":         "json",
		"a.properties":     "properties",
		"x.toml":           "toml",
		"page.xml":         "xml",
		"note.txt":         "text",
		"noext":            "",
		"weird.dat":        "",
		"path/to/app.yaml": "yaml",
	}
	for in, want := range cases {
		got, ok := InferConfigTypeFromDataID(in)
		if want == "" {
			if ok {
				t.Fatalf("%q inferred %q, want unknown", in, got)
			}
			continue
		}
		if !ok || got != want {
			t.Fatalf("%q -> %q ok=%v want %q", in, got, ok, want)
		}
	}
}

func TestResolvePublishType(t *testing.T) {
	got, err := ResolvePublishType("app.yaml", "", false)
	if err != nil || got != "yaml" {
		t.Fatalf("%q %v", got, err)
	}
	got, err = ResolvePublishType("app.yaml", "json", true)
	if err != nil || got != "json" {
		t.Fatalf("explicit override: %q %v", got, err)
	}
	_, err = ResolvePublishType("app.dat", "", false)
	if err == nil {
		t.Fatal("unknown ext without --type must error")
	}
	_, err = ResolvePublishType("app.dat", "", true)
	if err == nil {
		t.Fatal("empty --type must error")
	}
	got, err = ResolvePublishType("app.dat", "text", true)
	if err != nil || got != "text" {
		t.Fatalf("%q %v", got, err)
	}
}