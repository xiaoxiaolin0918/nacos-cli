package domain

import "strings"

// NormalizeNamespace maps Nacos display-name aliases to the real namespace ID.
// The built-in public namespace has ID "" and show name "public"; treating the
// display name as a tenant ID creates a phantom namespace.
func NormalizeNamespace(ns string) string {
	ns = strings.TrimSpace(ns)
	if strings.EqualFold(ns, "public") {
		return ""
	}
	return ns
}

// EffectiveNamespace prefers an explicit (flag) value when non-empty after trim,
// otherwise falls back to the default (profile/env). Both sides are normalized.
func EffectiveNamespace(explicit, fallback string) string {
	explicit = strings.TrimSpace(explicit)
	if explicit != "" {
		return NormalizeNamespace(explicit)
	}
	return NormalizeNamespace(fallback)
}