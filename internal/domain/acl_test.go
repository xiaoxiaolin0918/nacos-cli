package domain

import (
	"testing"
)

type fakeACL struct {
	users   []string
	binds   []string
	grants  []string
	revokes []string
	calls   []string
}

func (f *fakeACL) UserList(string, int, int) (any, error) {
	f.calls = append(f.calls, "user.list")
	return map[string]any{"pageItems": f.users}, nil
}
func (f *fakeACL) UserCreate(username, password string) (any, error) {
	f.calls = append(f.calls, "user.create:"+username)
	if password == "" {
		return nil, ErrNotFound
	}
	f.users = append(f.users, username)
	return "ok", nil
}
func (f *fakeACL) UserDelete(username string) (any, error) {
	f.calls = append(f.calls, "user.delete:"+username)
	return "ok", nil
}
func (f *fakeACL) UserUpdatePassword(username, newPassword string) (any, error) {
	f.calls = append(f.calls, "user.passwd:"+username)
	_ = newPassword
	return "ok", nil
}
func (f *fakeACL) RoleList(string, string, int, int) (any, error) {
	return map[string]any{"pageItems": f.binds}, nil
}
func (f *fakeACL) RoleBind(role, username string) (any, error) {
	f.calls = append(f.calls, "role.bind:"+role+":"+username)
	f.binds = append(f.binds, role+"|"+username)
	return "ok", nil
}
func (f *fakeACL) RoleUnbind(role, username string) (any, error) {
	f.calls = append(f.calls, "role.unbind:"+role+":"+username)
	return "ok", nil
}
func (f *fakeACL) PermissionList(string, int, int) (any, error) {
	return map[string]any{"pageItems": f.grants}, nil
}
func (f *fakeACL) PermissionGrant(role, resource, action string) (any, error) {
	f.calls = append(f.calls, "perm.grant:"+role+":"+resource+":"+action)
	f.grants = append(f.grants, role+"|"+resource+"|"+action)
	return "ok", nil
}
func (f *fakeACL) PermissionRevoke(role, resource, action string) (any, error) {
	f.calls = append(f.calls, "perm.revoke:"+role+":"+resource+":"+action)
	f.revokes = append(f.revokes, role+"|"+resource+"|"+action)
	return "ok", nil
}

func TestUserCreateRequiresFields(t *testing.T) {
	f := &fakeACL{}
	uc := UserUseCase{Port: f}
	if _, err := uc.Create("", "x"); err == nil {
		t.Fatal("expected username error")
	}
	if _, err := uc.Create("u1", ""); err == nil {
		t.Fatal("expected password error")
	}
	if _, err := uc.Create("u1", "secret"); err != nil {
		t.Fatal(err)
	}
	if len(f.calls) != 1 || f.calls[0] != "user.create:u1" {
		t.Fatal(f.calls)
	}
}

func TestRoleBindUnbind(t *testing.T) {
	f := &fakeACL{}
	uc := RoleUseCase{Port: f}
	if _, err := uc.Bind("", "u"); err == nil {
		t.Fatal("role required")
	}
	if _, err := uc.Bind("dev", "alice"); err != nil {
		t.Fatal(err)
	}
	if _, err := uc.Unbind("dev", "alice"); err != nil {
		t.Fatal(err)
	}
}

func TestPermissionActionNormalize(t *testing.T) {
	f := &fakeACL{}
	uc := PermissionUseCase{Port: f}
	if _, err := uc.Grant("dev", "*:DEFAULT_GROUP:*", "read"); err != nil {
		t.Fatal(err)
	}
	if f.calls[0] != "perm.grant:dev:*:DEFAULT_GROUP:*:r" {
		t.Fatal(f.calls)
	}
	if _, err := uc.Grant("dev", "x", "bad"); err == nil {
		t.Fatal("expected bad action")
	}
	if _, err := uc.Revoke("dev", "*:DEFAULT_GROUP:*", "rw"); err != nil {
		t.Fatal(err)
	}
}
