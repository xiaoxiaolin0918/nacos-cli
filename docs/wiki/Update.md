# Update / 无感检查

## 一句话

`nacos update --check` 查 GitHub Releases；`nacos update --yes` 校验 SHA-256 后替换本机二进制。后台检查与静默升级**默认关**。

## 命令

```text
nacos update --check          # 只比较，不下载
nacos update --yes            # 下载 + 校验 + 替换（缺 --yes 等同 --check）
nacos update --yes --dry-run  # 预览将替换的资产
```

资产名：`nacos-<os>-<arch>[.exe]`。校验优先 `checksums.txt`，否则用 Release asset 的 `digest`（sha256）。

## 后台检查（默认关）

开启任一即可：

| 方式 | 值 |
|------|-----|
| flag | `--check-update` |
| env | `NACOS_CHECK_UPDATE=1` |
| profile | `check_update: true` |

行为：命令结束后机会性检查（缓存 ≤24h、短超时）；有新版只写 **stderr** 一行提示，**不**污染 JSON stdout。`doctor --check-update` 同样只 stderr 提示。

## 静默自动替换（默认关，独立开关）

| 方式 | 值 |
|------|-----|
| flag | `--auto-update` |
| env | `NACOS_AUTO_UPDATE=1` |
| profile | `auto_update: true` |

仅当检查到新版且此开关打开时才下载替换；失败回滚并 stderr 报错。打开 auto-update 会隐含开启 check-update。

## Agent 能力表

```text
nacos commands
nacos --help-json
```

## 错误 code

失败信封 `error.code` 稳定字符串，见 [Errors.md](Errors.md)。
