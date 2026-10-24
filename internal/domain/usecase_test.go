package domain

import (
	"errors"
	"strings"
	"testing"
)

type fakeConfig struct {
	get     map[string]string
	types   map[string]string
	publish []string
}

func (f *fakeConfig) key(dataID, group, ns string) string {
	return dataID + "|" + group + "|" + ns
}

func (f *fakeConfig) ConfigGet(dataID, group, namespace string) (string, error) {
	v, ok := f.get[f.key(dataID, group, namespace)]
	if !ok {
		return "", ErrNotFound
	}
	return v, nil
}
func (f *fakeConfig) ConfigList(string, string, string, int, int) (any, error) {
	return map[string]any{"pageItems": []any{}}, nil
}
func (f *fakeConfig) ConfigPublish(dataID, group, content, namespace, typ string) (any, error) {
	f.publish = append(f.publish, dataID+"|"+typ+"|"+namespace)
	f.get[f.key(dataID, group, namespace)] = content
	if f.types == nil {
		f.types = map[string]string{}
	}
	f.types[f.key(dataID, group, namespace)] = typ
	return true, nil
}
func (f *fakeConfig) ConfigHistory(string, string, string, int, int) (any, error) {
	return map[string]any{"pageItems": []any{}}, nil
}
func (f *fakeConfig) ConfigHistoryDetail(dataID, group, namespace, nid string) (map[string]any, error) {
	if nid == "1" {
		return map[string]any{"content": "old"}, nil
	}
	return nil, ErrNotFound
}
func (f *fakeConfig) ConfigType(dataID, group, namespace string) (string, error) {
	if f.types == nil {
		return "", nil
	}
	return f.types[f.key(dataID, group, namespace)], nil
}

func TestConfigPublishCreateOnly(t *testing.T) {
	f := &fakeConfig{get: map[string]string{}}
	uc := ConfigUseCase{Port: f}
	if _, err := uc.Publish("a.yaml", "DEFAULT_GROUP", "x", "", "yaml", true, false); err != nil {
		t.Fatal(err)
	}
	_, err := uc.Publish("a.yaml", "DEFAULT_GROUP", "y", "", "yaml", true, false)
	if !errors.Is(err, ErrExists) {
		t.Fatalf("%v", err)
	}
	if _, err := uc.Publish("a.yaml", "DEFAULT_GROUP", "y", "", "yaml", false, false); err != nil {
		t.Fatal(err)
	}
}

func TestConfigPublishNormalizesPublicNamespace(t *testing.T) {
	f := &fakeConfig{get: map[string]string{}}
	uc := ConfigUseCase{Port: f}
	if _, err := uc.Publish("a.yaml", "DEFAULT_GROUP", "x", "public", "yaml", true, false); err != nil {
		t.Fatal(err)
	}
	if len(f.publish) != 1 || f.publish[0] != "a.yaml|yaml|" {
		t.Fatalf("expected tenant empty after public normalize, got %v", f.publish)
	}
	got, err := uc.Get("a.yaml", "DEFAULT_GROUP", "Public")
	if err != nil || got != "x" {
		t.Fatalf("%q %v", got, err)
	}
}

func TestConfigPublishChangeTypeRequired(t *testing.T) {
	f := &fakeConfig{
		get:   map[string]string{"a.yaml|DEFAULT_GROUP|": "body"},
		types: map[string]string{"a.yaml|DEFAULT_GROUP|": "yaml"},
	}
	uc := ConfigUseCase{Port: f}
	_, err := uc.Publish("a.yaml", "DEFAULT_GROUP", "body2", "", "json", false, false)
	if err == nil || !strings.Contains(err.Error(), "--change-type") {
		t.Fatalf("expected change-type error, got %v", err)
	}
	if _, err := uc.Publish("a.yaml", "DEFAULT_GROUP", "body2", "", "json", false, true); err != nil {
		t.Fatal(err)
	}
}

func TestConfigRollback(t *testing.T) {
	f := &fakeConfig{get: map[string]string{}}
	uc := ConfigUseCase{Port: f}
	out, err := uc.Rollback("a.yaml", "DEFAULT_GROUP", "", "1")
	if err != nil || out != true {
		t.Fatalf("%v %v", out, err)
	}
	if f.get["a.yaml|DEFAULT_GROUP|"] != "old" {
		t.Fatal(f.get)
	}
}

type fakeSvc struct {
	calls []string
}

func (f *fakeSvc) ServiceList(string, int, int) (any, error) { return map[string]any{}, nil }
func (f *fakeSvc) InstanceList(string, string, string) (any, error) {
	return map[string]any{}, nil
}
func (f *fakeSvc) InstanceRegister(service, ip string, port int, namespace, group, cluster string, ephemeral bool) (any, error) {
	f.calls = append(f.calls, "reg")
	if !ephemeral {
		f.calls = append(f.calls, "persistent")
	}
	return "ok", nil
}
func (f *fakeSvc) InstanceDeregister(string, string, int, string, string, string, bool) (any, error) {
	f.calls = append(f.calls, "dereg")
	return "ok", nil
}
func (f *fakeSvc) InstanceEnabled(string, string, int, bool, string, string, string, bool) (any, error) {
	f.calls = append(f.calls, "enabled")
	return "ok", nil
}

func TestServiceDefaultsEphemeral(t *testing.T) {
	f := &fakeSvc{}
	uc := ServiceUseCase{Port: f}
	if _, err := uc.Register("svc", "127.0.0.1", 1, "", "", "", true); err != nil {
		t.Fatal(err)
	}
	if len(f.calls) != 1 || f.calls[0] != "reg" {
		t.Fatal(f.calls)
	}
	if _, err := uc.Register("", "127.0.0.1", 1, "", "", "", true); err == nil {
		t.Fatal("expected error")
	}
}