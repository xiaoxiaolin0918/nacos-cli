# -*- coding: utf-8 -*-
"""TDD: ConfigUseCase core paths (fake port = mocked HTTP boundary)."""
from __future__ import annotations

import pytest

from nacos_cli.domain import ConfigUseCase
from tests.fakes import FakeConfigPort


@pytest.fixture
def uc() -> tuple[ConfigUseCase, FakeConfigPort]:
    port = FakeConfigPort()
    return ConfigUseCase(port), port


def test_get_requires_data_id(uc):
    use, _ = uc
    with pytest.raises(ValueError, match="data_id"):
        use.get("")


def test_get_returns_content(uc):
    use, port = uc
    port.config_publish("app.yaml", "DEFAULT_GROUP", "k: v", namespace="ns1")
    assert use.get("app.yaml", "DEFAULT_GROUP", "ns1") == "k: v"


def test_list_filters_and_pages(uc):
    use, port = uc
    port.config_publish("a.yaml", "DEFAULT_GROUP", "a", namespace="")
    port.config_publish("b.yaml", "DEFAULT_GROUP", "b", namespace="")
    port.config_publish("c.yaml", "OTHER", "c", namespace="")
    result = use.list(data_id="a.yaml", group="DEFAULT_GROUP", namespace="")
    assert result["totalCount"] == 1
    assert result["pageItems"][0]["dataId"] == "a.yaml"
    blur = use.list(data_id="a.ya", group="DEFAULT_GROUP", namespace="")
    assert blur["totalCount"] == 1


def test_publish_and_history(uc):
    use, port = uc
    assert use.publish("x.yaml", "DEFAULT_GROUP", "v1") is True
    assert use.publish("x.yaml", "DEFAULT_GROUP", "v2", create_only=False) is True
    hist = use.history("x.yaml", "DEFAULT_GROUP")
    assert hist["totalCount"] == 2
    assert hist["pageItems"][0]["content"] == "v2"
    assert port.store[("x.yaml", "DEFAULT_GROUP", "")] == "v2"


def test_publish_rejects_missing_content(uc):
    use, _ = uc
    with pytest.raises(ValueError, match="content"):
        use.publish("x.yaml", "DEFAULT_GROUP", None)  # type: ignore[arg-type]


def test_history_requires_data_id(uc):
    use, _ = uc
    with pytest.raises(ValueError, match="data_id"):
        use.history("")
