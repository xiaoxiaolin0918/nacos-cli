# -*- coding: utf-8 -*-
"""TDD: ServiceUseCase instances / up / down core paths."""
from __future__ import annotations

import pytest

from nacos_cli.domain import ServiceUseCase
from tests.fakes import FakeServicePort


@pytest.fixture
def uc() -> tuple[ServiceUseCase, FakeServicePort]:
    port = FakeServicePort()
    return ServiceUseCase(port), port


def test_instances_requires_service(uc):
    use, _ = uc
    with pytest.raises(ValueError, match="service_name"):
        use.instances("")


def test_instances_lists_hosts(uc):
    use, port = uc
    port.instance_register("svc-a", "10.0.0.1", 8080, namespace="")
    result = use.instances("svc-a", namespace="")
    assert len(result["hosts"]) == 1
    assert result["hosts"][0]["ip"] == "10.0.0.1"


def test_down_then_up(uc):
    use, port = uc
    port.instance_register("svc-a", "10.0.0.1", 8080, namespace="")
    assert use.down("svc-a", "10.0.0.1", 8080, namespace="") == "ok"
    hosts = use.instances("svc-a", namespace="")["hosts"]
    assert hosts[0]["enabled"] is False
    assert use.up("svc-a", "10.0.0.1", 8080, namespace="") == "ok"
    hosts = use.instances("svc-a", namespace="")["hosts"]
    assert hosts[0]["enabled"] is True
    enabled_calls = [c for c in port.calls if c[0] == "instance_enabled"]
    assert enabled_calls[0][1]["enabled"] is False
    assert enabled_calls[1][1]["enabled"] is True


def test_up_requires_ip_port(uc):
    use, _ = uc
    with pytest.raises(ValueError, match="ip"):
        use.up("svc-a", "", 8080)
    with pytest.raises(ValueError, match="port"):
        use.up("svc-a", "10.0.0.1", None)  # type: ignore[arg-type]
