# -*- coding: utf-8 -*-
"""Auth / doctor / whoami CLI handlers."""
from __future__ import annotations

import argparse
import getpass
import json
import sys
from typing import Any

from nacos_cli.config import load_settings
from nacos_cli.domain.auth_uc import AuthUseCase
from nacos_cli.infrastructure.credentials import CredentialStore, StoredCredentials, mask_token
from nacos_cli.infrastructure.http_client import NacosClient
from nacos_cli.infrastructure.web_login import run_web_login


def _print(data: Any) -> None:
    print(json.dumps(data, ensure_ascii=False, indent=2))


def _persist_login(result: dict[str, Any], *, via: str, username: str, server: str) -> dict[str, Any]:
    store = CredentialStore()
    mode = str(result.get("auth_mode") or "none")
    token = ""
    # token lives on client after auth_login; caller may pass access_token
    if mode == "token":
        token = str(result.get("access_token") or "")
    store.save(
        StoredCredentials(
            access_token=token if mode == "token" else "",
            username=username,
            server=server,
            auth_mode=mode,
            via=via,
        )
    )
    out = {
        "ok": True,
        "auth_mode": mode,
        "has_token": bool(token),
        "token_masked": mask_token(token) if token else "",
        "username": username,
        "server": server,
        "credentials_path": str(store.path),
        "via": via,
    }
    if "ttl" in result:
        out["ttl"] = result.get("ttl")
    return out


def cmd_auth(args: argparse.Namespace) -> int:
    action = args.action
    if action == "logout":
        store = CredentialStore()
        cleared = store.clear()
        _print({"ok": True, "cleared": cleared, "credentials_path": str(store.path)})
        return 0

    if action == "status":
        store = CredentialStore()
        st = store.status()
        settings = load_settings(require=False)
        st["server_env"] = settings.server if settings else ""
        st["ok"] = True
        if getattr(args, "probe", False):
            if not settings or not settings.server:
                st["probe"] = {"ok": False, "error": "NACOS_SERVER not set"}
            else:
                c = NacosClient(settings)
                c.login()
                st["probe"] = c.auth_probe()
                st["auth_mode"] = getattr(c, "_auth_mode", st.get("auth_mode"))
                st["has_token"] = bool(getattr(c, "_token", None))
                if c._token:
                    st["token_masked"] = mask_token(c._token)
        _print(st)
        return 0

    if action == "login":
        settings = load_settings(require=True)
        if settings is None:
            return 2
        if getattr(args, "dry_run", False):
            _print(
                {
                    "ok": True,
                    "dry_run": True,
                    "would": {
                        "action": "auth.login",
                        "web": bool(args.web),
                        "server": settings.server,
                    },
                }
            )
            return 0

        c = NacosClient(settings)
        uc = AuthUseCase(c)

        if args.web:
            def do_login(user: str, password: str) -> dict[str, Any]:
                result = uc.login(user, password)
                # capture token from client for persistence
                if result.get("auth_mode") == "token" and c._token:
                    result = {**result, "access_token": c._token}
                persisted = _persist_login(
                    result, via="web", username=user, server=settings.server
                )
                # return auth_mode for HTML; token not shown in page beyond mode
                return {"auth_mode": persisted["auth_mode"], "has_token": persisted["has_token"]}

            try:
                raw = run_web_login(
                    server_hint=settings.base_url,
                    prefill_user=args.username or settings.username or "nacos",
                    do_login=do_login,
                )
            except Exception as e:  # noqa: BLE001
                print(f"web login failed: {e}", file=sys.stderr)
                return 1
            # reload status for stdout (no secrets)
            st = CredentialStore().status()
            st["ok"] = True
            st["via"] = "web"
            st["auth_mode"] = raw.get("auth_mode")
            _print(st)
            return 0

        user = args.username or settings.username or ""
        password = args.password
        if password is None:
            password = settings.password
        if not password and sys.stdin.isatty():
            try:
                password = getpass.getpass("Nacos password (empty if auth disabled): ")
            except Exception:
                password = ""
        if not user:
            print("username required (--username or NACOS_USERNAME)", file=sys.stderr)
            return 1
        try:
            result = uc.login(user, password or "")
        except Exception as e:  # noqa: BLE001
            print(f"login failed: {e}", file=sys.stderr)
            return 1
        if result.get("auth_mode") == "token" and c._token:
            result = {**result, "access_token": c._token}
        out = _persist_login(result, via="password", username=user, server=settings.server)
        _print(out)
        return 0

    print(f"unknown auth action: {action}", file=sys.stderr)
    return 1


def cmd_whoami(args: argparse.Namespace) -> int:
    args.action = "status"
    return cmd_auth(args)


def cmd_doctor(args: argparse.Namespace) -> int:
    settings = load_settings(require=False)
    store = CredentialStore()
    report: dict[str, Any] = {
        "ok": True,
        "version": __import__("nacos_cli").__version__,
        "server": settings.server if settings else "",
        "context_path": settings.context_path if settings else "",
        "username_env": settings.username if settings else "",
        "has_password_env": bool(settings.password) if settings else False,
        "credentials": store.status(),
    }
    if not settings or not settings.server:
        report["ok"] = False
        report["error"] = "NACOS_SERVER not set; copy .env.example to .env"
        _print(report)
        return 2
    c = NacosClient(settings)
    try:
        login_meta = c.login()
        report["login"] = {
            k: v for k, v in login_meta.items() if k != "access_token"
        }
        probe = c.auth_probe()
        report["probe"] = probe
        if not probe.get("ok"):
            report["ok"] = False
    except Exception as e:  # noqa: BLE001
        report["ok"] = False
        report["error"] = str(e)
    _print(report)
    return 0 if report.get("ok") else 1
