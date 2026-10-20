# -*- coding: utf-8 -*-
"""TDD: register defaults ephemeral; up/down pass ephemeral through."""
from __future__ import annotations

from nacos_cli.domain import ServiceUseCase
from tests.fakes import FakeServicePort


def test_register_default_ephemeral_true():
    port = FakeServicePort()
    uc = ServiceUseCase(port)
    uc.register("svc", "10.0.0.1", 8080, namespace="")
    assert port.instances[("svc", "", "DEFAULT_GROUP")][0]["ephemeral"] is True
    reg = [c for c in port.calls if c[0] == "instance_register"][0]
    assert reg[1]["ephemeral"] is True


def test_register_persistent_false():
    port = FakeServicePort()
    uc = ServiceUseCase(port)
    uc.register("svc", "10.0.0.1", 8080, namespace="", ephemeral=False)
    assert port.instances[("svc", "", "DEFAULT_GROUP")][0]["ephemeral"] is False


def test_up_down_pass_ephemeral():
    port = FakeServicePort()
    uc = ServiceUseCase(port)
    uc.register("svc", "10.0.0.1", 8080, namespace="", ephemeral=True)
    uc.down("svc", "10.0.0.1", 8080, namespace="", ephemeral=True)
    uc.up("svc", "10.0.0.1", 8080, namespace="", ephemeral=True)
    enabled = [c for c in port.calls if c[0] == "instance_enabled"]
    assert enabled[0][1]["ephemeral"] is True
    assert enabled[1][1]["ephemeral"] is True
