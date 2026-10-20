// Package risk classifies commands and enforces the high-risk gate.
package risk

import (
	"fmt"
	"strings"
)

const (
	Read = "read"
	Write = "write"
	High = "high-risk-write"
	ExitConfirmationRequired = 10
)

type Policy struct {
	Level   string
	NeedYes bool
}

type GateResult struct {
	Action  string
	Risk    string
	Missing []string
	Hint    string
}

func (g GateResult) Error() string {
	return fmt.Sprintf("refusing %s without %s", g.Risk, join(g.Missing))
}

func IsConfirmation(err error) bool {
	_, ok := err.(GateResult)
	return ok
}

func ReadOnly() Policy { return Policy{Level: Read} }
func Mutation() Policy { return Policy{Level: Write} }
func HighRisk() Policy { return Policy{Level: High, NeedYes: true} }

func (p Policy) Missing(yes bool) []string {
	if p.Level != High {
		return nil
	}
	if p.NeedYes && !yes {
		return []string{"--yes"}
	}
	return nil
}

func (p Policy) Check(yes bool, action string) error {
	missing := p.Missing(yes)
	if len(missing) == 0 {
		return nil
	}
	return GateResult{
		Action:  action,
		Risk:    p.Level,
		Missing: missing,
		Hint:    "add --yes to confirm; or use --dry-run to preview",
	}
}

func join(parts []string) string {
	if len(parts) == 0 {
		return ""
	}
	if len(parts) == 1 {
		return parts[0]
	}
	return strings.Join(parts[:len(parts)-1], ", ") + " and " + parts[len(parts)-1]
}
