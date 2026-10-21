package domain

import (
	"errors"
	"fmt"
	"strings"
)

// ConfigEntry is a listed config identity (+ optional inline content).
type ConfigEntry struct {
	DataID    string
	Group     string
	Namespace string
	Content   string
}

// ParseConfigPage extracts pageItems from a Nacos config list payload.
func ParseConfigPage(raw any) []ConfigEntry {
	m, ok := raw.(map[string]any)
	if !ok {
		return nil
	}
	items, ok := m["pageItems"].([]any)
	if !ok {
		return nil
	}
	out := make([]ConfigEntry, 0, len(items))
	for _, it := range items {
		row, ok := it.(map[string]any)
		if !ok {
			continue
		}
		e := ConfigEntry{
			DataID:  asString(row["dataId"]),
			Group:   asString(row["group"]),
			Content: asString(row["content"]),
		}
		if e.Group == "" {
			e.Group = "DEFAULT_GROUP"
		}
		if e.DataID == "" {
			continue
		}
		out = append(out, e)
	}
	return out
}

func asString(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case fmt.Stringer:
		return t.String()
	default:
		return ""
	}
}

// ExportOne loads one config body.
func (uc ConfigUseCase) ExportOne(dataID, group, namespace string) (string, error) {
	return uc.Get(dataID, group, namespace)
}

// CollectForBackup lists matching configs and ensures each has content (Get when missing).
func (uc ConfigUseCase) CollectForBackup(dataID, group, namespace string, page, pageSize int) ([]ConfigEntry, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 100
	}
	raw, err := uc.List(dataID, group, namespace, page, pageSize)
	if err != nil {
		return nil, err
	}
	entries := ParseConfigPage(raw)
	if len(entries) == 0 && strings.TrimSpace(dataID) != "" {
		// Accurate single export when list blur returns empty but get may work.
		body, err := uc.Get(dataID, group, namespace)
		if err != nil {
			return nil, err
		}
		g := group
		if g == "" {
			g = "DEFAULT_GROUP"
		}
		return []ConfigEntry{{DataID: dataID, Group: g, Namespace: namespace, Content: body}}, nil
	}
	for i := range entries {
		entries[i].Namespace = namespace
		if entries[i].Content != "" {
			continue
		}
		body, err := uc.Get(entries[i].DataID, entries[i].Group, namespace)
		if err != nil {
			return nil, fmt.Errorf("get %s/%s: %w", entries[i].Group, entries[i].DataID, err)
		}
		entries[i].Content = body
	}
	return entries, nil
}

// CollectAllPages walks every list page (blur filters) and fills content.
func (uc ConfigUseCase) CollectAllPages(dataID, group, namespace string, pageSize int) ([]ConfigEntry, error) {
	if pageSize <= 0 {
		pageSize = 100
	}
	var out []ConfigEntry
	for page := 1; page <= 10000; page++ {
		raw, err := uc.List(dataID, group, namespace, page, pageSize)
		if err != nil {
			return nil, err
		}
		entries := ParseConfigPage(raw)
		for i := range entries {
			entries[i].Namespace = namespace
			if entries[i].Content == "" {
				body, err := uc.Get(entries[i].DataID, entries[i].Group, namespace)
				if err != nil {
					return nil, fmt.Errorf("get %s/%s: %w", entries[i].Group, entries[i].DataID, err)
				}
				entries[i].Content = body
			}
		}
		out = append(out, entries...)
		pages := 0
		if m, ok := raw.(map[string]any); ok {
			if v, ok := m["pagesAvailable"].(float64); ok {
				pages = int(v)
			}
		}
		if len(entries) == 0 || page >= pages {
			break
		}
	}
	if out == nil {
		out = []ConfigEntry{}
	}
	return out, nil
}

// RedactUserPasswords drops password hashes from a Nacos user list payload.
func RedactUserPasswords(raw any) any {
	m, ok := raw.(map[string]any)
	if !ok {
		return raw
	}
	items, ok := m["pageItems"].([]any)
	if !ok {
		return raw
	}
	for _, it := range items {
		if row, ok := it.(map[string]any); ok {
			delete(row, "password")
		}
	}
	return m
}

// ValidateExportArgs requires data-id or --all semantics (empty dataID filter means all blur).
func ValidateExportArgs(dataID string, all bool) error {
	if all {
		return nil
	}
	if strings.TrimSpace(dataID) == "" {
		return errors.New("export needs --data-id or --all")
	}
	return nil
}
