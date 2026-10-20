# nacos-cli

Nacos ops CLI: console-shaped habits, Open API under the hood.

**Mainline: Go + cobra** (same stack as plumelog-cli / yunxiao-cli).  
Python probe archived at [`archive/python-v0`](archive/python-v0) (deprecated).

Module: [`github.com/xiaoxiaolin0918/nacos-cli`](https://github.com/xiaoxiaolin0918/nacos-cli)

## Prepare

```powershell
cd D:\work\ops\nacos-cli
copy .env.example .env
# edit .env: NACOS_SERVER / USERNAME / PASSWORD / NAMESPACE / CONTEXT_PATH
```

Or use a **profile** (official layout `~/.nacos-cli/<name>.conf`):

```powershell
.\nacos.exe profile set lab --server http://127.0.0.1:8848 --username nacos
.\nacos.exe --profile lab namespace list
.\nacos.exe --config $env:USERPROFILE\.nacos-cli\lab.conf whoami
```

| Source | Notes |
|--------|------|
| flags | `--server` / `--username` / `--namespace` / `--profile` / `--config` |
| `--config` | explicit YAML file (overrides `--profile`) |
| `--profile` | default `default` → `~/.nacos-cli/default.conf` |
| env / `.env` | `NACOS_SERVER`, `NACOS_USERNAME`, `NACOS_PASSWORD`, … |
| defaults | context `/nacos` |

**Never commit** `.env`, profile passwords, tokens, or `backups/`.

## Build / test

```powershell
$env:PATH = "D:\tools\go\bin;" + $env:PATH
$env:GOPROXY = "goproxy.cn,direct"
go test ./...
go build -o nacos.exe .
.\nacos.exe --help
# or
.\nacos.ps1 --help
```

## Output format

Agent-friendly JSON envelope by default (unlike official CLI which defaults to pretty).

| Flag | Values |
|------|--------|
| `--format` | `json` (default) \| `pretty` |
| `--output` | alias of `--format` (official CLI parity) |

```powershell
.\nacos.exe namespace list --output json
.\nacos.exe whoami --format pretty
.\nacos.exe profile list --output pretty
```

## High-risk write gates

Writes need **`--dry-run`** (preview) or **`--yes`/`-y`** (execute). Missing gate → exit **10**.

| Write | Default |
|-------|---------|
| `config publish` | **create-only**; overwrite needs `--overwrite` |
| `config rollback` | needs `--nid` |
| `service register` | ephemeral; persistent needs `--persistent` |
| `service deregister` / `up` / `down` | ephemeral flag must match register |
| ACL user/role/permission writes | same `--yes` / `--dry-run` |

## Command surface

### Config

```powershell
.\nacos.ps1 config get --data-id app.yaml --group DEFAULT_GROUP
.\nacos.ps1 config list --data-id "" --group DEFAULT_GROUP
.\nacos.ps1 config publish --data-id nacos-cli-smoke.yaml --group DEFAULT_GROUP --content "k: v" --dry-run
.\nacos.ps1 config publish --data-id nacos-cli-smoke.yaml --group DEFAULT_GROUP --content "k: v" --yes
.\nacos.ps1 config publish --data-id app.yaml --group DEFAULT_GROUP --file .\app.yaml --overwrite --dry-run
.\nacos.ps1 config history --data-id app.yaml --group DEFAULT_GROUP
.\nacos.ps1 config rollback --data-id app.yaml --group DEFAULT_GROUP --nid <historyId> --dry-run
```

### Service instances

```powershell
.\nacos.ps1 service list
.\nacos.ps1 service instances --service my-service
.\nacos.ps1 service register --service nacos-cli-smoke-svc --ip 127.0.0.1 --port 19849 --dry-run
.\nacos.ps1 service register --service nacos-cli-smoke-svc --ip 127.0.0.1 --port 19849 --yes
.\nacos.ps1 service down --service nacos-cli-smoke-svc --ip 127.0.0.1 --port 19849 --yes
.\nacos.ps1 service up --service nacos-cli-smoke-svc --ip 127.0.0.1 --port 19849 --yes
.\nacos.ps1 service deregister --service nacos-cli-smoke-svc --ip 127.0.0.1 --port 19849 --yes
```

### Namespace / auth / skills / profile

```powershell
.\nacos.ps1 namespace list --output json
.\nacos.ps1 doctor
.\nacos.ps1 auth login
.\nacos.ps1 auth login --web
.\nacos.ps1 auth status
.\nacos.ps1 whoami --probe
.\nacos.ps1 skills list
.\nacos.ps1 profile list
.\nacos.ps1 profile show lab
.\nacos.ps1 profile path lab
.\nacos.ps1 profile set lab --server http://127.0.0.1:8848 --username nacos
.\nacos.ps1 profile delete lab --dry-run
.\nacos.ps1 profile delete lab --yes
```

### ACL (RBAC)

```powershell
.\nacos.ps1 user list
.\nacos.ps1 user create --username alice --password REDACTED --dry-run
.\nacos.ps1 user create --username alice --password REDACTED --yes
.\nacos.ps1 role bind --role dev --username alice --yes
.\nacos.ps1 permission grant --role dev --resource "*:*:*" --action r --yes
```

Requires auth-enabled Nacos + admin login. Passwords never appear in dry-run JSON.

## Methodology

- **TDD + DDD + PDCA** (aligned with plumelog-cli)
- Agent entry: [`AGENTS.md`](AGENTS.md)
- PDCA: [`docs/pdca.md`](docs/pdca.md)
- Wiki: [`docs/wiki/Home.md`](docs/wiki/Home.md)

## Official CLI

Upstream [`nacos-group/nacos-cli`](https://github.com/nacos-group/nacos-cli) focuses on Skill/AgentSpec + config.  
This ops CLI differentiates with naming up/down, ACL, `--yes`/`--dry-run` gates, and web login.