// Package client talks to Nacos Open API (accessToken query auth).
package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/xiaoxiaolin0918/nacos-cli/internal/config"
	"github.com/xiaoxiaolin0918/nacos-cli/internal/domain"
	"github.com/xiaoxiaolin0918/nacos-cli/internal/version"
)

// Client is the Nacos Open API adapter.
type Client struct {
	Settings config.Settings
	HTTP     *http.Client
	Token    string
	AuthMode string
}

func New(settings config.Settings) *Client {
	return &Client{
		Settings: settings,
		HTTP:     &http.Client{Timeout: 60 * time.Second},
	}
}

func (c *Client) base() string {
	return c.Settings.BaseURL()
}

func (c *Client) ns(explicit string) string {
	if explicit != "" {
		return explicit
	}
	return c.Settings.Namespace
}

type LoginResult struct {
	AuthMode  string `json:"auth_mode"`
	HasToken  bool   `json:"has_token"`
	Username  string `json:"username"`
	Server    string `json:"server"`
	TokenTTL  any    `json:"ttl,omitempty"`
	AccessTok string `json:"-"`
}

func (c *Client) Login(ctx context.Context, username, password string) (LoginResult, error) {
	form := url.Values{}
	form.Set("username", username)
	form.Set("password", password)
	rawURL := c.base() + "/v1/auth/login"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, rawURL, strings.NewReader(form.Encode()))
	if err != nil {
		return LoginResult{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", version.UserAgent())
	res, err := c.HTTP.Do(req)
	if err != nil {
		return LoginResult{}, err
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)
	var data map[string]any
	_ = json.Unmarshal(body, &data)
	if res.StatusCode >= 400 {
		msg := string(body)
		if data != nil {
			if m, ok := data["message"].(string); ok && m != "" {
				msg = m
			} else if m, ok := data["msg"].(string); ok && m != "" {
				msg = m
			}
		}
		return LoginResult{}, fmt.Errorf("login failed HTTP %d: %s", res.StatusCode, truncate(msg, 200))
	}
	token, _ := data["accessToken"].(string)
	if token == "" || token == "AUTH_DISABLED" {
		c.Token = ""
		c.AuthMode = "disabled"
		return LoginResult{
			AuthMode: "disabled",
			HasToken: false,
			Username: username,
			Server:   c.Settings.Server,
		}, nil
	}
	c.Token = token
	c.AuthMode = "token"
	return LoginResult{
		AuthMode:  "token",
		HasToken:  true,
		Username:  username,
		Server:    c.Settings.Server,
		TokenTTL:  data["tokenTtl"],
		AccessTok: token,
	}, nil
}

func (c *Client) ApplyToken(token, mode string) {
	c.AuthMode = mode
	if mode == "disabled" || token == "" || token == "AUTH_DISABLED" {
		c.Token = ""
		return
	}
	c.Token = token
}

func (c *Client) withToken(q url.Values) url.Values {
	if q == nil {
		q = url.Values{}
	}
	if c.Token != "" {
		q.Set("accessToken", c.Token)
	}
	return q
}

func (c *Client) do(ctx context.Context, method, path string, query url.Values, form url.Values) ([]byte, int, string, error) {
	q := c.withToken(cloneValues(query))
	u := c.base() + path
	if len(q) > 0 {
		u += "?" + q.Encode()
	}
	var body io.Reader
	headers := map[string]string{"User-Agent": version.UserAgent(), "Accept": "application/json, text/plain, */*"}
	if form != nil {
		body = strings.NewReader(form.Encode())
		headers["Content-Type"] = "application/x-www-form-urlencoded"
	}
	req, err := http.NewRequestWithContext(ctx, method, u, body)
	if err != nil {
		return nil, 0, u, err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	res, err := c.HTTP.Do(req)
	if err != nil {
		return nil, 0, u, err
	}
	defer res.Body.Close()
	b, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, res.StatusCode, u, err
	}
	return b, res.StatusCode, u, nil
}

func (c *Client) getJSON(ctx context.Context, path string, query url.Values) (any, error) {
	b, status, _, err := c.do(ctx, http.MethodGet, path, query, nil)
	if err != nil {
		return nil, err
	}
	if status == 404 {
		return nil, domain.ErrNotFound
	}
	if status >= 400 {
		return nil, fmt.Errorf("HTTP %d: %s", status, truncate(string(b), 200))
	}
	ctLooksJSON := len(b) > 0 && (b[0] == '{' || b[0] == '[')
	if ctLooksJSON {
		var v any
		if err := json.Unmarshal(b, &v); err == nil {
			return v, nil
		}
	}
	return string(b), nil
}

func (c *Client) getText(ctx context.Context, path string, query url.Values) (string, error) {
	b, status, _, err := c.do(ctx, http.MethodGet, path, query, nil)
	if err != nil {
		return "", err
	}
	if status == 404 {
		return "", domain.ErrNotFound
	}
	if status >= 400 {
		// Nacos config get often returns 404-ish messages as text
		msg := strings.ToLower(string(b))
		if strings.Contains(msg, "not found") || status == 404 {
			return "", domain.ErrNotFound
		}
		return "", fmt.Errorf("HTTP %d: %s", status, truncate(string(b), 200))
	}
	return string(b), nil
}

func (c *Client) ConfigGet(dataID, group, namespace string) (string, error) {
	q := url.Values{}
	q.Set("dataId", dataID)
	q.Set("group", group)
	q.Set("tenant", c.ns(namespace))
	return c.getText(context.Background(), "/v1/cs/configs", q)
}

func (c *Client) ConfigList(dataID, group, namespace string, pageNo, pageSize int) (any, error) {
	q := url.Values{}
	q.Set("search", "blur")
	q.Set("dataId", dataID) // may be empty string — required by some builds
	q.Set("group", group)
	q.Set("tenant", c.ns(namespace))
	q.Set("pageNo", strconv.Itoa(pageNo))
	q.Set("pageSize", strconv.Itoa(pageSize))
	return c.getJSON(context.Background(), "/v1/cs/configs", q)
}

func (c *Client) ConfigPublish(dataID, group, content, namespace, typ string) (any, error) {
	form := url.Values{}
	form.Set("dataId", dataID)
	form.Set("group", group)
	form.Set("content", content)
	form.Set("tenant", c.ns(namespace))
	form.Set("type", typ)
	b, status, _, err := c.do(context.Background(), http.MethodPost, "/v1/cs/configs", nil, form)
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

// ConfigType returns the stored config type (show=all), "" when unknown.
func (c *Client) ConfigType(dataID, group, namespace string) (string, error) {
	q := url.Values{}
	q.Set("dataId", dataID)
	q.Set("group", group)
	q.Set("tenant", c.ns(namespace))
	q.Set("show", "all")
	v, err := c.getJSON(context.Background(), "/v1/cs/configs", q)
	if err != nil {
		return "", err
	}
	if m, ok := v.(map[string]any); ok {
		if t, ok := m["type"].(string); ok {
			return t, nil
		}
	}
	return "", nil
}

// ConfigDelete removes one config (DELETE /v1/cs/configs).
func (c *Client) ConfigDelete(dataID, group, namespace string) (any, error) {
	q := url.Values{}
	q.Set("dataId", dataID)
	q.Set("group", group)
	q.Set("tenant", c.ns(namespace))
	b, status, _, err := c.do(context.Background(), http.MethodDelete, "/v1/cs/configs", q, nil)
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

func (c *Client) ConfigHistory(dataID, group, namespace string, pageNo, pageSize int) (any, error) {
	q := url.Values{}
	q.Set("search", "accurate")
	q.Set("dataId", dataID)
	q.Set("group", group)
	q.Set("tenant", c.ns(namespace))
	q.Set("pageNo", strconv.Itoa(pageNo))
	q.Set("pageSize", strconv.Itoa(pageSize))
	return c.getJSON(context.Background(), "/v1/cs/history", q)
}

func (c *Client) ConfigHistoryDetail(dataID, group, namespace, nid string) (map[string]any, error) {
	q := url.Values{}
	q.Set("dataId", dataID)
	q.Set("group", group)
	q.Set("tenant", c.ns(namespace))
	q.Set("nid", nid)
	v, err := c.getJSON(context.Background(), "/v1/cs/history", q)
	if err != nil {
		return nil, err
	}
	m, ok := v.(map[string]any)
	if !ok {
		return nil, errors.New("unexpected history detail shape")
	}
	return m, nil
}

func (c *Client) ServiceList(namespace string, pageNo, pageSize int) (any, error) {
	q := url.Values{}
	q.Set("pageNo", strconv.Itoa(pageNo))
	q.Set("pageSize", strconv.Itoa(pageSize))
	q.Set("namespaceId", c.ns(namespace))
	return c.getJSON(context.Background(), "/v1/ns/service/list", q)
}

func (c *Client) InstanceList(service, namespace, group string) (any, error) {
	q := url.Values{}
	q.Set("serviceName", service)
	q.Set("namespaceId", c.ns(namespace))
	q.Set("groupName", group)
	return c.getJSON(context.Background(), "/v1/ns/instance/list", q)
}

func (c *Client) InstanceRegister(service, ip string, port int, namespace, group, cluster string, ephemeral bool) (any, error) {
	form := url.Values{}
	form.Set("serviceName", service)
	form.Set("ip", ip)
	form.Set("port", strconv.Itoa(port))
	form.Set("namespaceId", c.ns(namespace))
	form.Set("groupName", group)
	form.Set("clusterName", cluster)
	form.Set("ephemeral", strconv.FormatBool(ephemeral))
	form.Set("weight", "1.0")
	form.Set("enabled", "true")
	form.Set("healthy", "true")
	b, status, _, err := c.do(context.Background(), http.MethodPost, "/v1/ns/instance", nil, form)
	if err != nil {
		return nil, err
	}
	if status >= 400 {
		return nil, fmt.Errorf("HTTP %d: %s", status, truncate(string(b), 200))
	}
	return strings.TrimSpace(string(b)), nil
}

func (c *Client) InstanceDeregister(service, ip string, port int, namespace, group, cluster string, ephemeral bool) (any, error) {
	q := url.Values{}
	q.Set("serviceName", service)
	q.Set("ip", ip)
	q.Set("port", strconv.Itoa(port))
	q.Set("namespaceId", c.ns(namespace))
	q.Set("groupName", group)
	q.Set("clusterName", cluster)
	q.Set("ephemeral", strconv.FormatBool(ephemeral))
	b, status, _, err := c.do(context.Background(), http.MethodDelete, "/v1/ns/instance", q, nil)
	if err != nil {
		return nil, err
	}
	if status >= 400 {
		return nil, fmt.Errorf("HTTP %d: %s", status, truncate(string(b), 200))
	}
	return strings.TrimSpace(string(b)), nil
}

func (c *Client) InstanceEnabled(service, ip string, port int, enabled bool, namespace, group, cluster string, ephemeral bool) (any, error) {
	q := url.Values{}
	q.Set("serviceName", service)
	q.Set("ip", ip)
	q.Set("port", strconv.Itoa(port))
	q.Set("enabled", strconv.FormatBool(enabled))
	q.Set("namespaceId", c.ns(namespace))
	q.Set("groupName", group)
	q.Set("clusterName", cluster)
	q.Set("ephemeral", strconv.FormatBool(ephemeral))
	b, status, _, err := c.do(context.Background(), http.MethodPut, "/v1/ns/instance", q, nil)
	if err != nil {
		return nil, err
	}
	if status >= 400 {
		return nil, fmt.Errorf("HTTP %d: %s", status, truncate(string(b), 200))
	}
	return strings.TrimSpace(string(b)), nil
}

func (c *Client) NamespaceList() (any, error) {
	return c.getJSON(context.Background(), "/v1/console/namespaces", nil)
}

func (c *Client) Probe(ctx context.Context) map[string]any {
	out := map[string]any{
		"auth_mode": c.AuthMode,
		"has_token": c.Token != "",
	}
	v, err := c.NamespaceList()
	if err != nil {
		out["ok"] = false
		out["error"] = err.Error()
		return out
	}
	out["ok"] = true
	if m, ok := v.(map[string]any); ok {
		if data, ok := m["data"].([]any); ok {
			n := 3
			if len(data) < n {
				n = len(data)
			}
			out["namespaces_sample"] = data[:n]
		}
	}
	return out
}

func (c *Client) Ping(ctx context.Context) (int, error) {
	u := strings.TrimRight(c.Settings.Server, "/") + "/"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("User-Agent", version.UserAgent())
	res, err := c.HTTP.Do(req)
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(res.Body, 4096))
	return res.StatusCode, nil
}

func cloneValues(in url.Values) url.Values {
	if in == nil {
		return url.Values{}
	}
	out := url.Values{}
	for k, vs := range in {
		for _, v := range vs {
			out.Add(k, v)
		}
	}
	return out
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

// Ensure Client implements domain ports.
var (
	_ domain.ConfigPort    = (*Client)(nil)
	_ domain.ServicePort   = (*Client)(nil)
	_ domain.NamespacePort = (*Client)(nil)
)
