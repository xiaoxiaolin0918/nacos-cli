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
