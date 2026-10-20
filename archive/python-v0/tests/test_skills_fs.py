# -*- coding: utf-8 -*-
from __future__ import annotations

from nacos_cli.infrastructure.skills_fs import list_skills, parse_frontmatter, read_skill


def test_parse_frontmatter():
    text = """---
name: nacos-demo
version: 0.1.0
description: "Demo skill for tests."
---

# body
"""
    meta = parse_frontmatter(text)
    assert meta["name"] == "nacos-demo"
    assert meta["version"] == "0.1.0"
    assert "Demo skill" in meta["description"]


def test_list_and_read_skills(tmp_path):
    skill_dir = tmp_path / "nacos-demo"
    skill_dir.mkdir()
    (skill_dir / "SKILL.md").write_text(
        """---
name: nacos-demo
version: 0.0.1
description: "Use when testing skills discovery."
---

# Demo
""",
        encoding="utf-8",
    )
    listed = list_skills(tmp_path)
    assert listed["source"] == "disk"
    assert len(listed["skills"]) == 1
    assert listed["skills"][0]["name"] == "nacos-demo"
    body = read_skill("nacos-demo", tmp_path)
    assert "Demo" in body["content"]
    assert body["version"] == "0.0.1"


def test_repo_skills_if_present():
    listed = list_skills()
    if listed["source"] == "disk":
        names = {s["dir"] for s in listed["skills"]}
        assert "nacos-shared" in names or bool(names)
