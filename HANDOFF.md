# HANDOFF — nacos-cli (2026-10-21 Asia/Shanghai)

## Status

Go mainline with ACL + **multi-profile** + **`--output` alias**.  
Python v0 in `archive/python-v0`. `backups/` gitignored.

## Module

https://github.com/xiaoxiaolin0918/nacos-cli (public)

## Build / test

```powershell
cd D:\work\ops\nacos-cli
$env:PATH = "D:\tools\go\bin;" + $env:PATH
$env:GOPROXY = "goproxy.cn,direct"
go test ./...
go build -o nacos.exe .
```

## Commands

config / service / namespace / user / role / permission / auth / doctor / whoami / **profile** / skills

Gates: `--dry-run` / `--yes` (exit 10). ACL + profile show never print password values.

### Profile (P1 done)

```powershell
.\nacos.exe profile set lab --server http://127.0.0.1:8848 --username nacos
.\nacos.exe --profile lab namespace list --output json
.\nacos.exe --config $env:USERPROFILE\.nacos-cli\lab.conf whoami
.\nacos.exe profile show lab
.\nacos.exe profile delete lab --yes
```

Layout: `~/.nacos-cli/<name>.conf` (YAML). Priority: flags > `--config` > `--profile` > env/.env.

### Output (P1 done)

`--format json|pretty` (default **json**). `--output` is an official-CLI alias of `--format`.

## Official CLI

nacos-group/nacos-cli: AI Skill/AgentSpec/config; no naming up-down or RBAC ACL — this ops CLI fills that.

## Next

- P2: interactive REPL; v1/v3 API detect
- Optional: Releases/CI; auth-enabled lab ACL smoke only with user approval
- Do **not** touch live prod without explicit approval