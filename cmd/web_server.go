package cmd

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/xiaoxiaolin0918/nacos-cli/internal/audit"
	"github.com/xiaoxiaolin0918/nacos-cli/internal/config"
)

type webServer struct {
	dir      string
	audit    *audit.Store
	credPath string // optional; empty -> config.CredentialsPath()
}

func newWebServer(dir string, store *audit.Store) *webServer {
	return &webServer{dir: dir, audit: store}
}

func (s *webServer) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleIndex)
	mux.HandleFunc("/api/profiles", s.handleProfiles)
	mux.HandleFunc("/api/profiles/", s.handleProfileOne)
	mux.HandleFunc("/api/audit", s.handleAudit)
	return mux
}

func (s *webServer) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	lang := webLang(r.Header.Get("Accept-Language"))
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	_, _ = io.WriteString(w, webConsoleHTML(lang))
}

func (s *webServer) handleProfiles(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	names, err := config.ListProfilesIn(s.dir)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": err.Error()})
		return
	}
	items := make([]map[string]any, 0, len(names))
	for _, name := range names {
		p, ok, err := config.LoadProfileIn(s.dir, name)
		if err != nil || !ok {
			continue
		}
		meta, _, _ := config.LoadMetaIn(s.dir, name)
		items = append(items, map[string]any{
			"name":    name,
			"profile": redactedProfileView(s.dir, name, p),
			"meta":    meta,
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "profiles": items, "profile_dir": s.dir})
}

func (s *webServer) handleProfileOne(w http.ResponseWriter, r *http.Request) {
	rest := strings.TrimPrefix(r.URL.Path, "/api/profiles/")
	rest = strings.Trim(rest, "/")
	if rest == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "invalid profile name"})
		return
	}
	parts := strings.Split(rest, "/")
	name := parts[0]
	if name == "" || strings.Contains(name, "..") {
		writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "invalid profile name"})
		return
	}
	if len(parts) == 2 && parts[1] == "check-auth" {
		s.handleCheckAuth(w, r, name)
		return
	}
	if len(parts) != 1 {
		writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "invalid profile name"})
		return
	}
	switch r.Method {
	case http.MethodGet:
		s.getProfile(w, name)
	case http.MethodPut:
		s.putProfile(w, r, name)
	case http.MethodDelete:
		s.deleteProfile(w, r, name)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *webServer) getProfile(w http.ResponseWriter, name string) {
	p, ok, err := config.LoadProfileIn(s.dir, name)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": err.Error()})
		return
	}
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]any{"ok": false, "error": "not found"})
		return
	}
	meta, _, _ := config.LoadMetaIn(s.dir, name)
	writeJSON(w, http.StatusOK, map[string]any{
		"ok": true, "name": name, "profile": redactedProfileView(s.dir, name, p), "meta": meta,
	})
}

func (s *webServer) putProfile(w http.ResponseWriter, r *http.Request, name string) {
	var req profileWriteRequest
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "invalid json: " + err.Error()})
		return
	}
	existing, _, err := config.LoadProfileIn(s.dir, name)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": err.Error()})
		return
	}
	hydrateProfilePassword(s.dir, name, &existing)
	p, err := mergeProfileWrite(existing, req)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": err.Error()})
		return
	}
	if strings.TrimSpace(p.Server) == "" && strings.TrimSpace(p.Host) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "server or host required"})
		return
	}
	if err := persistProfilePassword(s.dir, name, &p, req.ClearPassword); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": err.Error()})
		return
	}
	if _, err := config.SaveProfileIn(s.dir, name, p); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": err.Error()})
		return
	}
	meta := req.Meta
	if _, err := config.SaveMetaIn(s.dir, name, meta); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": err.Error()})
		return
	}
	s.note("web.profile.save", "write", name, true, "")
	got, _, _ := config.LoadProfileIn(s.dir, name)
	gotMeta, _, _ := config.LoadMetaIn(s.dir, name)
	writeJSON(w, http.StatusOK, map[string]any{
		"ok": true, "name": name, "profile": redactedProfileView(s.dir, name, got), "meta": gotMeta,
	})
}

func (s *webServer) deleteProfile(w http.ResponseWriter, r *http.Request, name string) {
	if r.URL.Query().Get("confirm") != "1" {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"ok": false, "error": "confirm=1 required to delete",
		})
		return
	}
	if err := config.DeleteProfileIn(s.dir, name); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": err.Error()})
		return
	}
	_ = persistProfilePassword(s.dir, name, &config.Profile{}, true)
	_ = config.DeleteMetaIn(s.dir, name)
	s.note("web.profile.delete", "high-risk-write", name, true, "")
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "deleted": name})
}

func (s *webServer) handleAudit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	f := audit.Filter{
		CommandContains: r.URL.Query().Get("command"),
		Profile:         r.URL.Query().Get("profile"),
		Limit:           500,
	}
	if since := strings.TrimSpace(r.URL.Query().Get("since")); since != "" {
		ts, err := time.Parse(time.RFC3339, since)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "since must be RFC3339"})
			return
		}
		f.Since = &ts
	}
	entries, err := s.audit.Query(f)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": err.Error()})
		return
	}
	// reverse for newest first in UI
	for i, j := 0, len(entries)-1; i < j; i, j = i+1, j-1 {
		entries[i], entries[j] = entries[j], entries[i]
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "entries": entries})
}

func (s *webServer) note(command, risk, profile string, ok bool, errMsg string) {
	if s.audit == nil {
		return
	}
	_ = s.audit.Append(audit.Entry{
		Time:    time.Now().UTC().Format(time.RFC3339),
		Command: command,
		Risk:    risk,
		Profile: profile,
		OK:      ok,
		Error:   errMsg,
	})
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(code)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(true)
	_ = enc.Encode(v)
}

