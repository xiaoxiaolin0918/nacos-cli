## Cycle 2026-10-24 (Issues #1 #2 #3)

### Plan
Fix inconsistent `public` namespace, config publish type inference, and Git Bash mangled context-path. TDD; local checkout.

### Do
- domain: NormalizeNamespace / EffectiveNamespace; InferConfigTypeFromDataID / ResolvePublishType
- client.ns uses EffectiveNamespace; config MergeSettings/FromEnv normalize public
- config publish: infer type, require --type when unknown, --change-type on type change; dry-run shows effective_namespace
- profile set: ValidateContextPath rejects MSYS paths; store public as empty; flag help mentions MSYS_NO_PATHCONV=1

### Check
go test ./...

### Act
Fixes #1 #2 #3

# PDCA 鈥?nacos-cli

## Cycle 2026-10-21 (P2 REPL + export/backup + CI)

### Plan
Land unfinished P2 work: interactive REPL, config export, local backup, GitHub Actions CI, Makefile VERSION/ldflags, docs. No secrets; no prod writes. Push when `go test ./...` green.

### Do
- Wire `repl` / `backup` into root; `config export` into config
- `internal/repl` tokenizer; `internal/backup` + `domain` export helpers + tests
- Doctor best-effort `ServerState` version probe
- `.github/workflows/ci.yml`; Makefile `VERSION`/`LDFLAGS`; version bump **0.3.0** (var + UserAgent())
- Docs: README / AGENTS / HANDOFF / wiki / pdca / skills shared+config

### Check
`go test ./...` green; `nacos --version` 鈫?0.3.0; help lists backup/repl/export.

### Act
P2 REPL/export/backup/CI closed. Remaining optional: Releases, v3 adapter, AgentSpec, auth-enabled ACL smoke with approval.


## Cycle 2026-10-21 (P1 profile + --output)

### Plan
Borrow official P1: multi-profile (`~/.nacos-cli/*.conf`, `--profile`, `--config`) and global `--output` alias for `--format`. Align docs/tests with plumelog habits. TDD; push GitHub; update HANDOFF/pdca. No prod writes; no secrets in git.

### Do
- `internal/config/profile.go` (+ tests): YAML load/save/list/delete, host+port compose, MergeSettings, NormalizeFormat, password redaction
- `cmd/profile.go`: list / show / path / set / delete (`delete` high-risk)
- Root flags: `--profile`, `--config`/`-c`, `--output`
- Resolve priority: flags > --config > --profile > env/.env > defaults
- Docs: README / AGENTS / HANDOFF / wiki / skills shared
- `.gitignore`: `.nacos-cli/`, `*.conf`

### Check
`go test ./...` green (config profile suite + cmd profile/output suite + output package).

### Act
P1 closed. Next borrow: P2 REPL / API v3 detect. Keep gates. Optional Releases/CI.

## Cycle 2026-10-20 (Go rewrite)

### Plan
Unify with plumelog-cli (Go + cobra + gates + skills). Archive Python v0.

### Do
Go module, domain/client/cmd layers, create-only publish, ephemeral register.

### Check
`go test ./...` green.

### Act
Open gaps closed in ACL cycle below.

## Cycle 2026-10-20 (ACL + GitHub)

### Plan
Public GitHub (no .env/secrets/backups). TDD user/role/permission with --dry-run/--yes. Update AGENTS/skills/pdca/HANDOFF.

### Do
AuthAdminPort + client v1 auth APIs + cmd user/role/permission + skill nacos-acl-ops + gh repo create/push.

### Check
`go test ./...` green; password masked in dry-run; grant without --yes exits 10.

### Act
Next: optional Releases/CI; auth-enabled lab ACL smoke only with user approval.

## Official CLI borrow list (nacos-group/nacos-cli)

Reference: https://github.com/nacos-group/nacos-cli and https://nacos.io/docs/latest/manual/admin/nacos-cli/

Do **not** wholesale rewrite our command surface. Pick high-value gaps.

| Priority | Item | Status / notes |
|----------|------|----------------|
| P1 | --output json / pretty (or keep --format) | **Done 2026-10-21** 鈥?`--output` aliases `--format`; default stays `json` |
| P1 | Profile (~/.nacos-cli/*.conf, profile show/set, --profile) | **Done 2026-10-21** 鈥?list/show/path/set/delete; `--config` overrides |
| P2 | Interactive terminal (nacos> REPL) | **Done 2026-10-21** 鈥?`nacos repl` / bare TTY; gates per line |
| P2 | Server version detect (v1 vs v3 API / Bearer) | Lab is 2.4.x (v1); design behind client adapter |
| P3 | AI Skill / AgentSpec lifecycle | Only if our Nacos deploys AI registry |
| P3 | Installer / npm / multi-arch Releases | Match plumelog when we cut v0.1 |
| skip-for-now | Naming up/down in official | We already cover register/up/down |
| skip-for-now | RBAC ACL in official | Official docs lack user/role/permission CLI 鈥?our ACL differentiates |

Act when borrowed: add failing tests first; keep Risk gates; update skills + pdca Check.

## Cycle 2026-10-21 (live smoke v0.3.0 鈫?v0.3.1)

### Plan
Approved smoke vs lab 2.4.3 (auth disabled): backup first; write only `nacos-cli-smoke-*`; `--dry-run` before `--yes`; clean up.

### Do
Ran 7 areas (doctor/auth/profile, read+backup, config lifecycle, service lifecycle, ACL lifecycle, REPL via stdin, output formats). Fixed bugs via TDD.

### Check
All areas pass on v0.3.1; `go test ./...` green. Record: [`smoke-20261021.md`](smoke-20261021.md).

### Act
Bugs fixed: export `--dry-run` wrote files; rollback reset type to text (+ no dataId check); `user list` leaked bcrypt hashes; backup/export `--all` only DEFAULT_GROUP + one page; same-second backups merged; `profile list` null; `config list` blur needed `*`. Added `config delete`. Open items in HANDOFF.
