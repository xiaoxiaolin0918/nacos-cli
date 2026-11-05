# nacos-cli wiki

## Home

Go 主实现；仓库入口见根目录 `README.md` / `AGENTS.md`。

| 主题 | 说明 |
|------|------|
| 构建 | `go test ./...` / `go build -o nacos.exe .` |
| 写门禁 | `--dry-run` 或 `--yes`（缺省 exit 10） |
| 输出 | `--format json\|pretty`（默认 json）；`--output` 为别名 |
| Profile | `~/.nacos-cli/<name>.conf`；`--profile` / `--config` |
| 鉴权 | `auth login [--remember]` / `--web`；token→credentials.json；可选加密 login.cred（Win=DPAPI / 其它 AES+login.key）+ `--auto-login`（401 续一次）；`logout` 默认只清 token，`--forget` 才双清（含 `.login/<profile>`） |
| 归档 | Python v0 → `archive/python-v0` |

示例主机只用 `http://127.0.0.1:8848`，不要在公开文档写实验室 IP。

### Profile 示例

```powershell
.\nacos.exe profile set lab --server http://127.0.0.1:8848 --username nacos
.\nacos.exe --profile lab namespace list --output json
.\nacos.exe profile show lab
.\nacos.exe profile delete lab --yes
```

优先级：flags > `--config` > `--profile` > env / `.env` > defaults。

## Export / Backup / REPL

```powershell
.\nacos.exe config export --data-id app.yaml --out .\app.yaml
.\nacos.exe backup create --all --group DEFAULT_GROUP --dry-run
.\nacos.exe repl
```

本地备份目录默认 `./backups`（可用 `NACOS_BACKUP_DIR`），已 gitignore。REPL 内写操作仍需 `--dry-run` / `--yes`。

## 本地监管 Web

```text
nacos web
```

仅绑定 `127.0.0.1`。可管理 `~/.nacos-cli/*.conf` 连接字段与 `*.meta.yaml` 备注（备注名 / lab|prod / 颜色 / 负责人 / 注意事项），并浏览 `~/.nacos-cli/audit.jsonl` 操作记录。不通过 Web 直接改远端 Nacos。密码在页面上永不回显明文。
