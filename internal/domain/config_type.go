package domain

import (
	"fmt"
	"path"
	"strings"
)

// InferConfigTypeFromDataID maps a dataId file extension to a Nacos config type.
// Returns ("", false) when the extension is unknown or missing.
func InferConfigTypeFromDataID(dataID string) (string, bool) {
	base := path.Base(strings.TrimSpace(dataID))
	ext := strings.ToLower(path.Ext(base))
	switch ext {
	case ".yaml", ".yml":
		return "yaml", true
	case ".json":
		return "json", true
	case ".properties":
		return "properties", true
	case ".toml":
		return "toml", true
	case ".xml":
		return "xml", true
	case ".html", ".htm":
		return "html", true
	case ".txt":
		return "text", true
	default:
		return "", false
	}
}

// ResolvePublishType picks the config type for publish.
// When typeFlagSet is true, typeFlag is used as-is (must be non-empty).
// Otherwise the type is inferred from dataId; unknown extensions require an explicit --type.
func ResolvePublishType(dataID, typeFlag string, typeFlagSet bool) (string, error) {
	if typeFlagSet {
		t := strings.TrimSpace(typeFlag)
		if t == "" {
			return "", fmt.Errorf("--type must not be empty")
		}
		return t, nil
	}
	if inferred, ok := InferConfigTypeFromDataID(dataID); ok {
		return inferred, nil
	}
	return "", fmt.Errorf("cannot infer config type from dataId %q; pass --type explicitly (yaml|json|properties|toml|text|...)", dataID)
}