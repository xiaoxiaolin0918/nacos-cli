# -*- coding: utf-8 -*-
"""Infrastructure: NacosClient against mocked requests.Session (no real network)."""
from __future__ import annotations

from unittest.mock import MagicMock, patch

import pytest

from nacos_cli.config import Settings
from nacos_cli.infrastructure.http_client import NacosClient


def _settings() -> Settings:
    return Settings(
        server="http://127.0.0.1:8848",
        username="nacos",
        password="secret",
        namespace="",
        context_path="/nacos",
    )


def _json_resp(payload, status=200):
    r = MagicMock()
    r.status_code = status
    r.headers = {"Content-Type": "application/json"}
    r.json.return_value = payload
    r.text = str(payload)
    r.raise_for_status = MagicMock()
    return r


def _text_resp(text, status=200):
    r = MagicMock()
    r.status_code = status
    r.headers = {"Content-Type": "text/plain"}
    r.text = text
    r.raise_for_status = MagicMock()
    return r


@pytest.fixture
def client():
    c = NacosClient(_settings())
    c.session = MagicMock()
    return c


def test_login_skips_auth_disabled_token(client):
    client.session.post.return_value = _json_resp({"accessToken": "AUTH_DISABLED"})
    client.login()
    assert client._token is None


def test_login_keeps_real_token(client):
    client.session.post.return_value = _json_resp({"accessToken": "tok-123"})
    client.login()
    assert client._token == "tok-123"


def test_config_get(client):
    client.session.get.return_value = _text_resp("hello: world")
    assert client.config_get("app.yaml", "DEFAULT_GROUP") == "hello: world"
    args, kwargs = client.session.get.call_args
    assert args[0].endswith("/v1/cs/configs")
    assert kwargs["params"]["dataId"] == "app.yaml"


def test_config_list_sends_empty_data_id(client):
    client.session.get.return_value = _json_resp({"totalCount": 0, "pageItems": []})
    client.config_list("", "DEFAULT_GROUP")
    params = client.session.get.call_args.kwargs["params"]
    assert "dataId" in params
    assert params["dataId"] == ""


def test_config_publish(client):
    client.session.post.return_value = _text_resp("true")
    assert client.config_publish("x.yaml", "DEFAULT_GROUP", "body") == "true"
    data = client.session.post.call_args.kwargs["data"]
    assert data["content"] == "body"


def test_config_history(client):
    client.session.get.return_value = _json_resp({"totalCount": 1, "pageItems": [{"id": "9"}]})
    out = client.config_history("x.yaml", "DEFAULT_GROUP")
    assert out["pageItems"][0]["id"] == "9"


def test_instance_list(client):
    client.session.get.return_value = _json_resp({"hosts": [{"ip": "1.1.1.1", "port": 80}]})
    out = client.instance_list("svc")
    assert out["hosts"][0]["ip"] == "1.1.1.1"


def test_instance_up_down(client):
    # Nacos up/down: enabled must be query params (body is ignored by server).
    client.session.put.return_value = _text_resp("ok")
    assert client.instance_enabled("svc", "1.1.1.1", 80, False) == "ok"
    params = client.session.put.call_args.kwargs["params"]
    assert params["enabled"] == "false"
    assert params["serviceName"] == "svc"
    assert client.instance_enabled("svc", "1.1.1.1", 80, True) == "ok"
    params = client.session.put.call_args.kwargs["params"]
    assert params["enabled"] == "true"
