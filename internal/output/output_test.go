package output

import (
	"bytes"
	"strings"
	"testing"
)

func TestWriteJSONAndPretty(t *testing.T) {
	env := Envelope{OK: true, Result: MustResult(map[string]any{"n": 1})}
	var buf bytes.Buffer
	if err := Write(&buf, "json", env, nil); err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(buf.String(), `{"ok":true`) {
		t.Fatalf("%s", buf.String())
	}
	buf.Reset()
	if err := Write(&buf, "pretty", env, nil); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "\n  ") {
		t.Fatalf("%s", buf.String())
	}
}

func TestScrub(t *testing.T) {
	raw := Scrub([]byte(`{"t":"abcd-secret"}`), []string{"abcd-secret"})
	if strings.Contains(string(raw), "abcd-secret") {
		t.Fatal(string(raw))
	}
}
func TestErrorBodyCode(t *testing.T) {
	env := Envelope{OK: false, Error: ErrorBody{
		Type: "confirmation", Subtype: "confirmation_required",
		Code: CodeConfirmationRequired, Message: "need --yes",
	}}
	var buf bytes.Buffer
	if err := Write(&buf, "json", env, nil); err != nil {
		t.Fatal(err)
	}
	s := buf.String()
	if !strings.Contains(s, `"code":"confirmation_required"`) {
		t.Fatalf("%s", s)
	}
	if HTTPCode(401) != "http_401" || HTTPCode(403) != "http_403" {
		t.Fatalf("HTTPCode")
	}
}
