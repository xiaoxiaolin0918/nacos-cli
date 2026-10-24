package config

import (
	"fmt"
	"regexp"
	"strings"
)

var drivePathRE = regexp.MustCompile(`(?i)^[a-z]:[/\\]`)

// ValidateContextPath rejects values that are not a URL path — notably Git Bash
// (MSYS) path conversion of "/nacos" into something like "D:/Program Files/Git/nacos".
// Empty is allowed (caller uses the default /nacos).
// Under Git Bash, set MSYS_NO_PATHCONV=1 when passing --context-path /nacos.
func ValidateContextPath(p string) error {
	p = strings.TrimSpace(p)
	if p == "" {
		return nil
	}
	if drivePathRE.MatchString(p) || strings.Contains(p, `\`) || strings.Contains(p, ":") {
		return fmt.Errorf("invalid context-path %q: looks like a filesystem path (under Git Bash, run with MSYS_NO_PATHCONV=1 --context-path /nacos)", p)
	}
	if !strings.HasPrefix(p, "/") {
		return fmt.Errorf("invalid context-path %q: must start with /", p)
	}
	if strings.Contains(p, " ") {
		return fmt.Errorf("invalid context-path %q: must not contain spaces", p)
	}
	return nil
}
