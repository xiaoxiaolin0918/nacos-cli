package output

// CapabilitiesSchemaVersion is the machine-readable capabilities contract version.
// Bump when envelope shape for Agents changes in a breaking way.
const CapabilitiesSchemaVersion = 1

// DefaultRedactFields lists JSON keys Agents / scrubbers should never echo.
var DefaultRedactFields = []string{
	"password", "passwd", "secret", "accessToken", "access_token",
	"token", "refresh_token", "refreshToken", "login.cred",
}

// AsWould normalizes one or more would-items into the envelope would[] array.
func AsWould(items ...map[string]any) []map[string]any {
	out := make([]map[string]any, 0, len(items))
	for _, m := range items {
		if m == nil {
			continue
		}
		out = append(out, m)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// FirstWould returns the first would item (for tests / simple consumers).
func FirstWould(would any) map[string]any {
	switch w := would.(type) {
	case []map[string]any:
		if len(w) > 0 {
			return w[0]
		}
	case []any:
		if len(w) > 0 {
			if m, ok := w[0].(map[string]any); ok {
				return m
			}
		}
	case map[string]any:
		return w
	}
	return nil
}
