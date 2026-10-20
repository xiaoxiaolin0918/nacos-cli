// Package output prints the JSON ok-envelope used by every command.
package output

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

type Envelope struct {
	OK           bool              `json:"ok"`
	DryRun       bool              `json:"dry_run,omitempty"`
	Risk         string            `json:"risk,omitempty"`
	Error        any               `json:"error,omitempty"`
	Hint         string            `json:"hint,omitempty"`
	Method       string            `json:"method,omitempty"`
	URL          string            `json:"url,omitempty"`
	Would        map[string]any    `json:"would,omitempty"`
	WouldRequire []string          `json:"would_require,omitempty"`
	Result       json.RawMessage   `json:"result,omitempty"`
}

type ErrorBody struct {
	Type    string `json:"type"`
	Subtype string `json:"subtype,omitempty"`
	Message string `json:"message"`
	Hint    string `json:"hint,omitempty"`
	Risk    string `json:"risk,omitempty"`
	Action  string `json:"action,omitempty"`
}

func ErrorMessage(v any) string {
	switch e := v.(type) {
	case nil:
		return ""
	case string:
		return e
	case ErrorBody:
		return e.Message
	case *ErrorBody:
		if e == nil {
			return ""
		}
		return e.Message
	default:
		return fmt.Sprint(e)
	}
}

func Write(w io.Writer, format string, env Envelope, secrets []string) error {
	raw, err := json.Marshal(env)
	if err != nil {
		return err
	}
	if format == "pretty" {
		var buf bytes.Buffer
		if err := json.Indent(&buf, raw, "", "  "); err == nil {
			raw = buf.Bytes()
		}
	}
	raw = Scrub(raw, secrets)
	if len(raw) == 0 || raw[len(raw)-1] != '\n' {
		raw = append(raw, '\n')
	}
	_, err = w.Write(raw)
	return err
}

func Scrub(b []byte, secrets []string) []byte {
	s := string(b)
	seen := map[string]struct{}{}
	for _, sec := range secrets {
		if len(sec) < 4 {
			continue
		}
		if _, ok := seen[sec]; ok {
			continue
		}
		seen[sec] = struct{}{}
		s = strings.ReplaceAll(s, sec, "***")
	}
	return []byte(s)
}

func RawResult(body []byte) json.RawMessage {
	body = bytes.TrimSpace(body)
	if len(body) == 0 {
		return nil
	}
	if json.Valid(body) {
		return json.RawMessage(body)
	}
	b, err := json.Marshal(string(body))
	if err != nil {
		return nil
	}
	return b
}

func MustResult(v any) json.RawMessage {
	b, err := json.Marshal(v)
	if err != nil {
		return json.RawMessage(`null`)
	}
	return b
}
