package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

func (c *Client) mutateJSON(method, path string, query, form url.Values) (any, error) {
	b, status, _, err := c.do(context.Background(), method, path, query, form)
	if err != nil {
		return nil, err
	}
	if status >= 400 {
		return nil, fmt.Errorf("HTTP %d: %s", status, truncate(string(b), 200))
	}
	if json.Valid(b) {
		var v any
		_ = json.Unmarshal(b, &v)
		return v, nil
	}
	return strings.TrimSpace(string(b)), nil
}

func (c *Client) UserList(username string, pageNo, pageSize int) (any, error) {
	q := url.Values{}
	q.Set("pageNo", strconv.Itoa(pageNo))
	q.Set("pageSize", strconv.Itoa(pageSize))
	q.Set("search", "blur")
	q.Set("username", username)
	return c.getJSON(context.Background(), "/v1/auth/users", q)
}

func (c *Client) UserCreate(username, password string) (any, error) {
	q := url.Values{}
	q.Set("username", username)
	q.Set("password", password)
	return c.mutateJSON(http.MethodPost, "/v1/auth/users", q, nil)
}

func (c *Client) UserDelete(username string) (any, error) {
	q := url.Values{}
	q.Set("username", username)
	return c.mutateJSON(http.MethodDelete, "/v1/auth/users", q, nil)
}

func (c *Client) UserUpdatePassword(username, newPassword string) (any, error) {
	q := url.Values{}
	q.Set("username", username)
	q.Set("newPassword", newPassword)
	return c.mutateJSON(http.MethodPut, "/v1/auth/users", q, nil)
}

func (c *Client) RoleList(username, role string, pageNo, pageSize int) (any, error) {
	q := url.Values{}
	q.Set("pageNo", strconv.Itoa(pageNo))
	q.Set("pageSize", strconv.Itoa(pageSize))
	q.Set("search", "blur")
	q.Set("username", username)
	if role != "" {
		q.Set("role", role)
	}
	return c.getJSON(context.Background(), "/v1/auth/roles", q)
}

func (c *Client) RoleBind(role, username string) (any, error) {
	q := url.Values{}
	q.Set("role", role)
	q.Set("username", username)
	return c.mutateJSON(http.MethodPost, "/v1/auth/roles", q, nil)
}

func (c *Client) RoleUnbind(role, username string) (any, error) {
	q := url.Values{}
	q.Set("role", role)
	q.Set("username", username)
	return c.mutateJSON(http.MethodDelete, "/v1/auth/roles", q, nil)
}

func (c *Client) PermissionList(role string, pageNo, pageSize int) (any, error) {
	q := url.Values{}
	q.Set("pageNo", strconv.Itoa(pageNo))
	q.Set("pageSize", strconv.Itoa(pageSize))
	q.Set("search", "accurate")
	q.Set("role", role)
	return c.getJSON(context.Background(), "/v1/auth/permissions", q)
}

func (c *Client) PermissionGrant(role, resource, action string) (any, error) {
	q := url.Values{}
	q.Set("role", role)
	q.Set("resource", resource)
	q.Set("action", action)
	return c.mutateJSON(http.MethodPost, "/v1/auth/permissions", q, nil)
}

func (c *Client) PermissionRevoke(role, resource, action string) (any, error) {
	q := url.Values{}
	q.Set("role", role)
	q.Set("resource", resource)
	q.Set("action", action)
	return c.mutateJSON(http.MethodDelete, "/v1/auth/permissions", q, nil)
}
