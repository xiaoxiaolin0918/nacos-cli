# -*- coding: utf-8 -*-
"""Ports (protocols) for Nacos domain use cases."""
from __future__ import annotations

from typing import Any, Protocol


class ConfigPort(Protocol):
    def config_get(
        self, data_id: str, group: str, namespace: str | None = None
    ) -> Any: ...

    def config_list(
        self,
        data_id: str = "",
        group: str = "",
        namespace: str | None = None,
        page_no: int = 1,
        page_size: int = 100,
    ) -> Any: ...

    def config_publish(
        self,
        data_id: str,
        group: str,
        content: str,
        namespace: str | None = None,
        type_: str = "text",
    ) -> Any: ...

    def config_history(
        self,
        data_id: str,
        group: str,
        namespace: str | None = None,
        page_no: int = 1,
        page_size: int = 20,
    ) -> Any: ...

    def config_rollback(
        self,
        data_id: str,
        group: str,
        nid: str,
        namespace: str | None = None,
    ) -> Any: ...


class ServicePort(Protocol):
    def service_list(
        self, namespace: str | None = None, page_no: int = 1, page_size: int = 100
    ) -> Any: ...

    def instance_list(
        self,
        service_name: str,
        namespace: str | None = None,
        group: str = "DEFAULT_GROUP",
    ) -> Any: ...

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
    ) -> Any: ...

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
    ) -> Any: ...

    def instance_deregister(
        self,
        service_name: str,
        ip: str,
        port: int,
        namespace: str | None = None,
        group: str = "DEFAULT_GROUP",
        cluster: str = "DEFAULT",
        ephemeral: bool = True,
    ) -> Any: ...


class NamespacePort(Protocol):
    def namespace_list(self) -> Any: ...
