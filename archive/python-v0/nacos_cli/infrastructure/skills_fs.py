# -*- coding: utf-8 -*-
"""Discover and read bundled agent skills/ tree."""
from __future__ import annotations

import re
from pathlib import Path
from typing import Any

FRONTMATTER_RE = re.compile(r"^---\s*\n(.*?)\n---\s*\n", re.DOTALL)
NAME_RE = re.compile(r'(?m)^name:\s*["\']?([^"\'\n]+)["\']?\s*$')
DESC_RE = re.compile(r'(?m)^description:\s*["\']?(.*?)["\']?\s*$')
VERSION_RE = re.compile(r'(?m)^version:\s*["\']?([^\s"\']+)["\']?\s*$')


def repo_root() -> Path:
    return Path(__file__).resolve().parents[2]


def find_skills_root(explicit: Path | None = None) -> Path | None:
    candidates: list[Path] = []
    if explicit is not None:
        candidates.append(explicit)
    candidates.append(repo_root() / "skills")
    candidates.append(Path.cwd() / "skills")
    for c in candidates:
        if c.is_dir() and _has_nacos_skill(c):
            return c.resolve()
    return None


def _has_nacos_skill(root: Path) -> bool:
    for child in root.iterdir():
        if child.is_dir() and child.name.startswith("nacos-"):
            if (child / "SKILL.md").is_file():
                return True
    return False


def parse_frontmatter(text: str) -> dict[str, str]:
    m = FRONTMATTER_RE.match(text)
    block = m.group(1) if m else text[:500]
    out: dict[str, str] = {}
    nm = NAME_RE.search(block)
    if nm:
        out["name"] = nm.group(1).strip()
    dm = DESC_RE.search(block)
    if dm:
        out["description"] = dm.group(1).strip()
    vm = VERSION_RE.search(block)
    if vm:
        out["version"] = vm.group(1).strip()
    return out


def list_skills(root: Path | None = None) -> dict[str, Any]:
    base = find_skills_root(root)
    if base is None:
        return {"skills": [], "skills_dir": None, "source": "missing"}
    items: list[dict[str, str]] = []
    for child in sorted(base.iterdir(), key=lambda p: p.name):
        if not child.is_dir() or not child.name.startswith("nacos-"):
            continue
        md = child / "SKILL.md"
        if not md.is_file():
            continue
        meta = parse_frontmatter(md.read_text(encoding="utf-8"))
        items.append(
            {
                "name": meta.get("name") or child.name,
                "dir": child.name,
                "version": meta.get("version") or "",
                "description": meta.get("description") or "",
                "path": str(md),
            }
        )
    return {"skills": items, "skills_dir": str(base), "source": "disk"}


def read_skill(name: str, root: Path | None = None) -> dict[str, Any]:
    base = find_skills_root(root)
    if base is None:
        raise FileNotFoundError("skills directory not found")
    # accept directory name or frontmatter name
    candidates = [base / name / "SKILL.md"]
    for child in base.iterdir():
        if child.is_dir() and (child / "SKILL.md").is_file():
            meta = parse_frontmatter((child / "SKILL.md").read_text(encoding="utf-8"))
            if meta.get("name") == name:
                candidates.insert(0, child / "SKILL.md")
    for md in candidates:
        if md.is_file():
            text = md.read_text(encoding="utf-8")
            meta = parse_frontmatter(text)
            return {
                "name": meta.get("name") or name,
                "path": str(md),
                "version": meta.get("version") or "",
                "content": text,
            }
    raise FileNotFoundError(f"skill {name!r} not found under {base}")
