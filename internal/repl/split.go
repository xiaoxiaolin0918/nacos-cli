// Package repl helpers for the interactive shell.
package repl

import (
	"fmt"
	"strings"
	"unicode"
)

// SplitArgs tokenizes a REPL line like a minimal shell (supports "double" and 'single' quotes).
func SplitArgs(line string) ([]string, error) {
	line = strings.TrimSpace(line)
	if line == "" {
		return nil, nil
	}
	var (
		out    []string
		cur    strings.Builder
		quote  rune // 0, '"' or '\''
		escape bool
	)
	for _, r := range line {
		if escape {
			cur.WriteRune(r)
			escape = false
			continue
		}
		if quote != 0 {
			if r == '\\' && quote == '"' {
				escape = true
				continue
			}
			if r == quote {
				quote = 0
				continue
			}
			cur.WriteRune(r)
			continue
		}
		switch {
		case r == '\\':
			escape = true
		case r == '"' || r == '\'':
			quote = r
		case unicode.IsSpace(r):
			if cur.Len() > 0 {
				out = append(out, cur.String())
				cur.Reset()
			}
		default:
			cur.WriteRune(r)
		}
	}
	if escape {
		return nil, fmt.Errorf("unterminated escape")
	}
	if quote != 0 {
		return nil, fmt.Errorf("unterminated quote")
	}
	if cur.Len() > 0 {
		out = append(out, cur.String())
	}
	return out, nil
}

// IsExit reports whether the first token means leave the REPL.
func IsExit(args []string) bool {
	if len(args) == 0 {
		return false
	}
	switch strings.ToLower(args[0]) {
	case "exit", "quit", "q":
		return true
	default:
		return false
	}
}

// IsHelp reports a bare help / ? request (not "cmd --help").
func IsHelp(args []string) bool {
	if len(args) == 0 {
		return false
	}
	if len(args) == 1 {
		switch strings.ToLower(args[0]) {
		case "help", "?":
			return true
		}
	}
	return false
}
