# HANDOFF — nacos-cli (2026-10-21 Asia/Shanghai)

## Status

Go mainline **v0.3.0**: ACL + multi-profile + `--output` + **REPL** + **config export** + **backup** + GitHub Actions CI.

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

config (+ export) / service / namespace / user / role / permission / auth / doctor / whoami / profile / skills / **backup** / **repl**

Gates: `--dry-run` / `--yes` (exit 10). ACL + profile show never print password values.

### REPL / export / backup (P2)

```powershell
.\nacos.exe repl
.\nacos.exe config export --data-id app.yaml --out .\app.yaml
.\nacos.exe backup create --all --group DEFAULT_GROUP --dry-run
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

- Auth-enabled lab ACL smoke (needs user approval)
- Multi-arch Releases / installers
- Full v3 Bearer client adapter
- AgentSpec lifecycle (only if Nacos AI registry deployed)
- Do **not** touch live prod without explicit approval
