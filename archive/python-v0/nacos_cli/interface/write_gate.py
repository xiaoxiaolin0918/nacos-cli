# -*- coding: utf-8 -*-
"""High-risk write gates aligned with plumelog-cli: --dry-run / --yes."""
from __future__ import annotations

import json
import sys
from typing import Any

CONFIRMATION_REQUIRED_EXIT = 10


def check_write_gate(
    *,
    yes: bool,
    dry_run: bool,
    action: str,
    detail: dict[str, Any] | None = None,
) -> int | None:
    """Gate mutating commands.

    Returns:
      0  - dry-run preview printed; caller must NOT send the request
      10 - missing --yes; caller must NOT send the request
      None - proceed with the real request
    """
    payload: dict[str, Any] = {"action": action}
    if detail:
        payload.update(detail)
    if dry_run:
        print(
            json.dumps(
                {"ok": True, "dry_run": True, "would": payload},
                ensure_ascii=False,
                indent=2,
            )
        )
        return 0
    if not yes:
        tip = (
            "\u9ad8\u98ce\u9669\u5199\u64cd\u4f5c\u9700\u8981 --yes"
            "\uff08\u6216 -y\uff09\u624d\u80fd\u6267\u884c\uff1b"
            "\u53ef\u5148\u52a0 --dry-run \u9884\u89c8\u3002"
        )
        print(tip + "\naction=" + action, file=sys.stderr)
        return CONFIRMATION_REQUIRED_EXIT
    return None
