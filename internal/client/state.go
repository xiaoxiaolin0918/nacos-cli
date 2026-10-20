package client

import (
	"context"
	"encoding/json"
)

// ServerState probes console server state (version etc.). Best-effort; empty on failure.
func (c *Client) ServerState(ctx context.Context) map[string]any {
	b, status, _, err := c.do(ctx, "GET", "/v1/console/server/state", nil, nil)
	out := map[string]any{"http_status": status}
	if err != nil {
		out["ok"] = false
		out["error"] = err.Error()
		return out
	}
	if status >= 400 {
		out["ok"] = false
		out["error"] = truncate(string(b), 200)
		return out
	}
	var v any
	if json.Valid(b) {
		_ = json.Unmarshal(b, &v)
		out["body"] = v
		if m, ok := v.(map[string]any); ok {
			if ver, ok := m["version"].(string); ok {
				out["version"] = ver
			}
		}
	} else {
		out["body"] = string(b)
	}
	out["ok"] = true
	return out
}
