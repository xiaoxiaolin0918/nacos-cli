# nacos-cli wiki

## Home

Go 主实现；仓库入口见根目录 `README.md` / `AGENTS.md`。

| 主题 | 说明 |
|------|------|
| 构建 | `go test ./...` / `go build -o nacos.exe .` |
| 写门禁 | `--dry-run` 或 `--yes`（缺省 exit 10） |
| 输出 | `--format json\|pretty`（默认 json）；`--output` 为别名 |
| Profile | `~/.nacos-cli/<name>.conf`；`--profile` / `--config` |
| 鉴权 | `auth login` / `--web`；token 落本机 credentials.json |
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