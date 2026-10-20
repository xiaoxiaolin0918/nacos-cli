# -*- coding: utf-8 -*-
"""Auth domain use cases (no HTTP)."""
from __future__ import annotations

from typing import Any, Protocol


class AuthPort(Protocol):
    def auth_login(self, username: str, password: str) -> dict[str, Any]: ...

    def auth_apply_token(self, access_token: str | None, auth_mode: str) -> None: ...

    def auth_probe(self) -> dict[str, Any]: ...


class AuthUseCase:
    def __init__(self, port: AuthPort) -> None:
        self._port = port

    def login(self, username: str, password: str) -> dict[str, Any]:
        if not username:
            raise ValueError("username is required")
        # password may be empty when auth disabled; still call port
        return self._port.auth_login(username, password)

    def probe(self) -> dict[str, Any]:
        return self._port.auth_probe()
