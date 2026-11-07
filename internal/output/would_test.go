package output

import (
	"encoding/json"
	"testing"
)

func TestAsWouldAndEnvelope(t *testing.T) {
	env := Envelope{
		OK:     true,
		DryRun: true,
		Would:  AsWould(map[string]any{"action": "x"}),
	}
	b, err := json.Marshal(env)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	arr, ok := m["would"].([]any)
	if !ok || len(arr) != 1 {
		t.Fatalf("%s", b)
	}
	if CapabilitiesSchemaVersion < 1 {
		t.Fatal("schema version")
	}
	if len(DefaultRedactFields) < 3 {
		t.Fatal("redact fields")
	}
}
