# -*- coding: utf-8 -*-
"""从 .env / 环境变量加载 Nacos 连接配置。"""
from __future__ import annotations

import os
from dataclasses import dataclass
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
ENV_FILE = ROOT / ".env"
EXAMPLE = ROOT / ".env.example"


def _load_dotenv(path: Path) -> None:
    if not path.is_file():
        return
    for raw in path.read_text(encoding="utf-8").splitlines():
        line = raw.strip()
        if not line or line.startswith("#") or "=" not in line:
            continue
        key, _, val = line.partition("=")
        key = key.strip()
        val = val.strip().strip('"').strip("'")
        if key and key not in os.environ:
            os.environ[key] = val


@dataclass
class Settings:
    server: str
    username: str
    password: str
    namespace: str
    context_path: str

    @property
    def base_url(self) -> str:
        return self.server.rstrip("/") + self.context_path.rstrip("/")


def load_settings(*, require: bool = True) -> Settings | None:
    _load_dotenv(ENV_FILE)
    server = os.environ.get("NACOS_SERVER", "").strip()
    username = os.environ.get("NACOS_USERNAME", "nacos").strip()
    password = os.environ.get("NACOS_PASSWORD", "").strip()
    namespace = os.environ.get("NACOS_NAMESPACE", "").strip()
    context_path = os.environ.get("NACOS_CONTEXT_PATH", "/nacos").strip() or "/nacos"
    if not server:
        if require:
            print(
                "未配置 NACOS_SERVER。\n"
                f"请复制模板并填写：\n"
                f"  copy {EXAMPLE} {ENV_FILE}\n"
                f"然后编辑 .env 中的 SERVER / USERNAME / PASSWORD / NAMESPACE。"
            )
            return None
        return Settings("", username, password, namespace, context_path)
    return Settings(server, username, password, namespace, context_path)
