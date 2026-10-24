package config

import "strings"

// NormalizeNamespaceID maps display-name "public" to the built-in empty tenant ID.
// Kept in config (not domain) so profile load/merge can normalize without importing domain.
func NormalizeNamespaceID(ns string) string {
	ns = strings.TrimSpace(ns)
	if strings.EqualFold(ns, "public") {
		return ""
	}
	return ns
}