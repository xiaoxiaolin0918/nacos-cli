# -*- coding: utf-8 -*-
from __future__ import annotations

from unittest.mock import MagicMock, patch

from nacos_cli.config import Settings
from nacos_cli.domain.auth_uc import AuthUseCase
from nacos_cli.infrastructure.credentials import CredentialStore, StoredCredentials
from nacos_cli.infrastructure.http_client import NacosClient


def _settings() -> Settings:
    return Settings(
        server="http://127.0.0.1:8848",
        username="nacos",
        password="",
        namespace="",
        context_path="/nacos",
    )


def test_auth_login_token(monkeypatch, tmp_path):
    monkeypatch.setenv("NACOS_CREDENTIALS_PATH", str(tmp_path / "c.json"))
    c = NacosClient(_settings())
    resp = MagicMock()
    resp.status_code = 200
    resp.headers = {"Content-Type": "application/json"}
    resp.json.return_value = {"accessToken": "abcXYZTOKEN123456", "tokenTtl": 18000}
    resp.text = ""
    with patch.object(c.session, "post", return_value=resp) as post:
        out = c.auth_login("nacos", "secret")
    post.assert_called_once()
    assert out["auth_mode"] == "token"
    assert out["has_token"] is True
    assert c._token == "abcXYZTOKEN123456"
    assert "secret" not in str(out)


def test_auth_login_disabled(monkeypatch, tmp_path):
    monkeypatch.setenv("NACOS_CREDENTIALS_PATH", str(tmp_path / "c.json"))
    c = NacosClient(_settings())
    resp = MagicMock()
    resp.status_code = 200
    resp.headers = {"Content-Type": "application/json"}
    resp.json.return_value = {"accessToken": "AUTH_DISABLED"}
    resp.text = ""
    with patch.object(c.session, "post", return_value=resp):
        out = c.auth_login("nacos", "")
    assert out["auth_mode"] == "disabled"
    assert out["has_token"] is False
    assert c._token is None


def test_login_uses_credentials_file(monkeypatch, tmp_path):
    path = tmp_path / "c.json"
    monkeypatch.setenv("NACOS_CREDENTIALS_PATH", str(path))
    CredentialStore(path).save(
        StoredCredentials(
            access_token="fileTOKEN99999999",
            username="nacos",
            server="http://127.0.0.1:8848",
            auth_mode="token",
            via="web",
        )
    )
    c = NacosClient(_settings())
    out = c.login()
    assert out["auth_mode"] == "token"
    assert out.get("token_source") == "credentials_file"
    assert c._token == "fileTOKEN99999999"


def test_auth_uc_requires_username():
    port = MagicMock()
    uc = AuthUseCase(port)
    try:
        uc.login("", "x")
        assert False, "expected ValueError"
    except ValueError as e:
        assert "username" in str(e)
    port.auth_login.assert_not_called()


def test_auth_uc_delegates():
    port = MagicMock()
    port.auth_login.return_value = {"auth_mode": "token", "has_token": True}
    uc = AuthUseCase(port)
    assert uc.login("u", "p")["has_token"] is True
    port.auth_login.assert_called_once_with("u", "p")
