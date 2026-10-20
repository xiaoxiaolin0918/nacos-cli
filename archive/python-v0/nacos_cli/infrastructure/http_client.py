# -*- coding: utf-8 -*-
"""Nacos Open API HTTP client (infrastructure)."""
from __future__ import annotations

from typing import Any

import requests

from ..config import Settings


class NacosClient:
    def __init__(self, settings: Settings) -> None:
        self.s = settings
        self._token: str | None = None
        self.session = requests.Session()
        self.session.headers.update({"User-Agent": "nacos-cli/0.1.0"})

    def login(self, username: str | None = None, password: str | None = None) -> dict:
        """Authenticate via Open API. Returns auth metadata (never includes password).

        When Nacos auth is disabled, returns auth_mode=disabled without raising.
        """
        from .credentials import CredentialStore

        user = (username if username is not None else self.s.username) or ""
        # Explicit password argument wins; else settings password; else try store.
        if password is not None:
            return self.auth_login(user, password)
        if self.s.password:
            return self.auth_login(user or self.s.username or "", self.s.password)

        stored = CredentialStore().load()
        if stored.auth_mode == "disabled":
            self._token = None
            self._auth_mode = "disabled"
            return {
                "auth_mode": "disabled",
                "has_token": False,
                "username": stored.username or user,
                "server": self.s.server,
                "token_source": "credentials_file",
            }
        if stored.access_token and stored.access_token != "AUTH_DISABLED":
            self._token = stored.access_token
            self._auth_mode = "token"
            return {
                "auth_mode": "token",
                "has_token": True,
                "username": stored.username or user,
                "server": self.s.server,
                "token_source": "credentials_file",
            }
        # No stored creds: attempt login with empty password (AUTH_DISABLED servers).
        if user:
            return self.auth_login(user, "")
        self._auth_mode = "none"
        return {"auth_mode": "none", "has_token": False, "server": self.s.server}

    def auth_login(self, username: str, password: str) -> dict:
        url = f"{self.s.base_url}/v1/auth/login"
        r = self.session.post(
            url,
            data={"username": username, "password": password},
            timeout=30,
        )
        # Some builds return 200 with AUTH_DISABLED; others may 403 when wrong
        ct = r.headers.get("Content-Type", "")
        data = r.json() if "application/json" in ct else {}
        if r.status_code >= 400:
            msg = ""
            if isinstance(data, dict):
                msg = str(data.get("message") or data.get("msg") or data)
            raise RuntimeError(
                f"login failed HTTP {r.status_code}: {msg or r.text[:200]}"
            )
        token = ""
        if isinstance(data, dict):
            token = str(data.get("accessToken") or "")
        if not token or token == "AUTH_DISABLED":
            self._token = None
            self._auth_mode = "disabled"
            return {
                "auth_mode": "disabled",
                "has_token": False,
                "username": username,
                "server": self.s.server,
            }
        self._token = token
        self._auth_mode = "token"
        return {
            "auth_mode": "token",
            "has_token": True,
            "username": username,
            "server": self.s.server,
            "ttl": data.get("tokenTtl") if isinstance(data, dict) else None,
        }

    def auth_apply_token(self, access_token: str | None, auth_mode: str) -> None:
        self._auth_mode = auth_mode or "none"
        if auth_mode == "disabled" or not access_token or access_token == "AUTH_DISABLED":
            self._token = None
        else:
            self._token = access_token

    def auth_probe(self) -> dict:
        """Read-only probe: namespace list (works with or without auth)."""
        try:
            data = self.namespace_list()
            ok = True
            err = ""
        except Exception as e:  # noqa: BLE001
            data = None
            ok = False
            err = str(e)
        return {
            "ok": ok,
            "auth_mode": getattr(self, "_auth_mode", None) or ("token" if self._token else "none"),
            "has_token": bool(self._token),
            "error": err,
            "namespaces_sample": (
                (data.get("data") or data)[:3]
                if isinstance(data, dict) and isinstance(data.get("data"), list)
                else (data if not isinstance(data, dict) else {"keys": list(data.keys())[:5]})
            ),
        }

    def _params(self, extra: dict[str, Any] | None = None) -> dict[str, Any]:
        p: dict[str, Any] = {}
        if self._token:
            p["accessToken"] = self._token
        if extra:
            # 保留空字符串：部分 Open API（如 configs?search=blur）缺 dataId= 会 500
            p.update({k: v for k, v in extra.items() if v is not None})
        return p

    def get(self, path: str, params: dict[str, Any] | None = None) -> Any:
        r = self.session.get(
            f"{self.s.base_url}{path}",
            params=self._params(params),
            timeout=60,
        )
        r.raise_for_status()
        ct = r.headers.get("Content-Type", "")
        if "application/json" in ct:
            return r.json()
        return r.text

    def post_form(self, path: str, data: dict[str, Any] | None = None) -> Any:
        r = self.session.post(
            f"{self.s.base_url}{path}",
            params=self._params(),
            data=data or {},
            timeout=60,
        )
        r.raise_for_status()
        ct = r.headers.get("Content-Type", "")
        if "application/json" in ct:
            try:
                return r.json()
            except Exception:
                return r.text
        return r.text

    def put_form(
        self,
        path: str,
        data: dict[str, Any] | None = None,
        params: dict[str, Any] | None = None,
    ) -> Any:
        r = self.session.put(
            f"{self.s.base_url}{path}",
            params=self._params(params),
            data=data or {},
            timeout=60,
        )
        r.raise_for_status()
        ct = r.headers.get("Content-Type", "")
        if "application/json" in ct:
            try:
                return r.json()
            except Exception:
                return r.text
        return r.text

    def delete_form(
        self,
        path: str,
        data: dict[str, Any] | None = None,
        params: dict[str, Any] | None = None,
    ) -> Any:
        r = self.session.delete(
            f"{self.s.base_url}{path}",
            params=self._params(params),
            data=data or {},
            timeout=60,
        )
        r.raise_for_status()
        ct = r.headers.get("Content-Type", "")
        if "application/json" in ct:
            try:
                return r.json()
            except Exception:
                return r.text
        return r.text

    def config_get(self, data_id: str, group: str, namespace: str | None = None) -> str:
        tenant = namespace if namespace is not None else self.s.namespace
        return self.get(
            "/v1/cs/configs",
            {"dataId": data_id, "group": group, "tenant": tenant},
        )

    def config_list(
        self,
        data_id: str = "",
        group: str = "",
        namespace: str | None = None,
        page_no: int = 1,
        page_size: int = 100,
    ) -> Any:
        tenant = namespace if namespace is not None else self.s.namespace
        return self.get(
            "/v1/cs/configs",
            {
                "search": "blur",
                "dataId": data_id,
                "group": group,
                "tenant": tenant,
                "pageNo": page_no,
                "pageSize": page_size,
            },
        )

    def config_publish(
        self,
        data_id: str,
        group: str,
        content: str,
        namespace: str | None = None,
        type_: str = "text",
    ) -> Any:
        tenant = namespace if namespace is not None else self.s.namespace
        return self.post_form(
            "/v1/cs/configs",
            {
                "dataId": data_id,
                "group": group,
                "content": content,
                "tenant": tenant,
                "type": type_,
            },
        )

    def config_history(
        self,
        data_id: str,
        group: str,
        namespace: str | None = None,
        page_no: int = 1,
        page_size: int = 20,
    ) -> Any:
        tenant = namespace if namespace is not None else self.s.namespace
        return self.get(
            "/v1/cs/history",
            {
                "search": "accurate",
                "dataId": data_id,
                "group": group,
                "tenant": tenant,
                "pageNo": page_no,
                "pageSize": page_size,
            },
        )

    def config_rollback(
        self,
        data_id: str,
        group: str,
        nid: str,
        namespace: str | None = None,
    ) -> Any:
        """基于历史 id 回滚：先取 history 详情内容，再 publish。"""
        tenant = namespace if namespace is not None else self.s.namespace
        detail = self.get(
            "/v1/cs/history",
            {"dataId": data_id, "group": group, "tenant": tenant, "nid": nid},
        )
        if isinstance(detail, dict):
            content = detail.get("content") or detail.get("contentData") or ""
        else:
            content = str(detail)
        if not content:
            raise RuntimeError(f"历史 nid={nid} 无内容，无法回滚")
        return self.config_publish(data_id, group, content, namespace=tenant)

    def service_list(
        self, namespace: str | None = None, page_no: int = 1, page_size: int = 100
    ) -> Any:
        ns = namespace if namespace is not None else self.s.namespace
        return self.get(
            "/v1/ns/service/list",
            {"pageNo": page_no, "pageSize": page_size, "namespaceId": ns},
        )

    def instance_list(
        self, service_name: str, namespace: str | None = None, group: str = "DEFAULT_GROUP"
    ) -> Any:
        ns = namespace if namespace is not None else self.s.namespace
        return self.get(
            "/v1/ns/instance/list",
            {"serviceName": service_name, "namespaceId": ns, "groupName": group},
        )

    def instance_enabled(
        self,
        service_name: str,
        ip: str,
        port: int,
        enabled: bool,
        namespace: str | None = None,
        group: str = "DEFAULT_GROUP",
        cluster: str = "DEFAULT",
        ephemeral: bool = True,
    ) -> Any:
        """上下线：PUT /v1/ns/instance，参数必须走 query（body 会被忽略）。"""
        ns = namespace if namespace is not None else self.s.namespace
        return self.put_form(
            "/v1/ns/instance",
            data=None,
            params={
                "serviceName": service_name,
                "ip": ip,
                "port": port,
                "enabled": str(enabled).lower(),
                "namespaceId": ns,
                "groupName": group,
                "clusterName": cluster,
                "ephemeral": str(ephemeral).lower(),
            },
        )

    def instance_catalog(
        self,
        service_name: str,
        namespace: str | None = None,
        group: str = "DEFAULT_GROUP",
        cluster: str = "DEFAULT",
        page_no: int = 1,
        page_size: int = 100,
    ) -> Any:
        """控制台同款实例列表（含 enabled=false / unhealthy），用于运维核验。"""
        ns = namespace if namespace is not None else self.s.namespace
        return self.get(
            "/v1/ns/catalog/instances",
            {
                "serviceName": service_name,
                "clusterName": cluster,
                "groupName": group,
                "namespaceId": ns,
                "pageNo": page_no,
                "pageSize": page_size,
            },
        )

    def instance_register(
        self,
        service_name: str,
        ip: str,
        port: int,
        namespace: str | None = None,
        group: str = "DEFAULT_GROUP",
        cluster: str = "DEFAULT",
        ephemeral: bool = True,
        weight: float = 1.0,
        enabled: bool = True,
        healthy: bool = True,
        metadata: str = "",
    ) -> Any:
        """Register instance. Default ephemeral=true (temporary; use persistent=False only when explicit)."""
        ns = namespace if namespace is not None else self.s.namespace
        data: dict[str, Any] = {
            "serviceName": service_name,
            "ip": ip,
            "port": port,
            "namespaceId": ns,
            "groupName": group,
            "clusterName": cluster,
            "ephemeral": str(ephemeral).lower(),
            "weight": weight,
            "enabled": str(enabled).lower(),
            "healthy": str(healthy).lower(),
        }
        if metadata:
            data["metadata"] = metadata
        return self.post_form("/v1/ns/instance", data=data)

    def instance_deregister(
        self,
        service_name: str,
        ip: str,
        port: int,
        namespace: str | None = None,
        group: str = "DEFAULT_GROUP",
        cluster: str = "DEFAULT",
        ephemeral: bool = True,
    ) -> Any:
        """Deregister instance. ephemeral must match registration."""
        ns = namespace if namespace is not None else self.s.namespace
        # DELETE 与官方 curl 一致：参数走 query；body 在部分环境下清不干净
        return self.delete_form(
            "/v1/ns/instance",
            data=None,
            params={
                "serviceName": service_name,
                "ip": ip,
                "port": port,
                "namespaceId": ns,
                "groupName": group,
                "clusterName": cluster,
                "ephemeral": str(ephemeral).lower(),
            },
        )

    def namespace_list(self) -> Any:
        return self.get("/v1/console/namespaces")
