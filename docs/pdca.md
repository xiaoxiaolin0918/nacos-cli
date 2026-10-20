# PDCA — nacos-cli

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
