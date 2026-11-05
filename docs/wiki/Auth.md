# Auth / 自动续期

## 一句话

`auth login --remember` 把密码加密进 `login.cred`；之后开 `--auto-login`（或 `NACOS_AUTO_LOGIN=1`），遇 **401**（或带 token 过期语义的 403）静默重登一次并只写回 token。纯 ACL 403 **不**触发续期、**不**清存密。

## 七条口径

1. `--remember` / `--store-password`：**显式 opt-in** 才写加密存密；默认关。
2. `--auto-login` / `--auto-renew` / `NACOS_AUTO_LOGIN`：**默认关**；`login.cred` 里 `auto_login: true` 时等同打开（对齐 plumelog）。
3. 密码落盘：加密 `login.cred`（**Win=DPAPI**；**macOS/Linux=AES-GCM + login.key 0600**）。不进 conf / 审计 / git / 信封。不用 Credential Manager / Keychain。
4. 续期：只读与写路径（含 `config publish/delete`、`instance register/deregister/enabled`、ACL mutate）统一 `doWithRenew`；每 Client **只重登一次**；失败 **透出** `auto-login failed: …; run: nacos auth login --remember`。
5. `auth logout`：**只清 token**（`credentials.json`）；**默认不清** `login.cred`。
6. `auth logout --forget` / `--clear-password`：**双清**——credentials 目录的 `login.cred`/`login.key`，以及当前 `--profile` 的 `~/.nacos-cli/.login/<name>/`（监管 Web 存密）。
7. Profile YAML **禁止**明文密码（`profile set --password` 拒绝；load 忽略、save 剥离）。Web 取消「记住」会 `ClearLogin`。

## 文件

| 文件 | 位置 | 内容 |
|------|------|------|
| `credentials.json` | OS config dir（`NACOS_CREDENTIALS_PATH` 可覆盖） | accessToken，无密码 |
| `login.cred` | 与 credentials.json 同目录 | 加密密码 + `auto_login` |
| `login.key` | 同上（非 Windows） | AES-256 本地密钥，0600 |
| `.login/<profile>/login.cred` | `~/.nacos-cli/.login/<name>/` | 监管 Web 按 profile 加密存密 |

## 旗标

| 旗标 / 环境变量 | 默认 | 作用 |
|-----------------|------|------|
| `auth login --remember` / `--store-password` | off | 写入 login.cred |
| Web 表单取消「记住」 | — | Clear 已有 login.cred |
| `--auto-login` / `--auto-renew` / `NACOS_AUTO_LOGIN=1` | off | 401（token 过期语义）静默重登一次并重放 |
| `auth logout` | — | 只清 credentials.json |
| `auth logout --forget` / `--clear-password` | — | 双清 login.cred + 当前 profile `.login/<name>` |
| `auth logout --dry-run` | — | 预览将清项 |

## 凭证原子源

静默登录取密顺序（成对，不混用）：

1. `NACOS_USERNAME` **且** `NACOS_PASSWORD`
2. 否则整包 `login.cred`

坏密（login HTTP 401 / 明显口令错误）：清 login.cred 并关闭本进程 auto-login。ACL 权限不足不清存密。

## status / doctor

脱敏字段：`password_stored` / `auto_renew` / `can_auto_login` / `remembered.auto_login_pref`；永不打印密码或裸 token。
