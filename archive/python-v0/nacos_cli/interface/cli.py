# -*- coding: utf-8 -*-
"""CLI adapter: argparse -> domain use cases -> infrastructure HTTP client."""
from __future__ import annotations

import argparse
import json
import sys
from pathlib import Path

from nacos_cli import __version__
from nacos_cli.config import load_settings
from nacos_cli.domain import ConfigUseCase, NamespaceUseCase, ServiceUseCase
from nacos_cli.infrastructure import NacosClient
from nacos_cli.interface.write_gate import check_write_gate
from nacos_cli.interface.auth_cmds import cmd_auth, cmd_doctor, cmd_whoami
from nacos_cli.interface.skills_cmds import cmd_skills


def _parse_bool(value: str) -> bool:
    v = str(value).strip().lower()
    if v in ("1", "true", "yes", "y", "on"):
        return True
    if v in ("0", "false", "no", "n", "off"):
        return False
    raise argparse.ArgumentTypeError(f"expected bool, got {value!r}")


def _resolve_ephemeral(args: argparse.Namespace) -> bool:
    """Default ephemeral=True; --persistent or --ephemeral false => persistent."""
    if getattr(args, "persistent", False):
        return False
    return bool(getattr(args, "ephemeral", True))




def _resolve_create_only(args: argparse.Namespace) -> bool:
    """Default create-only (safe). --overwrite allows update of existing dataId."""
    if getattr(args, "overwrite", False):
        if getattr(args, "create_only", False):
            raise ValueError("use either --create-only or --overwrite, not both")
        return False
    # default True; --create-only is affirmative no-op for docs/compat
    return True

def _print(data) -> None:
    if isinstance(data, (dict, list)):
        print(json.dumps(data, ensure_ascii=False, indent=2))
    else:
        print(data)


def _client() -> NacosClient | None:
    settings = load_settings(require=True)
    if settings is None:
        return None
    c = NacosClient(settings)
    try:
        c.login()
    except Exception as e:
        print(f"login failed: {e}", file=sys.stderr)
        return None
    return c


def _add_write_flags(p: argparse.ArgumentParser) -> None:
    p.add_argument(
        "--yes",
        "-y",
        action="store_true",
        help="confirm high-risk write (required unless --dry-run)",
    )
    p.add_argument(
        "--dry-run",
        action="store_true",
        help="print the would-be action and do not call Nacos",
    )


def cmd_config(args: argparse.Namespace) -> int:
    ns = args.namespace
    try:
        if args.action == "publish":
            if args.file:
                content = Path(args.file).read_text(encoding="utf-8")
            elif args.content is not None:
                content = args.content
            else:
                print("publish needs --file or --content", file=sys.stderr)
                return 1
            try:
                create_only = _resolve_create_only(args)
            except ValueError as e:
                print(f"failed: {e}", file=sys.stderr)
                return 1
            gate = check_write_gate(
                yes=args.yes,
                dry_run=args.dry_run,
                action="config.publish",
                detail={
                    "data_id": args.data_id,
                    "group": args.group,
                    "namespace": ns,
                    "type": args.type,
                    "create_only": create_only,
                    "overwrite": not create_only,
                    "content_bytes": len(content.encode("utf-8")),
                },
            )
            if gate is not None:
                return gate
            c = _client()
            if c is None:
                return 2
            _print(
                ConfigUseCase(c).publish(
                    args.data_id,
                    args.group,
                    content,
                    ns,
                    args.type,
                    create_only=create_only,
                )
            )
            return 0
        if args.action == "rollback":
            if not args.nid:
                print("rollback needs --nid (from config history)", file=sys.stderr)
                return 1
            gate = check_write_gate(
                yes=args.yes,
                dry_run=args.dry_run,
                action="config.rollback",
                detail={
                    "data_id": args.data_id,
                    "group": args.group,
                    "namespace": ns,
                    "nid": args.nid,
                },
            )
            if gate is not None:
                return gate
            c = _client()
            if c is None:
                return 2
            _print(ConfigUseCase(c).rollback(args.data_id, args.group, args.nid, ns))
            return 0

        c = _client()
        if c is None:
            return 2
        uc = ConfigUseCase(c)
        if args.action == "get":
            _print(uc.get(args.data_id, args.group, ns))
        elif args.action == "list":
            _print(
                uc.list(
                    args.data_id or "", args.group or "", ns, args.page, args.page_size
                )
            )
        elif args.action == "history":
            _print(uc.history(args.data_id, args.group, ns, args.page, args.page_size))
        else:
            print(f"unknown action: {args.action}", file=sys.stderr)
            return 1
    except Exception as e:
        print(f"failed: {e}", file=sys.stderr)
        return 1
    return 0


def cmd_service(args: argparse.Namespace) -> int:
    ns = args.namespace
    ephemeral = _resolve_ephemeral(args)
    write_actions = {"register", "deregister", "up", "down"}
    try:
        if args.action in write_actions:
            if not args.service or not args.ip or args.port is None:
                print(f"{args.action} needs --service --ip --port", file=sys.stderr)
                return 1
            detail = {
                "service": args.service,
                "ip": args.ip,
                "port": args.port,
                "namespace": ns,
                "group": args.group,
                "cluster": args.cluster,
                "ephemeral": ephemeral,
            }
            gate = check_write_gate(
                yes=args.yes,
                dry_run=args.dry_run,
                action=f"service.{args.action}",
                detail=detail,
            )
            if gate is not None:
                return gate
            c = _client()
            if c is None:
                return 2
            uc = ServiceUseCase(c)
            if args.action == "register":
                _print(
                    uc.register(
                        args.service,
                        args.ip,
                        args.port,
                        ns,
                        args.group,
                        args.cluster,
                        ephemeral=ephemeral,
                    )
                )
            elif args.action == "deregister":
                _print(
                    uc.deregister(
                        args.service,
                        args.ip,
                        args.port,
                        ns,
                        args.group,
                        args.cluster,
                        ephemeral=ephemeral,
                    )
                )
            elif args.action == "up":
                _print(
                    uc.up(
                        args.service,
                        args.ip,
                        args.port,
                        ns,
                        args.group,
                        args.cluster,
                        ephemeral=ephemeral,
                    )
                )
            else:
                _print(
                    uc.down(
                        args.service,
                        args.ip,
                        args.port,
                        ns,
                        args.group,
                        args.cluster,
                        ephemeral=ephemeral,
                    )
                )
            return 0

        c = _client()
        if c is None:
            return 2
        uc = ServiceUseCase(c)
        if args.action == "list":
            _print(uc.list(ns, args.page, args.page_size))
        elif args.action == "instances":
            _print(uc.instances(args.service, ns, args.group))
        else:
            print(f"unknown action: {args.action}", file=sys.stderr)
            return 1
    except Exception as e:
        print(f"failed: {e}", file=sys.stderr)
        return 1
    return 0


def cmd_namespace(args: argparse.Namespace) -> int:
    c = _client()
    if c is None:
        return 2
    uc = NamespaceUseCase(c)
    try:
        if args.action == "list":
            _print(uc.list())
        else:
            print(f"unknown action: {args.action}", file=sys.stderr)
            return 1
    except Exception as e:
        print(f"failed: {e}", file=sys.stderr)
        return 1
    return 0


def build_parser() -> argparse.ArgumentParser:
    p = argparse.ArgumentParser(
        prog="nacos",
        description="Nacos ops CLI (config / service instance / namespace)",
    )
    p.add_argument("--version", action="version", version=f"nacos-cli {__version__}")
    sub = p.add_subparsers(dest="command", required=True)

    pc = sub.add_parser("config", help="config: get/list/publish/history/rollback")
    pc.add_argument(
        "action",
        choices=["get", "list", "publish", "history", "rollback"],
        help="action",
    )
    pc.add_argument("--data-id", default="", help="dataId")
    pc.add_argument("--group", default="DEFAULT_GROUP", help="group, default DEFAULT_GROUP")
    pc.add_argument("--namespace", default=None, help="override NAMESPACE from .env")
    pc.add_argument("--file", help="publish: read content from file")
    pc.add_argument("--content", help="publish: content string")
    pc.add_argument("--type", default="text", help="publish: config type")
    pc.add_argument(
        "--create-only",
        action="store_true",
        help="publish: create only (DEFAULT). Fail if dataId already exists",
    )
    pc.add_argument(
        "--overwrite",
        action="store_true",
        help="publish: allow overwrite of existing dataId (explicit; not default)",
    )
    pc.add_argument("--nid", help="rollback: history id")
    pc.add_argument("--page", type=int, default=1)
    pc.add_argument("--page-size", type=int, default=100)
    _add_write_flags(pc)
    pc.set_defaults(func=cmd_config)

    ps = sub.add_parser(
        "service", help="service: list/instances/register/deregister/up/down"
    )
    ps.add_argument(
        "action",
        choices=["list", "instances", "register", "deregister", "up", "down"],
        help="action",
    )
    ps.add_argument("--service", help="service name")
    ps.add_argument("--ip", help="instance IP")
    ps.add_argument("--port", type=int, help="instance port")
    ps.add_argument("--group", default="DEFAULT_GROUP")
    ps.add_argument("--cluster", default="DEFAULT")
    ps.add_argument(
        "--ephemeral",
        type=_parse_bool,
        nargs="?",
        const=True,
        default=True,
        help="ephemeral instance (default true). Use --ephemeral false for persistent.",
    )
    ps.add_argument(
        "--persistent",
        action="store_true",
        help="persistent instance (alias for --ephemeral false)",
    )
    ps.add_argument("--namespace", default=None)
    ps.add_argument("--page", type=int, default=1)
    ps.add_argument("--page-size", type=int, default=100)
    _add_write_flags(ps)
    ps.set_defaults(func=cmd_service)

    pn = sub.add_parser("namespace", help="namespace")
    pn.add_argument("action", choices=["list"], help="action")
    pn.set_defaults(func=cmd_namespace)

    pa = sub.add_parser("auth", help="login / logout / status (Open API token or AUTH_DISABLED)")
    pa.add_argument("action", choices=["login", "logout", "status"], help="action")
    pa.add_argument("--username", default=None, help="login username (else NACOS_USERNAME)")
    pa.add_argument("--password", nargs="?", const="", default=None, help="login password (else env / prompt; omit value if AUTH_DISABLED)")
    pa.add_argument("--web", action="store_true", help="open local 127.0.0.1 browser form (plumelog-style)")
    pa.add_argument("--probe", action="store_true", help="status: probe namespace list")
    pa.add_argument("--dry-run", action="store_true", help="login: preview only")
    pa.set_defaults(func=cmd_auth)

    pw = sub.add_parser("whoami", help="alias of auth status (never prints raw token)")
    pw.add_argument("--probe", action="store_true", help="probe namespace list")
    pw.set_defaults(func=cmd_whoami, action="status")

    pd = sub.add_parser("doctor", help="read-only connectivity + auth_mode check")
    pd.set_defaults(func=cmd_doctor)

    psk = sub.add_parser("skills", help="list/read bundled agent skills/")
    psk.add_argument("action", choices=["list", "read", "path"], help="action")
    psk.add_argument("name", nargs="?", default=None, help="skill name for read/path")
    psk.set_defaults(func=cmd_skills)

    return p


def main(argv: list[str] | None = None) -> int:
    parser = build_parser()
    args = parser.parse_args(argv)
    return args.func(args)


if __name__ == "__main__":
    raise SystemExit(main())
