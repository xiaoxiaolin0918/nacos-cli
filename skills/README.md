# nacos-cli skills

Trae-style agent skills for the **Go** mainline.

| Skill | Role |
|-------|------|
| `nacos-shared` | Auth, doctor, gates, credentials |
| `nacos-config-ops` | Config get/list/publish/history/rollback |
| `nacos-service-ops` | Service list/instances/register/up/down |
| `nacos-acl-ops` | User / role / permission (RBAC ACL) |

```powershell
.\nacos.ps1 skills list
.\nacos.ps1 skills read nacos-acl-ops
```

## AI surfaces

- `nacos commands` / `--help-json` — capability table
- `nacos examples [--command X]` — baked JSON samples
- `--trace` — envelope `meta` (default off)
- dry-run `would[]` array; list `--limit`/`--query`

