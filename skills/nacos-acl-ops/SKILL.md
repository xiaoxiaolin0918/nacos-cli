---
name: nacos-acl-ops
version: 0.1.0
description: "Use when managing Nacos RBAC via nacos-cli (Go): user list/create/delete/passwd, role list/bind/unbind, permission list/grant/revoke. High-risk writes need --yes; --dry-run previews. Never echo passwords."
metadata:
  requires:
    bins: ["nacos", "go"]
  cliHelp: "nacos user --help; nacos role --help; nacos permission --help"
---

# nacos-cli ACL (user / role / permission)

Read `nacos-shared` first. Requires auth-enabled Nacos + admin session (`auth login`).

## Use when

- Creating/deleting local Nacos users or rotating passwords
- Binding/unbinding roles to users
- Granting/revoking resource permissions (namespace:group:dataId, action r|w|rw)

## Do NOT

- Print or log passwords / tokens
- Run writes without `--dry-run` preview or explicit `--yes`
- Assume lab AUTH_DISABLED supports admin ACL APIs
- Use on prod without backup + user confirmation

## Input / Output

| Input | Source |
|-------|--------|
| Admin session | `auth login` credentials file |
| Username / role / resource / action | flags |
| Password | `--password` only (never in dry-run would) |

Stdout JSON `{ok,result}` or `{ok,dry_run,would}`. Exit 10 if high-risk without `--yes`.

## Steps

1. `nacos doctor` / `nacos auth status`
2. Preview: `nacos user create --username alice --password REDACTED --dry-run`
3. Execute with `--yes` after confirmation
4. `nacos role bind --role dev --username alice --yes`
5. `nacos permission grant --role dev --resource '*:*:*' --action r --yes`
6. Verify with list subcommands

## Failure

- HTTP 403/401: re-login as ROLE_ADMIN; confirm auth enabled
- Exit 10: add `--yes` or use `--dry-run`
