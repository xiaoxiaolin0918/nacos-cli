package domain

import (
	"errors"
	"testing"
)

type typedFakeConfig struct {
	*fakeConfig
	typ       string
	typeCalls int
	lastType  string
	deleted   []string
}

func (f *typedFakeConfig) ConfigType(dataID, group, namespace string) (string, error) {
	f.typeCalls++
	return f.typ, nil
}

func (f *typedFakeConfig) ConfigPublish(dataID, group, content, namespace, typ string) (any, error) {
	f.lastType = typ
	return f.fakeConfig.ConfigPublish(dataID, group, content, namespace, typ)
}

func (f *typedFakeConfig) ConfigDelete(dataID, group, namespace string) (any, error) {
	f.deleted = append(f.deleted, dataID)
	delete(f.get, f.key(dataID, group, namespace))
	return true, nil
}

func TestRollbackPreservesConfigType(t *testing.T) {
	f := &typedFakeConfig{fakeConfig: &fakeConfig{get: map[string]string{"a.yaml|DEFAULT_GROUP|": "new"}}, typ: "yaml"}
	uc := ConfigUseCase{Port: f}
	if _, err := uc.Rollback("a.yaml", "DEFAULT_GROUP", "", "1"); err != nil {
		t.Fatal(err)
	}
	if f.lastType != "yaml" {
		t.Fatalf("rollback must keep current type, got %q", f.lastType)
	}
}

type mismatchHistory struct{ *fakeConfig }

func (m mismatchHistory) ConfigHistoryDetail(dataID, group, namespace, nid string) (map[string]any, error) {
	return map[string]any{"content": "other", "dataId": "other.yaml", "group": "DEFAULT_GROUP"}, nil
}

func TestRollbackRejectsHistoryOfOtherDataID(t *testing.T) {
	f := mismatchHistory{&fakeConfig{get: map[string]string{}}}
	uc := ConfigUseCase{Port: f}
	if _, err := uc.Rollback("a.yaml", "DEFAULT_GROUP", "", "9"); err == nil {
		t.Fatal("expected mismatch error")
	}
	if len(f.publish) != 0 {
		t.Fatal("must not publish on mismatch")
	}
}

func TestConfigDeleteRequiresExisting(t *testing.T) {
	f := &typedFakeConfig{fakeConfig: &fakeConfig{get: map[string]string{"a.yaml|DEFAULT_GROUP|": "x"}}}
	uc := ConfigUseCase{Port: f}
	if _, err := uc.Delete("missing.yaml", "DEFAULT_GROUP", ""); !errors.Is(err, ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
	if _, err := uc.Delete("a.yaml", "DEFAULT_GROUP", ""); err != nil {
		t.Fatal(err)
	}
	if len(f.deleted) != 1 || f.deleted[0] != "a.yaml" {
		t.Fatalf("%v", f.deleted)
	}
}

func TestConfigDeleteUnsupportedPort(t *testing.T) {
	uc := ConfigUseCase{Port: &fakeConfig{get: map[string]string{"a.yaml|DEFAULT_GROUP|": "x"}}}
	if _, err := uc.Delete("a.yaml", "DEFAULT_GROUP", ""); err == nil {
		t.Fatal("expected unsupported error")
	}
}

func TestBlurPattern(t *testing.T) {
	cases := map[string]string{"": "", "blade": "*blade*", "*blade": "*blade", "a*b": "a*b"}
	for in, want := range cases {
		if got := BlurPattern(in); got != want {
			t.Fatalf("BlurPattern(%q)=%q want %q", in, got, want)
		}
	}
}

type pagedPort struct {
	fakeExportPort
	pages map[int]any
	calls []int
}

func (p *pagedPort) ConfigList(dataID, group, namespace string, pageNo, pageSize int) (any, error) {
	p.calls = append(p.calls, pageNo)
	return p.pages[pageNo], nil
}

func TestCollectAllPages(t *testing.T) {
	p := &pagedPort{pages: map[int]any{
		1: map[string]any{"pagesAvailable": float64(2), "pageItems": []any{map[string]any{"dataId": "a", "group": "G1", "content": "1"}}},
		2: map[string]any{"pagesAvailable": float64(2), "pageItems": []any{map[string]any{"dataId": "b", "group": "G2", "content": "2"}}},
	}}
	uc := ConfigUseCase{Port: p}
	got, err := uc.CollectAllPages("", "", "ns", 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[1].Group != "G2" || got[0].Namespace != "ns" {
		t.Fatalf("%+v", got)
	}
	if len(p.calls) != 2 {
		t.Fatalf("calls %v", p.calls)
	}
}

func TestRedactUserPasswords(t *testing.T) {
	raw := map[string]any{"pageItems": []any{map[string]any{"username": "nacos", "password": "$2a$10$hash"}}}
	out := RedactUserPasswords(raw).(map[string]any)
	row := out["pageItems"].([]any)[0].(map[string]any)
	if _, ok := row["password"]; ok {
		t.Fatalf("password hash must be removed: %v", row)
	}
	if row["username"] != "nacos" {
		t.Fatal(row)
	}
}
