# AI 友好契约

## 原则

CLI 默认对 Agent 友好：机器可读、可预览、可重试；**不**提供无闸写通道。

## 信封

- 默认 ok-json：`{"ok":…}`
- 写操作 `--dry-run` → `would` 为 **数组** `[{action, …}]`，另有 `would_require`
- 失败带稳定 `error.code`（如 `confirmation_required` / `password_required` / `http_401`）
- 可选 `--trace`（或 `NACOS_TRACE=1`）→ 信封 `meta.duration_ms` / `renewed` / `update_checked`

## Agent 入口

| 命令 | 用途 |
|------|------|
| `nacos commands` / `--help-json` | 能力表 |
| `nacos examples [--command X]` | 成功/失败 JSON 样例 |
| `nacos doctor` | 含 `can_auto_login` |
| list `--limit` / `--query` | 分页别名与过滤 |

## 脱敏

manifest `capabilities.redact_fields`：password / token / accessToken 等永不回显。

## 无 TTY

非交互不 prompt；缺密码（已设 username）→ `password_required`。
