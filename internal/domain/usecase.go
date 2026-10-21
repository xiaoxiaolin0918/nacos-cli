// Package domain holds use cases independent of HTTP and cobra.
package domain

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrExists   = errors.New("config already exists")
	ErrNotFound = errors.New("not found")
)

// ConfigPort is the infrastructure seam for config Open API.
type ConfigPort interface {
	ConfigGet(dataID, group, namespace string) (string, error)
	ConfigList(dataID, group, namespace string, pageNo, pageSize int) (any, error)
	ConfigPublish(dataID, group, content, namespace, typ string) (any, error)
	ConfigHistory(dataID, group, namespace string, pageNo, pageSize int) (any, error)
	ConfigHistoryDetail(dataID, group, namespace, nid string) (map[string]any, error)
}

type ConfigUseCase struct {
	Port ConfigPort
}

func (uc ConfigUseCase) Get(dataID, group, namespace string) (string, error) {
	if strings.TrimSpace(dataID) == "" {
		return "", errors.New("data_id is required")
	}
	if group == "" {
		group = "DEFAULT_GROUP"
	}
	return uc.Port.ConfigGet(dataID, group, namespace)
}

func (uc ConfigUseCase) List(dataID, group, namespace string, pageNo, pageSize int) (any, error) {
	if pageNo <= 0 {
		pageNo = 1
	}
	if pageSize <= 0 {
		pageSize = 100
	}
	return uc.Port.ConfigList(dataID, group, namespace, pageNo, pageSize)
}

func (uc ConfigUseCase) Publish(dataID, group, content, namespace, typ string, createOnly bool) (any, error) {
	if strings.TrimSpace(dataID) == "" {
		return nil, errors.New("data_id is required")
	}
	if group == "" {
		group = "DEFAULT_GROUP"
	}
	if typ == "" {
		typ = "text"
	}
	if createOnly {
		exists, err := uc.exists(dataID, group, namespace)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, fmt.Errorf("%w: dataId=%s group=%s namespace=%s", ErrExists, dataID, group, namespace)
		}
	}
	return uc.Port.ConfigPublish(dataID, group, content, namespace, typ)
}

func (uc ConfigUseCase) History(dataID, group, namespace string, pageNo, pageSize int) (any, error) {
	if strings.TrimSpace(dataID) == "" {
		return nil, errors.New("data_id is required")
	}
	if group == "" {
		group = "DEFAULT_GROUP"
	}
	if pageNo <= 0 {
		pageNo = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	return uc.Port.ConfigHistory(dataID, group, namespace, pageNo, pageSize)
}

// ConfigTypeGetter is an optional port: lets rollback keep the current config type.
type ConfigTypeGetter interface {
	ConfigType(dataID, group, namespace string) (string, error)
}

// ConfigDeleter is an optional port for removing a config.
type ConfigDeleter interface {
	ConfigDelete(dataID, group, namespace string) (any, error)
}

func (uc ConfigUseCase) Rollback(dataID, group, namespace, nid string) (any, error) {
	if strings.TrimSpace(dataID) == "" {
		return nil, errors.New("data_id is required")
	}
	if strings.TrimSpace(nid) == "" {
		return nil, errors.New("nid is required")
	}
	if group == "" {
		group = "DEFAULT_GROUP"
	}
	detail, err := uc.Port.ConfigHistoryDetail(dataID, group, namespace, nid)
	if err != nil {
		return nil, err
	}
	if hid, ok := detail["dataId"].(string); ok && hid != "" && hid != dataID {
		return nil, fmt.Errorf("history nid=%s belongs to dataId=%s, not %s", nid, hid, dataID)
	}
	if hg, ok := detail["group"].(string); ok && hg != "" && hg != group {
		return nil, fmt.Errorf("history nid=%s belongs to group=%s, not %s", nid, hg, group)
	}
	content, _ := detail["content"].(string)
	if content == "" {
		if v, ok := detail["contentData"].(string); ok {
			content = v
		}
	}
	if content == "" {
		return nil, fmt.Errorf("history nid=%s has no content", nid)
	}
	typ := "text"
	if tg, ok := uc.Port.(ConfigTypeGetter); ok {
		if t, err := tg.ConfigType(dataID, group, namespace); err == nil && strings.TrimSpace(t) != "" {
			typ = t
		}
	}
	return uc.Port.ConfigPublish(dataID, group, content, namespace, typ)
}

// Delete removes an existing config; refuses when it does not exist.
func (uc ConfigUseCase) Delete(dataID, group, namespace string) (any, error) {
	if strings.TrimSpace(dataID) == "" {
		return nil, errors.New("data_id is required")
	}
	if group == "" {
		group = "DEFAULT_GROUP"
	}
	del, ok := uc.Port.(ConfigDeleter)
	if !ok {
		return nil, errors.New("config delete not supported by this port")
	}
	exists, err := uc.exists(dataID, group, namespace)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("%w: dataId=%s group=%s namespace=%s", ErrNotFound, dataID, group, namespace)
	}
	return del.ConfigDelete(dataID, group, namespace)
}

// BlurPattern wraps a non-empty filter with * (Nacos blur search does not add wildcards itself).
func BlurPattern(s string) string {
	if s == "" || strings.Contains(s, "*") {
		return s
	}
	return "*" + s + "*"
}

func (uc ConfigUseCase) exists(dataID, group, namespace string) (bool, error) {
	_, err := uc.Port.ConfigGet(dataID, group, namespace)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, ErrNotFound) {
		return false, nil
	}
	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "404") || strings.Contains(msg, "not found") {
		return false, nil
	}
	return false, err
}

// ServicePort is the infrastructure seam for naming Open API.
type ServicePort interface {
	ServiceList(namespace string, pageNo, pageSize int) (any, error)
	InstanceList(service, namespace, group string) (any, error)
	InstanceRegister(service, ip string, port int, namespace, group, cluster string, ephemeral bool) (any, error)
	InstanceDeregister(service, ip string, port int, namespace, group, cluster string, ephemeral bool) (any, error)
	InstanceEnabled(service, ip string, port int, enabled bool, namespace, group, cluster string, ephemeral bool) (any, error)
}

type ServiceUseCase struct {
	Port ServicePort
}

func requireInstance(service, ip string, port int) error {
	if strings.TrimSpace(service) == "" {
		return errors.New("service_name is required")
	}
	if strings.TrimSpace(ip) == "" {
		return errors.New("ip is required")
	}
	if port <= 0 {
		return errors.New("port is required")
	}
	return nil
}

func (uc ServiceUseCase) List(namespace string, pageNo, pageSize int) (any, error) {
	if pageNo <= 0 {
		pageNo = 1
	}
	if pageSize <= 0 {
		pageSize = 100
	}
	return uc.Port.ServiceList(namespace, pageNo, pageSize)
}

func (uc ServiceUseCase) Instances(service, namespace, group string) (any, error) {
	if strings.TrimSpace(service) == "" {
		return nil, errors.New("service_name is required")
	}
	if group == "" {
		group = "DEFAULT_GROUP"
	}
	return uc.Port.InstanceList(service, namespace, group)
}

func (uc ServiceUseCase) Register(service, ip string, port int, namespace, group, cluster string, ephemeral bool) (any, error) {
	if err := requireInstance(service, ip, port); err != nil {
		return nil, err
	}
	if group == "" {
		group = "DEFAULT_GROUP"
	}
	if cluster == "" {
		cluster = "DEFAULT"
	}
	return uc.Port.InstanceRegister(service, ip, port, namespace, group, cluster, ephemeral)
}

func (uc ServiceUseCase) Deregister(service, ip string, port int, namespace, group, cluster string, ephemeral bool) (any, error) {
	if err := requireInstance(service, ip, port); err != nil {
		return nil, err
	}
	if group == "" {
		group = "DEFAULT_GROUP"
	}
	if cluster == "" {
		cluster = "DEFAULT"
	}
	return uc.Port.InstanceDeregister(service, ip, port, namespace, group, cluster, ephemeral)
}

func (uc ServiceUseCase) Up(service, ip string, port int, namespace, group, cluster string, ephemeral bool) (any, error) {
	if err := requireInstance(service, ip, port); err != nil {
		return nil, err
	}
	if group == "" {
		group = "DEFAULT_GROUP"
	}
	if cluster == "" {
		cluster = "DEFAULT"
	}
	return uc.Port.InstanceEnabled(service, ip, port, true, namespace, group, cluster, ephemeral)
}

func (uc ServiceUseCase) Down(service, ip string, port int, namespace, group, cluster string, ephemeral bool) (any, error) {
	if err := requireInstance(service, ip, port); err != nil {
		return nil, err
	}
	if group == "" {
		group = "DEFAULT_GROUP"
	}
	if cluster == "" {
		cluster = "DEFAULT"
	}
	return uc.Port.InstanceEnabled(service, ip, port, false, namespace, group, cluster, ephemeral)
}

type NamespacePort interface {
	NamespaceList() (any, error)
}

type NamespaceUseCase struct {
	Port NamespacePort
}

func (uc NamespaceUseCase) List() (any, error) {
	return uc.Port.NamespaceList()
}
