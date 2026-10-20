# -*- coding: utf-8 -*-
"""In-memory / recording fakes for domain ports (no real HTTP)."""
from __future__ import annotations

from typing import Any


class FakeConfigPort:
    def __init__(self) -> None:
        self.store: dict[tuple[str, str, str], str] = {}
        self.history: dict[tuple[str, str, str], list[dict[str, Any]]] = {}
        self.calls: list[tuple[str, tuple, dict]] = []

    def _key(self, data_id: str, group: str, namespace: str | None) -> tuple[str, str, str]:
        return (data_id, group, namespace or "")

    def config_get(self, data_id: str, group: str, namespace: str | None = None) -> Any:
        self.calls.append(("config_get", (data_id, group, namespace), {}))
        key = self._key(data_id, group, namespace)
        if key not in self.store:
            raise KeyError(f"config not found: {key}")
        return self.store[key]

    def config_list(
        self,
        data_id: str = "",
        group: str = "",
        namespace: str | None = None,
        page_no: int = 1,
        page_size: int = 100,
    ) -> Any:
        self.calls.append(
            ("config_list", (data_id, group, namespace, page_no, page_size), {})
        )
        items = []
        for (did, grp, ns), content in self.store.items():
            if namespace is not None and ns != (namespace or ""):
                continue
            if data_id and data_id not in did:
                continue
            if group and group != grp:
                continue
            items.append({"dataId": did, "group": grp, "tenant": ns, "content": content})
        return {"totalCount": len(items), "pageItems": items, "pageNumber": page_no}

    def config_publish(
        self,
        data_id: str,
        group: str,
        content: str,
        namespace: str | None = None,
        type_: str = "text",
    ) -> Any:
        self.calls.append(
            ("config_publish", (data_id, group, namespace, type_), {"content": content})
        )
        key = self._key(data_id, group, namespace)
        self.store[key] = content
        hist = self.history.setdefault(key, [])
        nid = str(len(hist) + 1)
        hist.append({"id": nid, "content": content, "dataId": data_id, "group": group})
        return True

    def config_history(
        self,
        data_id: str,
        group: str,
        namespace: str | None = None,
        page_no: int = 1,
        page_size: int = 20,
    ) -> Any:
        self.calls.append(
            ("config_history", (data_id, group, namespace, page_no, page_size), {})
        )
        key = self._key(data_id, group, namespace)
        items = list(reversed(self.history.get(key, [])))
        return {"totalCount": len(items), "pageItems": items}

    def config_rollback(
        self,
        data_id: str,
        group: str,
        nid: str,
        namespace: str | None = None,
    ) -> Any:
        self.calls.append(("config_rollback", (data_id, group, nid, namespace), {}))
        key = self._key(data_id, group, namespace)
        for item in self.history.get(key, []):
            if str(item.get("id")) == str(nid):
                return self.config_publish(data_id, group, item["content"], namespace)
        raise RuntimeError(f"history nid={nid} not found")


class FakeServicePort:
    def __init__(self) -> None:
        # key: (service, ns, group) -> list of instance dicts
        self.instances: dict[tuple[str, str, str], list[dict[str, Any]]] = {}
        self.calls: list[tuple[str, dict]] = []

    def service_list(
        self, namespace: str | None = None, page_no: int = 1, page_size: int = 100
    ) -> Any:
        self.calls.append(("service_list", {"namespace": namespace}))
        names = sorted({k[0] for k in self.instances if k[1] == (namespace or "")})
        return {"count": len(names), "doms": names}

    def instance_list(
        self,
        service_name: str,
        namespace: str | None = None,
        group: str = "DEFAULT_GROUP",
    ) -> Any:
        self.calls.append(
            ("instance_list", {"service": service_name, "namespace": namespace, "group": group})
        )
        key = (service_name, namespace or "", group)
        hosts = self.instances.get(key, [])
        return {"hosts": hosts, "name": service_name}

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
        self.calls.append(
            (
                "instance_enabled",
                {
                    "service": service_name,
                    "ip": ip,
                    "port": port,
                    "enabled": enabled,
                    "namespace": namespace,
                    "group": group,
                    "cluster": cluster,
                    "ephemeral": ephemeral,
                },
            )
        )
        key = (service_name, namespace or "", group)
        for inst in self.instances.get(key, []):
            if inst["ip"] == ip and inst["port"] == port:
                inst["enabled"] = enabled
                return "ok"
        raise KeyError("instance not found")

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
        self.calls.append(
            (
                "instance_register",
                {
                    "service": service_name,
                    "ip": ip,
                    "port": port,
                    "ephemeral": ephemeral,
                },
            )
        )
        key = (service_name, namespace or "", group)
        self.instances.setdefault(key, []).append(
            {
                "ip": ip,
                "port": port,
                "enabled": enabled,
                "healthy": healthy,
                "weight": weight,
                "clusterName": cluster,
                "ephemeral": ephemeral,
            }
        )
        return "ok"

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
        self.calls.append(
            (
                "instance_deregister",
                {
                    "service": service_name,
                    "ip": ip,
                    "port": port,
                    "ephemeral": ephemeral,
                },
            )
        )
        key = (service_name, namespace or "", group)
        hosts = self.instances.get(key, [])
        self.instances[key] = [h for h in hosts if not (h["ip"] == ip and h["port"] == port)]
        return "ok"
