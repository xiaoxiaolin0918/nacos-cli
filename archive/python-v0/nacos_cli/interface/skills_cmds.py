# -*- coding: utf-8 -*-
"""skills list / read / path commands."""
from __future__ import annotations

import argparse
import json
import sys

from nacos_cli.infrastructure.skills_fs import find_skills_root, list_skills, read_skill


def _print(data) -> None:
    print(json.dumps(data, ensure_ascii=False, indent=2))


def cmd_skills(args: argparse.Namespace) -> int:
    action = args.action
    try:
        if action == "list":
            _print(list_skills())
            return 0
        if action == "read":
            if not args.name:
                print("skills read needs NAME", file=sys.stderr)
                return 1
            _print(read_skill(args.name))
            return 0
        if action == "path":
            root = find_skills_root()
            if root is None:
                _print({"path": None, "source": "missing"})
                return 1
            if args.name:
                p = root / args.name
                md = p / "SKILL.md"
                if not p.is_dir():
                    print(f"skill {args.name!r} not found under {root}", file=sys.stderr)
                    return 1
                _print(
                    {
                        "name": args.name,
                        "path": str(p),
                        "skill_md": str(md),
                        "source": "disk",
                    }
                )
            else:
                _print({"path": str(root), "source": "disk"})
            return 0
    except FileNotFoundError as e:
        print(str(e), file=sys.stderr)
        return 1
    print(f"unknown skills action: {action}", file=sys.stderr)
    return 1
