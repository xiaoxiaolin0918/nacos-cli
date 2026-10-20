# -*- coding: utf-8 -*-
"""Unified entry re-export (interface layer)."""
from nacos_cli.interface.cli import build_parser, main

__all__ = ["build_parser", "main"]

if __name__ == "__main__":
    raise SystemExit(main())
