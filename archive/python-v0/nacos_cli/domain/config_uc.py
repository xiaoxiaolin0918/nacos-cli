# -*- coding: utf-8 -*-
"""Config domain use cases."""
from __future__ import annotations

from typing import Any

from .ports import ConfigPort


class ConfigUseCase:
    def __init__(self, port: ConfigPort) -> None:
        self._port = port

    def get(self, data_id: str, group: str = "DEFAULT_GROUP", namespace: str | None = None) -> Any:
        if not data_id:
            raise ValueError("data_id is required")
        return self._port.config_get(data_id, group, namespace)

    def list(
        self,
        data_id: str = "",
        group: str = "",
        namespace: str | None = None,
        page_no: int = 1,
        page_size: int = 100,
    ) -> Any:
        return self._port.config_list(data_id, group, namespace, page_no, page_size)

    def publish(
        self,
        data_id: str,
        group: str,
        content: str,
        namespace: str | None = None,
        type_: str = "text",
        create_only: bool = True,
    ) -> Any:
        if not data_id:
            raise ValueError("data_id is required")
        if content is None:
            raise ValueError("content is required")
        if create_only and self._exists(data_id, group, namespace):
            raise FileExistsError(
                f"config already exists: dataId={data_id} group={group} namespace={namespace or ''}"
            )
        return self._port.config_publish(data_id, group, content, namespace, type_)

    def history(
        self,
        data_id: str,
        group: str = "DEFAULT_GROUP",
        namespace: str | None = None,
        page_no: int = 1,
        page_size: int = 20,
    ) -> Any:
        if not data_id:
            raise ValueError("data_id is required")
        return self._port.config_history(data_id, group, namespace, page_no, page_size)

    def rollback(
        self,
        data_id: str,
        group: str,
        nid: str,
        namespace: str | None = None,
    ) -> Any:
        if not data_id:
            raise ValueError("data_id is required")
        if not nid:
            raise ValueError("nid is required")
        return self._port.config_rollback(data_id, group, nid, namespace)

    def _exists(self, data_id: str, group: str, namespace: str | None) -> bool:
        try:
            self._port.config_get(data_id, group, namespace)
            return True
        except KeyError:
            return False
        except Exception as e:
            status = getattr(getattr(e, "response", None), "status_code", None)
            msg = str(e).lower()
            if status == 404 or "404" in msg or "not found" in msg:
                return False
            raise
