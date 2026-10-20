# -*- coding: utf-8 -*-
"""TDD: --dry-run / --yes write gates and CLI defaults."""
from __future__ import annotations

import json

from nacos_cli.interface.write_gate import CONFIRMATION_REQUIRED_EXIT, check_write_gate
from nacos_cli.interface.cli import build_parser


def test_dry_run_prints_and_skips(capsys):
    code = check_write_gate(
        yes=False,
        dry_run=True,
        action="config.publish",
        detail={"data_id": "x.yaml"},
    )
    assert code == 0
    out = json.loads(capsys.readouterr().out)
    assert out["dry_run"] is True
    assert out["would"]["action"] == "config.publish"
    assert out["would"]["data_id"] == "x.yaml"


def test_missing_yes_refuses(capsys):
    code = check_write_gate(yes=False, dry_run=False, action="service.down")
    assert code == CONFIRMATION_REQUIRED_EXIT
    err = capsys.readouterr().err
    assert "--yes" in err
    assert "service.down" in err


def test_yes_allows_proceed():
    assert check_write_gate(yes=True, dry_run=False, action="service.up") is None


def test_dry_run_wins_even_with_yes(capsys):
    code = check_write_gate(yes=True, dry_run=True, action="config.rollback")
    assert code == 0
    assert json.loads(capsys.readouterr().out)["dry_run"] is True


def test_parser_register_defaults_ephemeral_true():
    p = build_parser()
    args = p.parse_args(
        ["service", "register", "--service", "s", "--ip", "1.1.1.1", "--port", "80"]
    )
    assert args.ephemeral is True
    assert args.persistent is False


def test_parser_persistent_flag():
    p = build_parser()
    args = p.parse_args(
        [
            "service",
            "register",
            "--service",
            "s",
            "--ip",
            "1.1.1.1",
            "--port",
            "80",
            "--persistent",
        ]
    )
    assert args.persistent is True


def test_parser_ephemeral_false_string():
    p = build_parser()
    args = p.parse_args(
        [
            "service",
            "register",
            "--service",
            "s",
            "--ip",
            "1.1.1.1",
            "--port",
            "80",
            "--ephemeral",
            "false",
        ]
    )
    assert args.ephemeral is False


def test_parser_write_flags_on_config_and_service():
    p = build_parser()
    c = p.parse_args(
        ["config", "publish", "--data-id", "x", "--content", "a", "--yes", "--dry-run"]
    )
    assert c.yes is True and c.dry_run is True and c.create_only is False
    assert getattr(c, "overwrite", False) is False
    c2 = p.parse_args(
        ["config", "publish", "--data-id", "x", "--content", "a", "--create-only"]
    )
    assert c2.create_only is True
    c3 = p.parse_args(
        ["config", "publish", "--data-id", "x", "--content", "a", "--overwrite"]
    )
    assert c3.overwrite is True
    s = p.parse_args(
        ["service", "down", "--service", "s", "--ip", "1.1.1.1", "--port", "1", "-y"]
    )
    assert s.yes is True


def test_cli_dry_run_publish_no_login(monkeypatch, capsys):
    """--dry-run must not require .env / login."""
    def boom():
        raise AssertionError("must not connect on dry-run")

    monkeypatch.setattr("nacos_cli.interface.cli._client", boom)
    from nacos_cli.interface.cli import main

    code = main(
        [
            "config",
            "publish",
            "--data-id",
            "new-only.yaml",
            "--content",
            "k: v",
            "--dry-run",
        ]
    )
    assert code == 0
    out = capsys.readouterr().out
    assert '"dry_run": true' in out.replace(" ", "").lower() or '"dry_run": true' in out


def test_cli_publish_without_yes_exits_10(monkeypatch, capsys):
    def boom():
        raise AssertionError("must not connect without --yes")

    monkeypatch.setattr("nacos_cli.interface.cli._client", boom)
    from nacos_cli.interface.cli import main

    code = main(
        ["config", "publish", "--data-id", "x.yaml", "--content", "a"]
    )
    assert code == 10
