# -*- coding: utf-8 -*-
"""Service / instance domain use cases."""
from __future__ import annotations

from typing import Any

from .ports import ServicePort


class ServiceUseCase:
    def __init__(self, port: ServicePort) -> None:
        self._port = port

    def list(
        self, namespace: str | None = None, page_no: int = 1, page_size: int = 100
    ) -> Any:
        return self._port.service_list(namespace, page_no, page_size)

    def instances(
        self,
        service_name: str,
        namespace: str | None = None,
        group: str = "DEFAULT_GROUP",
    ) -> Any:
        if not service_name:
            raise ValueError("service_name is required")
        return self._port.instance_list(service_name, namespace, group)

    def up(
        self,
        service_name: str,
        ip: str,
        port: int,
        namespace: str | None = None,
        group: str = "DEFAULT_GROUP",
        cluster: str = "DEFAULT",
        ephemeral: bool = True,
    ) -> Any:
        self._require_instance(service_name, ip, port)
        return self._port.instance_enabled(
            service_name, ip, port, True, namespace, group, cluster, ephemeral
        )

    def down(
        self,
        service_name: str,
        ip: str,
        port: int,
        namespace: str | None = None,
        group: str = "DEFAULT_GROUP",
        cluster: str = "DEFAULT",
        ephemeral: bool = True,
    ) -> Any:
        self._require_instance(service_name, ip, port)
        return self._port.instance_enabled(
            service_name, ip, port, False, namespace, group, cluster, ephemeral
        )

    def register(
        self,
        service_name: str,
        ip: str,
        port: int,
        namespace: str | None = None,
        group: str = "DEFAULT_GROUP",
        cluster: str = "DEFAULT",
        ephemeral: bool = True,
    ) -> Any:
        self._require_instance(service_name, ip, port)
        return self._port.instance_register(
            service_name, ip, port, namespace, group, cluster, ephemeral=ephemeral
        )

    def deregister(
        self,
        service_name: str,
        ip: str,
        port: int,
        namespace: str | None = None,
        group: str = "DEFAULT_GROUP",
        cluster: str = "DEFAULT",
        ephemeral: bool = True,
    ) -> Any:
        self._require_instance(service_name, ip, port)
        return self._port.instance_deregister(
            service_name, ip, port, namespace, group, cluster, ephemeral=ephemeral
        )

    @staticmethod
    def _require_instance(service_name: str, ip: str, port: int | None) -> None:
        if not service_name:
            raise ValueError("service_name is required")
        if not ip:
            raise ValueError("ip is required")
        if port is None:
            raise ValueError("port is required")
