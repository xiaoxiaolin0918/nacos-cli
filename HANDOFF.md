# HANDOFF — nacos-cli (2026-10-21 Asia/Shanghai)

## Status

Go mainline **v0.3.1**: ACL + multi-profile + `--output` + **REPL** + **config export/delete** + **backup** + GitHub Actions CI.

**Live smoke 2026-10-21** against lab `192.168.100.41:8848` (Nacos 2.4.3, auth disabled): all 7 areas pass after fixes — see [`docs/smoke-20261021.md`](docs/smoke-20261021.md).

Python v0 in `archive/python-v0`. `backups/` gitignored.

## Module

https://github.com/xiaoxiaolin0918/nacos-cli (public)

## Build / test

```powershell
cd D:\work\ops\nacos-cli
$env:PATH = "D:\tools\go\bin;" + $env:PATH
$env:GOPROXY = "goproxy.cn,direct"
go test ./...
make build   # or: go build -o nacos.exe .
.\nacos.exe --version
```

## Commands

config (+ export / delete) / service / namespace / user / role / permission / auth / doctor / whoami / profile / skills / **backup** / **repl**

Gates: `--dry-run` / `--yes` (exit 10). ACL + profile show never print password values.

### REPL / export / backup (P2)

```powershell
.\nacos.exe repl
.\nacos.exe config export --data-id app.yaml --out .\app.yaml
.\nacos.exe backup create --all --dry-run   # all groups + all pages of one namespace
.\nacos.exe backup list
```

Doctor best-effort probes `/v1/console/server/state` (version) without failing the whole check.

## Official CLI borrow

| Priority | Item | Status |
|----------|------|--------|
| P1 | profile + `--output` | Done |
| P2 | REPL | Done (writes still gated per line) |
| P2 | v1/v3 API detect | Partial: doctor server_state; adapter later |
| P3 | AgentSpec / Releases | Optional |

## Suggested stop

Ops surface + gates + profile + REPL + export/backup + CI is enough for daily lab use.

## Remaining optional

- Auth-enabled lab ACL smoke (needs user approval; 2026-10-21 smoke ran with auth disabled)
- `service delete` for empty services (Nacos auto-cleans empty ephemeral services ~1 min)
- `service instances` hides disabled instances (v1 list API); add catalog view if needed
- `user create --password` is a flag (visible in process list); consider `--password-stdin`
- Doctor `reachable` shows `HTTP 404` for the root probe (cosmetic)
- Pre-existing smoke leftover: `zhiyi-debug / DEFAULT_GROUP / nacos-cli-smoke-20261020-214546.yaml` (from 2026-10-20, marked 可删) — not removed without approval
- Multi-arch Releases / installers
- Full v3 Bearer client adapter
- AgentSpec lifecycle (only if Nacos AI registry deployed)
- Do **not** touch live prod without explicit approval
