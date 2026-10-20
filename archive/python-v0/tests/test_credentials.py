# -*- coding: utf-8 -*-
from __future__ import annotations

from nacos_cli.infrastructure.credentials import (
    CredentialStore,
    StoredCredentials,
    mask_token,
)


def test_mask_token():
    assert mask_token("") == ""
    assert mask_token("short") == "***"
    assert mask_token("abcdefghijklmnop") == "abcd...mnop"
    assert mask_token("AUTH_DISABLED") == "AUTH_DISABLED"


def test_store_roundtrip(tmp_path, monkeypatch):
    path = tmp_path / "cred.json"
    monkeypatch.setenv("NACOS_CREDENTIALS_PATH", str(path))
    store = CredentialStore(path)
    assert store.load().auth_mode == "none"
    store.save(
        StoredCredentials(
            access_token="tok_ABCDEFGH12345678",
            username="nacos",
            server="http://127.0.0.1:8848",
            auth_mode="token",
            via="password",
        )
    )
    loaded = store.load()
    assert loaded.username == "nacos"
    assert loaded.access_token.startswith("tok_")
    st = store.status()
    assert st["has_token"] is True
    assert "ABCDEFGH12345678" not in st["token_masked"]
    assert st["token_masked"].startswith("tok_")
    assert store.clear() is True
    assert store.load().auth_mode == "none"
    assert store.status()["has_token"] is False


def test_status_never_exposes_raw_token(tmp_path):
    path = tmp_path / "c.json"
    store = CredentialStore(path)
    store.save(
        StoredCredentials(
            access_token="supersecrettokenvalue",
            username="u",
            server="http://x",
            auth_mode="token",
        )
    )
    raw = str(store.status())
    assert "supersecrettokenvalue" not in raw
