---
name: nacos-service-ops
version: 0.2.0
description: "Nacos service/instance ops: list/instances/register/deregister/up/down. Default ephemeral temporary instances; never touch live prod instances without approval. Risk: read | write (mutations need --yes)."
metadata:
  requires:
    bins: ["nacos", "go"]
  cliHelp: "nacos service --help"
---

# service / instance ops

Read [`../nacos-shared/SKILL.md`](../nacos-shared/SKILL.md) first.

## Use when

- Listing services or instances
- Smoke-testing register → down → up → deregister with a **temporary** ephemeral instance
- Enabling/disabling an instance the user named

## Do NOT

- Register persistent instances by default (`--persistent` only when asked)
- up/down / deregister production instances without explicit user approval + backup note
- Leave smoke instances registered after the test
- Confuse with config publish (see nacos-config-ops)

## Commands

| Command | Risk | Notes |
|---------|------|-------|
| `service list` | read | |
| `service instances` | read | `--service` required |
| `service register` | write | default **ephemeral=true**; needs `--yes` |
| `service deregister` | write | ephemeral must match register |
| `service up` / `down` | write | enabled true/false; pass same ephemeral |

```bash
nacos service list
nacos service instances --service zhiyi-print
nacos service register --service nacos-cli-smoke-svc --ip 127.0.0.1 --port 19849 --dry-run
nacos service register --service nacos-cli-smoke-svc --ip 127.0.0.1 --port 19849 --yes
nacos service down --service nacos-cli-smoke-svc --ip 127.0.0.1 --port 19849 --yes
nacos service up --service nacos-cli-smoke-svc --ip 127.0.0.1 --port 19849 --yes
nacos service deregister --service nacos-cli-smoke-svc --ip 127.0.0.1 --port 19849 --dry-run
nacos service deregister --service nacos-cli-smoke-svc --ip 127.0.0.1 --port 19849 --yes
```

## Failure strategy

| Case | Action |
|------|--------|
| Exit 10 | Re-run with user-approved `--yes` after `--dry-run` |
| deregister mismatch ephemeral | Retry with same ephemeral flag used at register |
| Leftover smoke instance | Always deregister in the same session |
