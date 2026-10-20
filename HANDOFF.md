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
