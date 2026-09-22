---
name: aideploy
description: AI 原型快速部署工具。通过 aideploy 服务器实现网站绑定、发布/部署静态网站、版本管理与回滚。当用户要求"绑定网站、发布、部署、上线、更新网站、回滚、查看版本、下载网站文件"等操作时使用本技能。
---

# aideploy 网站部署

通过 Python 脚本与本项目的 aideploy 服务器交互，实现网站绑定与发布。**不依赖 Go 客户端**，脚本仅使用 Python 3 标准库。

脚本位置（相对于本 skill 目录）: `scripts/aideploy.py`

## 前置检查

执行任何操作前，先运行：

```bash
python "<skill目录>/scripts/aideploy.py" config --json
```

- 如果 `server_url` 或用户名/密码未配置，向用户询问服务器地址、用户名、密码，然后用 `config-set` 写入配置。
- 配置完后可用 `test` 命令验证连接与认证是否正常。

配置文件位于 `~/.aideploy/config.json`（与 Go 客户端共用，可互操作），格式：

```json
{
  "server_url": "http://localhost:8080/api",
  "username": "admin",
  "password": "admin123",
  "site_paths": { "网站名": "本地目录绝对路径" }
}
```

## 命令参考

所有命令都支持 `--json` 参数输出结构化结果（便于解析）。出错时脚本以非零退出码输出 `{"ok": false, "error": "..."}` 到 stderr。

| 操作 | 命令 |
|---|---|
| 查看配置 | `config [--json]` |
| 设置服务器地址 | `config-set server <url>` |
| 设置用户名/密码 | `config-set username <name>` / `config-set password <pwd>` |
| 跳过 TLS 证书校验 | `config-set insecure true`（自签名/过期证书的服务器需要） |
| **绑定网站目录** | `config-set site <网站名> <本地目录>` |
| 移除绑定 | `config-remove <网站名>` |
| 测试连接 | `test` |
| 创建网站 | `create <网站名> [--desc 描述]` |
| 删除网站 | `delete <网站名> --yes` |
| 更新描述/授权用户 | `update <网站名> [--desc 描述] [--users 用户1,用户2]` |
| 列出网站 | `list [--json]` |
| **发布（智能增量/全量）** | `deploy <网站名> [-m 说明] [--dir 目录] [--mode auto\|full\|incremental]` |
| 查看版本历史 | `versions <网站名> [--json]` |
| 回滚 | `rollback <网站名> <版本hash> --yes [-m 说明]` |
| 从服务器下载覆盖本地 | `pull <网站名> [--dir 目录]` |

> 注意：脚本使用的是 `config-set` / `config-remove`（连字符），不是 `config set`。

## 常见对话场景

### 1. 绑定网站（"帮我把 ./dist 绑定到网站 xxx 并发布"）

```bash
# 网站不存在时先创建
python aideploy.py create xxx --desc "描述"
# 绑定本地目录（目录必须已存在）
python aideploy.py config-set site xxx F:/work/demo/dist
# 发布
python aideploy.py deploy xxx -m "首次发布"
```

### 2. 发布/更新（"发布我的改动"）

```bash
python aideploy.py deploy <网站名> -m "本次改动说明"
```

- 默认 `--mode auto`：无跟踪信息时自动走全量，有跟踪信息时走增量（只上传 MD5 有变化的文件）。
- `--mode full` 强制全量；`--mode incremental` 强制增量。
- 成功后本地 `~/.aideploy/tracking/<网站名>.json` 会记录文件状态。

### 3. 版本管理与回滚

```bash
python aideploy.py versions <网站名> --json   # 拿到 hash 列表
python aideploy.py rollback <网站名> <hash> --yes -m "回滚到 xx"
```

### 4. 下载线上版本到本地

```bash
python aideploy.py pull <网站名>   # 会清空本地目录（保留隐藏文件）后解压
```

## 安全规则

1. **删除网站、回滚版本、pull 覆盖本地**是破坏性操作：执行前必须向用户确认（`delete`/`rollback` 脚本层面也强制要求 `--yes`）。
2. **发布（deploy）**：用户明确要求发布时可直接执行；如果用户只是问问题或意图不明，先确认。
3. 首次为新用户配置密码时，如果用户没有提供凭据，先询问，不要猜测。

## 实现说明

- 认证：请求头 `X-Username` + `X-Password`（或 `X-API-Key`）。
- 打包格式：tar.gz，包内路径为 POSIX 风格相对路径，跳过 `.` 开头的隐藏文件/目录。
- 服务端 API（base 为 `server_url`，如 `http://localhost:8080/api`）：
  - `POST /sites/create`（JSON: name, desc）
  - `POST /sites/delete`（JSON: name）
  - `POST /sites/update`（JSON: name, desc, users）
  - `GET  /sites/list`
  - `POST /sites/deploy-full`（multipart: name, message, package=tar.gz）
  - `POST /sites/deploy-incremental`（同上）
  - `GET  /sites/versions?name=`
  - `POST /sites/rollback`（JSON: name, hash, message）
  - `GET  /sites/export?name=`（返回 tar.gz）
