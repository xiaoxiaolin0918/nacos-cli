# Error codes

JSON 失败信封：

```json
{"ok":false,"error":{"type":"cli","code":"server_unset","message":"..."}}
```

| code | 含义 | 典型 exit |
|------|------|-----------|
| `confirmation_required` | 高风险写缺 `--yes` | 10 |
| `server_unset` | 未配置 `NACOS_SERVER` | 1 或 2 |
| `http_401` / `http_403` / `http_N` | 上游 HTTP 状态 | 1 |
| `cli_error` | 其它 CLI 错误 | 1 |
| `doctor_failed` | doctor 检查未通过 | 1 |
| `update_failed` | update 检查/替换失败 | 1 |

成功升级与 dry-run 仍走 `ok:true` 信封。
