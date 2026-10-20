// Package domain ACL use cases (user / role / permission) via Nacos Open API.
package domain

import (
	"errors"
	"strings"
)

// AuthAdminPort is the infrastructure seam for Nacos RBAC admin APIs (v1).
type AuthAdminPort interface {
	UserList(username string, pageNo, pageSize int) (any, error)
	UserCreate(username, password string) (any, error)
	UserDelete(username string) (any, error)
	UserUpdatePassword(username, newPassword string) (any, error)

	RoleList(username, role string, pageNo, pageSize int) (any, error)
	RoleBind(role, username string) (any, error)
	RoleUnbind(role, username string) (any, error)

	PermissionList(role string, pageNo, pageSize int) (any, error)
	PermissionGrant(role, resource, action string) (any, error)
	PermissionRevoke(role, resource, action string) (any, error)
}

type UserUseCase struct{ Port AuthAdminPort }
type RoleUseCase struct{ Port AuthAdminPort }
type PermissionUseCase struct{ Port AuthAdminPort }

func (uc UserUseCase) List(username string, pageNo, pageSize int) (any, error) {
	if pageNo <= 0 {
		pageNo = 1
	}
	if pageSize <= 0 {
		pageSize = 100
	}
	return uc.Port.UserList(username, pageNo, pageSize)
}

func (uc UserUseCase) Create(username, password string) (any, error) {
	if strings.TrimSpace(username) == "" {
		return nil, errors.New("username is required")
	}
	if password == "" {
		return nil, errors.New("password is required")
	}
	return uc.Port.UserCreate(username, password)
}

func (uc UserUseCase) Delete(username string) (any, error) {
	if strings.TrimSpace(username) == "" {
		return nil, errors.New("username is required")
	}
	return uc.Port.UserDelete(username)
}

func (uc UserUseCase) UpdatePassword(username, newPassword string) (any, error) {
	if strings.TrimSpace(username) == "" {
		return nil, errors.New("username is required")
	}
	if newPassword == "" {
		return nil, errors.New("new password is required")
	}
	return uc.Port.UserUpdatePassword(username, newPassword)
}

func (uc RoleUseCase) List(username, role string, pageNo, pageSize int) (any, error) {
	if pageNo <= 0 {
		pageNo = 1
	}
	if pageSize <= 0 {
		pageSize = 100
	}
	return uc.Port.RoleList(username, role, pageNo, pageSize)
}

func (uc RoleUseCase) Bind(role, username string) (any, error) {
	if strings.TrimSpace(role) == "" {
		return nil, errors.New("role is required")
	}
	if strings.TrimSpace(username) == "" {
		return nil, errors.New("username is required")
	}
	return uc.Port.RoleBind(role, username)
}

func (uc RoleUseCase) Unbind(role, username string) (any, error) {
	if strings.TrimSpace(role) == "" {
		return nil, errors.New("role is required")
	}
	if strings.TrimSpace(username) == "" {
		return nil, errors.New("username is required")
	}
	return uc.Port.RoleUnbind(role, username)
}

func (uc PermissionUseCase) List(role string, pageNo, pageSize int) (any, error) {
	if pageNo <= 0 {
		pageNo = 1
	}
	if pageSize <= 0 {
		pageSize = 100
	}
	return uc.Port.PermissionList(role, pageNo, pageSize)
}

func (uc PermissionUseCase) Grant(role, resource, action string) (any, error) {
	if strings.TrimSpace(role) == "" {
		return nil, errors.New("role is required")
	}
	if strings.TrimSpace(resource) == "" {
		return nil, errors.New("resource is required")
	}
	action = strings.TrimSpace(strings.ToLower(action))
	if action == "" {
		return nil, errors.New("action is required (r|w|rw)")
	}
	switch action {
	case "r", "w", "rw", "read", "write":
		// normalize
		if action == "read" {
			action = "r"
		}
		if action == "write" {
			action = "w"
		}
	default:
		return nil, errors.New("action must be r, w, or rw")
	}
	return uc.Port.PermissionGrant(role, resource, action)
}

func (uc PermissionUseCase) Revoke(role, resource, action string) (any, error) {
	if strings.TrimSpace(role) == "" {
		return nil, errors.New("role is required")
	}
	if strings.TrimSpace(resource) == "" {
		return nil, errors.New("resource is required")
	}
	action = strings.TrimSpace(strings.ToLower(action))
	if action == "" {
		return nil, errors.New("action is required (r|w|rw)")
	}
	if action == "read" {
		action = "r"
	}
	if action == "write" {
		action = "w"
	}
	switch action {
	case "r", "w", "rw":
	default:
		return nil, errors.New("action must be r, w, or rw")
	}
	return uc.Port.PermissionRevoke(role, resource, action)
}
