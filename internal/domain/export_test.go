package domain

import "testing"

type fakeExportPort struct {
	get  map[string]string
	list any
	err  error
}

func (f fakeExportPort) ConfigGet(dataID, group, namespace string) (string, error) {
	if f.err != nil {
		return "", f.err
	}
	key := group + "/" + dataID
	v, ok := f.get[key]
	if !ok {
		return "", ErrNotFound
	}
	return v, nil
}
func (f fakeExportPort) ConfigList(dataID, group, namespace string, pageNo, pageSize int) (any, error) {
	return f.list, f.err
}
func (f fakeExportPort) ConfigPublish(string, string, string, string, string) (any, error) {
	return nil, nil
}
func (f fakeExportPort) ConfigHistory(string, string, string, int, int) (any, error) { return nil, nil }
func (f fakeExportPort) ConfigHistoryDetail(string, string, string, string) (map[string]any, error) {
	return nil, nil
}

func TestParseConfigPage(t *testing.T) {
	raw := map[string]any{
		"pageItems": []any{
			map[string]any{"dataId": "a.yaml", "group": "G", "content": "x"},
			map[string]any{"dataId": "b.yaml"}, // default group
		},
	}
	got := ParseConfigPage(raw)
	if len(got) != 2 || got[0].Content != "x" || got[1].Group != "DEFAULT_GROUP" {
		t.Fatalf("%+v", got)
	}
}

func TestCollectForBackupFillsContent(t *testing.T) {
	uc := ConfigUseCase{Port: fakeExportPort{
		list: map[string]any{
			"pageItems": []any{
				map[string]any{"dataId": "a.yaml", "group": "DEFAULT_GROUP"},
			},
		},
		get: map[string]string{"DEFAULT_GROUP/a.yaml": "body"},
	}}
	items, err := uc.CollectForBackup("", "DEFAULT_GROUP", "", 1, 100)
	if err != nil || len(items) != 1 || items[0].Content != "body" {
		t.Fatalf("%v %+v", err, items)
	}
}

func TestValidateExportArgs(t *testing.T) {
	if ValidateExportArgs("", false) == nil {
		t.Fatal("expected error")
	}
	if ValidateExportArgs("", true) != nil {
		t.Fatal("all ok")
	}
	if ValidateExportArgs("x", false) != nil {
		t.Fatal("data-id ok")
	}
}
