# -*- coding: utf-8 -*-
"""Namespace domain use cases."""
from __future__ import annotations

from typing import Any

from .ports import NamespacePort


class NamespaceUseCase:
    def __init__(self, port: NamespacePort) -> None:
        self._port = port

    def list(self) -> Any:
        return self._port.namespace_list()
