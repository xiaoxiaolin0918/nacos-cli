# -*- coding: utf-8 -*-
"""Local credential store for Nacos accessToken (never commit)."""
from __future__ import annotations

import json
import os
from dataclasses import dataclass
from pathlib import Path
from typing import Any


def default_credentials_path() -> Path:
    override = os.environ.get("NACOS_CREDENTIALS_PATH", "").strip()
    if override:
        return Path(override)
    appdata = os.environ.get("APPDATA", "").strip()
    if appdata:
        return Path(appdata) / "nacos-cli" / "credentials.json"
    xdg = os.environ.get("XDG_CONFIG_HOME", "").strip()
    if xdg:
        return Path(xdg) / "nacos-cli" / "credentials.json"
    return Path.home() / ".config" / "nacos-cli" / "credentials.json"


def mask_token(token: str | None) -> str:
    if not token:
        return ""
    t = str(token)
    if t in ("AUTH_DISABLED", "<redacted>"):
        return t
    if len(t) <= 8:
        return "***"
    return t[:4] + "..." + t[-4:]


@dataclass
class StoredCredentials:
    access_token: str = ""
    username: str = ""
    server: str = ""
    auth_mode: str = "none"  # token | disabled | none
    via: str = ""  # password | web | env | none

    def to_dict(self) -> dict[str, Any]:
        return {
            "access_token": self.access_token,
            "username": self.username,
            "server": self.server,
            "auth_mode": self.auth_mode,
            "via": self.via,
        }

    @classmethod
    def from_dict(cls, data: dict[str, Any] | None) -> "StoredCredentials":
        if not data:
            return cls()
        return cls(
            access_token=str(data.get("access_token") or ""),
            username=str(data.get("username") or ""),
            server=str(data.get("server") or ""),
            auth_mode=str(data.get("auth_mode") or "none"),
            via=str(data.get("via") or ""),
        )


class CredentialStore:
    def __init__(self, path: Path | None = None) -> None:
        self.path = path or default_credentials_path()

    def load(self) -> StoredCredentials:
        if not self.path.is_file():
            return StoredCredentials()
        try:
            data = json.loads(self.path.read_text(encoding="utf-8"))
        except (OSError, json.JSONDecodeError):
            return StoredCredentials()
        if not isinstance(data, dict):
            return StoredCredentials()
        return StoredCredentials.from_dict(data)

    def save(self, creds: StoredCredentials) -> Path:
        self.path.parent.mkdir(parents=True, exist_ok=True)
        self.path.write_text(
            json.dumps(creds.to_dict(), ensure_ascii=False, indent=2) + "\n",
            encoding="utf-8",
        )
        try:
            os.chmod(self.path, 0o600)
        except OSError:
            pass
        return self.path

    def clear(self) -> bool:
        if self.path.is_file():
            self.path.unlink()
            return True
        return False

    def status(self) -> dict[str, Any]:
        c = self.load()
        has = bool(c.access_token) and c.access_token != "AUTH_DISABLED"
        return {
            "credentials_path": str(self.path),
            "has_token": has,
            "token_masked": mask_token(c.access_token) if has else "",
            "username": c.username,
            "server": c.server,
            "auth_mode": c.auth_mode or "none",
            "via": c.via or "",
        }
