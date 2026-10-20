# AGENTS.md — nacos-cli

Practical rules for humans and AI agents working on **nacos-cli** (Nacos ops CLI).

Aligns with sibling CLIs **yunxiao-cli** / **plumelog-cli**: English `AGENTS.md` + Chinese `docs/wiki/`.

## Purpose

- Work checkout: `D:\work\ops\nacos-cli` (this tree).
- Bot revival pack (not git): `D:\ai\grok-bot\nacos-cli\` (`AGENT.md` / `HANDOFF.md`).
- Open API under the hood; console-shaped command habits.
- **Mainline: Go + cobra**. Module: `github.com/xiaoxiaolin0918/nacos-cli`.
- Python v0 archived at `archive/python-v0/` (deprecated; do not extend).

## Methodology (required)

| Practice | Rule |
|----------|------|
| **TDD** | Prefer failing `go test` first for domain / risk / gates / profile; then implement until green. Run `go test ./...` before claiming done. |
| **DDD** | Keep layers strict: `internal/domain` (use cases + ports) / `internal/client` (HTTP) / `cmd` (cobra). Domain must not import `net/http` or cobra. |
| **PDCA** | Record Plan/Do/Check/Act in `docs/pdca.md`; close gaps explicitly. |

## Directory boundaries

```
nacos-cli/
  AGENTS.md / HANDOFF.md / README.md
  main.go / go.mod / Makefile / nacos.ps1
  cmd/
  internal/domain|client|config|credstore|risk|output|version
  skills/
  docs/pdca.md / docs/wiki/
  backups/                 # gitignored
  archive/python-v0/
```

Profiles live **outside** the repo: `~/.nacos-cli/<name>.conf` (YAML). Never commit them.

## Command surface

| Group | Actions |
|-------|---------|
| `config` | `get` `list` `publish` `history` `rollback` |
| `service` | `list` `instances` `register` `deregister` `up` `down` |
| `namespace` | `list` |
| `user` | `list` `create` `delete` `passwd` |
| `role` | `list` `bind` `unbind` |
| `permission` | `list` `grant` `revoke` |
| `auth` | `login` `logout` `status` (password or `--web`) |
| `whoami` / `doctor` | read-only session / connectivity |
| `profile` | `list` `show` `path` `set` `delete` |
| `skills` | `list` `read` `path` |

ACL uses Open API `/v1/auth/users|roles|permissions` (RBAC). Needs auth-enabled Nacos + admin token.

## Config priority

1. Flags (`--server`, `--username`, `--namespace`, …)
2. `--config` / `-c` explicit YAML file
3. `--profile` (default `default` → `~/.nacos-cli/default.conf`)
4. env / `.env`
5. defaults (`contextPath=/nacos`)

## Output

| Flag | Values | Default |
|------|--------|---------|
| `--format` | `json` \| `pretty` | `json` (agent-friendly; official CLI defaults pretty) |
| `--output` | alias of `--format` | empty → use `--format` / profile `format` |

Read commands (`list` / `get` / `whoami` / `profile show` / …) always emit the JSON ok-envelope unless `--format pretty` / `--output pretty`.

## High-risk write gates (required)

| Flag | Behavior |
|------|----------|
| `--dry-run` | Print JSON `{ok, dry_run, would}`; never calls Nacos. |
| `--yes` / `-y` | Required for high-risk writes; else exit **10**. |

**High-risk**: config publish/rollback; service register/deregister/up/down; user create/delete/passwd; role bind/unbind; permission grant/revoke; **profile delete**.

- publish defaults create-only; register defaults ephemeral.
- ACL dry-run never echoes passwords (only `password_set` / `password_length`).
- `profile show` masks password as `******`.
- Doc hosts: `http://127.0.0.1:8848` (redact lab IPs).

## Secrets — hard rules

Never commit `.env`, tokens, passwords, `backups/`, `~/.nacos-cli/*.conf`, or credentials files.

## Build & test

```powershell
cd D:\work\ops\nacos-cli
$env:PATH = "D:\tools\go\bin;" + $env:PATH
$env:GOPROXY = "goproxy.cn,direct"
go test ./...
go build -o nacos.exe .
```

## Official CLI note

Borrow from [`nacos-group/nacos-cli`](https://github.com/nacos-group/nacos-cli) selectively (see `docs/pdca.md`). Keep our naming up/down + ACL + gates as differentiators. Do not wholesale rewrite the command surface.