package config

import "testing"

func TestValidateContextPathOK(t *testing.T) {
	for _, p := range []string{"", "/nacos", "/nacos/", "/api", "/"} {
		if err := ValidateContextPath(p); err != nil {
			t.Fatalf("%q: %v", p, err)
		}
	}
}

func TestValidateContextPathRejectsMSYS(t *testing.T) {
	bads := []string{
		`D:/Program Files/Git/nacos`,
		`D:\Program Files\Git\nacos`,
		`C:/nacos`,
		`/nacos with space`,
		`nacos`,
		`relative/nacos`,
	}
	for _, p := range bads {
		if err := ValidateContextPath(p); err == nil {
			t.Fatalf("expected reject for %q", p)
		}
	}
}