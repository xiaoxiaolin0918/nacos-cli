---
name: nacos-shared
version: 0.3.0
description: "Use for nacos-cli (Go) setup/auth: auth login/status/logout (--web), whoami, doctor, .env + local credentials file, AUTH_DISABLED skip, --dry-run/--yes write gates, skills list/read, profile list/show/set, --profile/--config/--output, repl/interactive shell, backup path/list/create. Risk: read (status/doctor) | write (login)."
metadata:
  requires:
    bins: ["nacos", "go"]
  cliHelp: "nacos --help"
---

# nacos-cli shared (auth / doctor / gates)

Read this before any other `nacos-*` skill. **Mainline is Go + cobra** (Python v0 archived under `archive/python-v0`).

## Use when

- Setting up NACOS_SERVER / credentials for the first time
- Running `auth login` (password flags or `--web` local browser form)
- Checking session with `auth status` / `whoami` / `doctor`
- Listing or reading bundled skills

## Do NOT

- Paste real passwords, accessTokens, or credentials.json into chat / Issues / docs
- Commit `.env` or `%APPDATA%/nacos-cli/credentials.json`
- Default `--yes` on high-risk writes without user confirmation
- Mutate live prod configs; prefer new smoke dataIds / ephemeral instances

## Input / Output

| Input | Source |
|-------|--------|
| Server | `NACOS_SERVER` / `.env` |
| Username | `--username` / `NACOS_USERNAME` |
| Password | `--password` / `NACOS_PASSWORD` / web form (never logged) |
| Token | local credentials file after login |

Success stdout JSON `{ok,result}`. Failures on stderr; never print raw token (only `token_masked`).

## Steps

1. Copy `.env.example` → `.env` and set `NACOS_SERVER` (placeholder host in docs).
2. `go build -o nacos.exe .` then `.\nacos.exe doctor` (or `.\nacos.ps1 doctor`).
3. If auth enabled: `nacos auth login --username ... --password ...` **or** `auth login --web`.
4. `nacos auth status` / `whoami` — confirm `has_token` / `auth_mode` without secrets.
5. High-risk writes need `--dry-run` first, then user-approved `--yes`.
6. `config publish` defaults to create-only; use `--overwrite` only when the user asks to update an existing dataId.

```powershell
.\nacos.ps1 doctor
.\nacos.ps1 auth login --web
.\nacos.ps1 auth status
.\nacos.ps1 whoami --probe
.\nacos.ps1 auth logout
.\nacos.ps1 skills list
.\nacos.ps1 skills read nacos-shared
```


## REPL / backup (P2)

```powershell
.\nacos.ps1 repl
.\nacos.ps1 backup create --all --group DEFAULT_GROUP --dry-run
```

## Profiles and output (P1)

```powershell
.\nacos.ps1 profile set lab --server http://127.0.0.1:8848 --username nacos
.\nacos.ps1 --profile lab namespace list --output json
.\nacos.ps1 whoami --format pretty
.\nacos.ps1 profile show lab
```

- Profiles: `~/.nacos-cli/<name>.conf` (YAML). Never commit. `profile show` masks password.
- Priority: flags > `--config` > `--profile` > env/.env.
- `--output` aliases `--format` (`json` default, `pretty` optional).
## Failure strategy

| Case | Action |
|------|--------|
| `AUTH_DISABLED` / auth_mode=disabled | Proceed without token; do not fail doctor |
| Login HTTP 4xx | Fix username/password; do not retry blindly |
| Missing NACOS_SERVER | Point user to `.env.example` |
| Web login timeout | Re-run `auth login --web`; check browser blocked loopback |
| Write without `--yes` | Exit 10; show action; wait for explicit approval |

## Credentials path

| OS | Default |
|----|---------|
| Windows | `%APPDATA%\nacos-cli\credentials.json` |
| Unix | `$XDG_CONFIG_HOME/nacos-cli/credentials.json` or `~/.config/nacos-cli/credentials.json` |

Override: `NACOS_CREDENTIALS_PATH`.