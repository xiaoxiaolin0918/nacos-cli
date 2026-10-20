# nacos-cli

Nacos 常用运维命令封装。贴近控制台习惯，底层走 Open API。

**主实现：Go + cobra**（与 plumelog-cli / yunxiao-cli 同栈）。  
Python 探路版已归档到 [`archive/python-v0`](archive/python-v0)（deprecated）。

模块：`github.com/xiaoxiaolin0918/nacos-cli`（远端未建时可本地 `go build`）。

## 准备

```powershell
cd D:\work\ops\nacos-cli
copy .env.example .env
# 编辑 .env：NACOS_SERVER / USERNAME / PASSWORD / NAMESPACE / CONTEXT_PATH
```

| 变量 | 说明 |
|------|------|
| `NACOS_SERVER` | 如 `http://127.0.0.1:8848` |
| `NACOS_USERNAME` | 登录用户 |
| `NACOS_PASSWORD` | 登录密码（勿提交、勿写入文档） |
| `NACOS_NAMESPACE` | 默认命名空间 ID |
| `NACOS_CONTEXT_PATH` | 默认 `/nacos` |

## 构建 / 测试

```powershell
$env:PATH = "D:\tools\go\bin;" + $env:PATH
$env:GOPROXY = "goproxy.cn,direct"
go test ./...
go build -o nacos.exe .
.\nacos.exe --help
# 或
.\nacos.ps1 --help
```

## 高风险写门禁（必读）

写操作必须带 **`--dry-run`**（只预览）或 **`--yes`/`-y`**（确认执行）。无门禁 → exit **10**。

| 写操作 | 额外默认 |
|--------|----------|
| `config publish` | **默认 create-only**；覆盖需 `--overwrite` |
| `config rollback` | 需 `--nid` |
| `service register` | 默认临时实例（ephemeral）；持久用 `--persistent` |
| `service deregister` / `up` / `down` | ephemeral 须与注册一致 |

## 命令一览

### 配置

```powershell
.\nacos.ps1 config get --data-id app.yaml --group DEFAULT_GROUP
.\nacos.ps1 config list --data-id "" --group DEFAULT_GROUP
.\nacos.ps1 config publish --data-id nacos-cli-smoke.yaml --group DEFAULT_GROUP --content "k: v" --dry-run
.\nacos.ps1 config publish --data-id nacos-cli-smoke.yaml --group DEFAULT_GROUP --content "k: v" --yes
.\nacos.ps1 config publish --data-id app.yaml --group DEFAULT_GROUP --file .\app.yaml --overwrite --dry-run
.\nacos.ps1 config history --data-id app.yaml --group DEFAULT_GROUP
.\nacos.ps1 config rollback --data-id app.yaml --group DEFAULT_GROUP --nid <historyId> --dry-run
.\nacos.ps1 config rollback --data-id app.yaml --group DEFAULT_GROUP --nid <historyId> --yes
```

### 服务实例

```powershell
.\nacos.ps1 service list
.\nacos.ps1 service instances --service my-service
.\nacos.ps1 service register --service nacos-cli-smoke-svc --ip 127.0.0.1 --port 19849 --dry-run
.\nacos.ps1 service register --service nacos-cli-smoke-svc --ip 127.0.0.1 --port 19849 --yes
.\nacos.ps1 service down --service nacos-cli-smoke-svc --ip 127.0.0.1 --port 19849 --yes
.\nacos.ps1 service up --service nacos-cli-smoke-svc --ip 127.0.0.1 --port 19849 --yes
.\nacos.ps1 service deregister --service nacos-cli-smoke-svc --ip 127.0.0.1 --port 19849 --yes
```

### 命名空间 / 鉴权 / skills

```powershell
.\nacos.ps1 namespace list
- user list|create|delete|passwd
- role list|bind|unbind
- permission list|grant|revoke
.\nacos.ps1 doctor
.\nacos.ps1 auth login --web
.\nacos.ps1 auth status
.\nacos.ps1 skills list
```

## 方法论

- **TDD + DDD + PDCA**（对齐 plumelog-cli）
- Agent 英文入口：[`AGENTS.md`](AGENTS.md)
- PDCA：[`docs/pdca.md`](docs/pdca.md)
- 中文 wiki：[`docs/wiki/Home.md`](docs/wiki/Home.md)

## 注意

- 不要把真实密码写进仓库、聊天或文档
- 联调优先造新 smoke dataId / 临时实例，不动现网配置
- `backups/` 勿删

## ACL (RBAC)

```powershell
.\nacos.ps1 user list
.\nacos.ps1 user create --username alice --password REDACTED --dry-run
.\nacos.ps1 user create --username alice --password REDACTED --yes
.\nacos.ps1 role bind --role dev --username alice --yes
.\nacos.ps1 permission grant --role dev --resource "*:*:*" --action r --yes
```

Requires auth-enabled Nacos + admin login. Passwords never appear in dry-run JSON.
