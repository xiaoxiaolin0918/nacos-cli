# HANDOFF — nacos-cli (2026-10-20 Asia/Shanghai)

## Status
Go mainline with ACL (user/role/permission). Python v0 in archive/python-v0. backups/ gitignored.

## Module
github.com/xiaoxiaolin0918/nacos-cli (public GitHub).

## Build / test
```powershell
cd D:\work\ops\nacos-cli
$env:PATH = "D:\tools\go\bin;" + $env:PATH
$env:GOPROXY = "goproxy.cn,direct"
go test ./...
go build -o nacos.exe .
```

## Commands
config / service / namespace / user / role / permission / auth / doctor / whoami / skills

Gates: --dry-run / --yes (exit 10). ACL dry-run never prints password values.

## Official CLI
nacos-group/nacos-cli: AI Skill/AgentSpec/config; no naming up-down or RBAC ACL — this ops CLI fills that.

## Next
Auth-enabled lab ACL smoke only with approval; optional Releases/CI.

## Official CLI borrow list (nacos-group/nacos-cli)

Reference: https://github.com/nacos-group/nacos-cli and https://nacos.io/docs/latest/manual/admin/nacos-cli/

Do **not** wholesale rewrite our command surface this cycle. Pick high-value gaps later.

| Priority | Item | Why / notes for this ops CLI |
|----------|------|------------------------------|
| P1 | --output json / pretty (or keep --format) | Align scripting contract with official; we already have --format json|pretty — document parity, avoid dual flags unless needed |
| P1 | Profile (~/.nacos-cli/*.conf, profile edit/show, --profile) | Multi-env ops without juggling .env; map to our settings + credstore carefully (never store password in repo) |
| P2 | Interactive terminal (
acos> REPL: help/ns/server/quit) | Faster ops exploration; optional later, keep non-interactive gates (--yes/--dry-run) mandatory for writes |
| P2 | Server version detect (v1 vs v3 API / Bearer) | Lab is 2.4.x (v1); future 3.x needs /v3/auth/* path — design behind client adapter |
| P3 | AI Skill / AgentSpec lifecycle | Official focus; only if our Nacos deploys AI registry — not core ops ACL/naming |
| P3 | Installer / npm / multi-arch Releases | Match plumelog Releases habit when we cut v0.1 |
| skip-for-now | Naming up/down in official | Official marks naming ❌ — we already cover register/up/down; no borrow needed |
| skip-for-now | RBAC ACL in official | Official docs do not expose user/role/permission CLI — our ACL stays differentiating |

Act when borrowed: add failing tests first; keep Risk gates; update skills + pdca Check.
