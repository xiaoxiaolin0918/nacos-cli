# -*- coding: utf-8 -*-
"""TDD: config publish defaults to create-only; --overwrite updates."""
from __future__ import annotations

import pytest

from nacos_cli.domain import ConfigUseCase
from nacos_cli.interface.cli import build_parser, _resolve_create_only
from tests.fakes import FakeConfigPort


def test_default_create_only_allows_new():
    port = FakeConfigPort()
    uc = ConfigUseCase(port)
    assert uc.publish("new.yaml", "DEFAULT_GROUP", "v1") is True
    assert port.store[("new.yaml", "DEFAULT_GROUP", "")] == "v1"


def test_default_create_only_rejects_existing():
    port = FakeConfigPort()
    uc = ConfigUseCase(port)
    uc.publish("x.yaml", "DEFAULT_GROUP", "v1")
    with pytest.raises(FileExistsError, match="already exists"):
        uc.publish("x.yaml", "DEFAULT_GROUP", "v2")
    assert port.store[("x.yaml", "DEFAULT_GROUP", "")] == "v1"


def test_create_only_explicit_still_rejects():
    port = FakeConfigPort()
    uc = ConfigUseCase(port)
    uc.publish("x.yaml", "DEFAULT_GROUP", "v1")
    with pytest.raises(FileExistsError, match="already exists"):
        uc.publish("x.yaml", "DEFAULT_GROUP", "v2", create_only=True)


def test_overwrite_via_create_only_false():
    port = FakeConfigPort()
    uc = ConfigUseCase(port)
    uc.publish("x.yaml", "DEFAULT_GROUP", "v1")
    assert uc.publish("x.yaml", "DEFAULT_GROUP", "v2", create_only=False) is True
    assert port.store[("x.yaml", "DEFAULT_GROUP", "")] == "v2"


def test_parser_default_create_only_true_via_resolve():
    p = build_parser()
    args = p.parse_args(["config", "publish", "--data-id", "x", "--content", "a"])
    assert args.create_only is False  # flag not passed
    assert args.overwrite is False
    assert _resolve_create_only(args) is True


def test_parser_overwrite_disables_create_only():
    p = build_parser()
    args = p.parse_args(
        ["config", "publish", "--data-id", "x", "--content", "a", "--overwrite"]
    )
    assert args.overwrite is True
    assert _resolve_create_only(args) is False


def test_parser_create_only_and_overwrite_conflict():
    p = build_parser()
    args = p.parse_args(
        [
            "config",
            "publish",
            "--data-id",
            "x",
            "--content",
            "a",
            "--create-only",
            "--overwrite",
        ]
    )
    with pytest.raises(ValueError, match="either"):
        _resolve_create_only(args)
