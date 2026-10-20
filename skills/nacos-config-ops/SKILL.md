---
name: nacos-config-ops
version: 0.2.0
description: "Nacos config ops: config get/list/publish/history/rollback. publish defaults to create-only; overwrite needs --overwrite. Risk: read | write (publish/rollback need --yes or --dry-run)."
metadata:
  requires:
    bins: ["nacos", "go"]
  cliHelp: "nacos config --help"
---

# config ops

Read [`../nacos-shared/SKILL.md`](../nacos-shared/SKILL.md) first.

## Use when

- Reading or listing configs
- Publishing a **new** smoke dataId for tests (default create-only)
- Inspecting history / rolling back a config you own

## Do NOT

- Overwrite existing production dataIds unless the user explicitly asks (then use `--overwrite`)
- Skip backup before publish/rollback against shared namespaces
- Show or run publish/rollback without `--dry-run` or `--yes`
- Use this skill for service register/up/down (see nacos-service-ops)

## Commands

| Command | Risk | Notes |
|---------|------|-------|
| `config get` | read | `--data-id` required |
| `config list` | read | blur search; empty dataId ok |
| `config publish` | write | needs `--yes` or `--dry-run`; **default create-only**; `--overwrite` to update |
| `config history` | read | for rollback nid |
| `config rollback` | write | needs `--nid` + `--yes` or `--dry-run` |

```bash
nacos config list --namespace zhiyi-debug
nacos config get --data-id blade-dev.yaml --group DEFAULT_GROUP
# new smoke (default create-only)
nacos config publish --data-id nacos-cli-smoke.yaml --content "k: v" --dry-run
nacos config publish --data-id nacos-cli-smoke.yaml --content "k: v" --yes
# update existing only when user asked
nacos config publish --data-id existing.yaml --content "k: v2" --overwrite --dry-run
nacos config publish --data-id existing.yaml --content "k: v2" --overwrite --yes
nacos config history --data-id nacos-cli-smoke.yaml
nacos config rollback --data-id nacos-cli-smoke.yaml --nid <id> --dry-run
nacos config rollback --data-id nacos-cli-smoke.yaml --nid <id> --yes
```

## Failure strategy

| Case | Action |
|------|--------|
| Exit 10 / missing `--yes` | Show dry-run; ask user; re-run with `--yes` |
| Exists under create-only (default) | Choose a new dataId, or only if user asks: `--overwrite` |
| Empty history | Cannot rollback; publish a new version instead |
