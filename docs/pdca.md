# PDCA — nacos-cli

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
| P1 | --output json / pretty (or keep --format) | **Done 2026-10-21** — `--output` aliases `--format`; default stays `json` |
| P1 | Profile (~/.nacos-cli/*.conf, profile show/set, --profile) | **Done 2026-10-21** — list/show/path/set/delete; `--config` overrides |
| P2 | Interactive terminal (nacos> REPL) | Later; keep non-interactive gates mandatory for writes |
| P2 | Server version detect (v1 vs v3 API / Bearer) | Lab is 2.4.x (v1); design behind client adapter |
| P3 | AI Skill / AgentSpec lifecycle | Only if our Nacos deploys AI registry |
| P3 | Installer / npm / multi-arch Releases | Match plumelog when we cut v0.1 |
| skip-for-now | Naming up/down in official | We already cover register/up/down |
| skip-for-now | RBAC ACL in official | Official docs lack user/role/permission CLI — our ACL differentiates |

Act when borrowed: add failing tests first; keep Risk gates; update skills + pdca Check.