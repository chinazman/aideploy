# CLI 命令快速参考

## 配置命令

### 设置服务器地址
```bash
deploy-cli config set server http://192.168.1.100:8080/api
```

### 设置 API 密钥
```bash
deploy-cli config set api-key your-secret-key
```

### 设置网站发布目录
```bash
deploy-cli config set site my-prototype ./dist
```

**注意**：
- 设置网站目录后，部署时无需指定路径
- 目录会自动转换为绝对路径
- 如果目录不存在，会提示确认

### 清除 API 密钥（不使用密钥验证）
```bash
deploy-cli config set api-key ""
```

### 移除网站目录配置
```bash
deploy-cli config remove site my-prototype
```

### 查看当前配置
```bash
deploy-cli config get
```

**输出示例**：
```
当前配置:
------------------------------------------------------------
服务器地址: http://192.168.1.100:8080/api
API密钥:    your****-key

网站发布目录:
  my-prototype         -> F:\project\dist
  my-project           -> /path/to/project/dist
------------------------------------------------------------

配置文件位置: C:\Users\YourName\.aideploy\config.json
```

---

## 网站管理命令

### 创建网站
```bash
deploy-cli create my-website
```

### 删除网站
```bash
deploy-cli delete my-website
```

### 列出所有网站
```bash
deploy-cli list
```

---

## 部署命令

### 智能部署（自动选择增量或全量）
```bash
# 如果已配置网站目录，直接部署
deploy-cli deploy my-website

# 或者临时指定目录
deploy-cli deploy my-website ./dist
```

### 全量部署
```bash
deploy-cli deploy-full my-website
```

### 增量部署
```bash
deploy-cli deploy-inc my-website
```

### 带版本说明的部署
```bash
deploy-cli deploy my-website "修复了导航栏问题"
```

**说明**：
- 如果配置了网站目录，可以省略路径参数
- 也可以在命令中临时指定路径，覆盖配置
- 如果未配置目录，必须在命令中指定路径

---

## 版本管理命令

### 查看版本历史
```bash
deploy-cli versions my-website
```

### 回滚到指定版本
```bash
deploy-cli rollback my-website abc1234
```

### 带说明的回滚
```bash
deploy-cli rollback my-website abc1234 "回滚到上一个稳定版本"
```

---

## 帮助命令

### 查看帮助
```bash
deploy-cli help
```

### 查看配置命令帮助
```bash
deploy-cli config
```

---

## 完整使用流程示例

### 1. 首次配置
```bash
# 设置服务器地址
deploy-cli config set server http://192.168.1.100:8080/api

# 设置 API 密钥（如果服务端配置了密钥）
deploy-cli config set api-key my-secret-key

# 验证配置
deploy-cli config get
```

### 2. 创建并部署网站
```bash
# 创建网站
deploy-cli create my-project

# 首次部署（自动全量）
deploy-cli deploy my-project ./dist "首次部署"

# 修改后重新部署（自动增量）
deploy-cli deploy my-project ./dist "更新首页样式"
```

### 3. 版本管理
```bash
# 查看所有版本
deploy-cli versions my-project

# 发现问题，回滚到上一个版本
deploy-cli rollback my-project abc1234 "回滚：新版本有bug"
```

---

## 常见场景

### 场景 1：连接到不同服务器
```bash
# 开发环境
deploy-cli config set server http://localhost:8080/api

# 测试环境
deploy-cli config set server http://test-server:8080/api

# 生产环境
deploy-cli config set server https://deploy.example.com/api
```

### 场景 2：多服务器部署
```bash
# 部署到测试服务器
deploy-cli config set server http://test-server:8080/api
deploy-cli deploy my-project ./dist

# 部署到生产服务器
deploy-cli config set server http://prod-server:8080/api
deploy-cli deploy my-project ./dist
```

### 场景 3：快速测试
```bash
# 不使用密钥的本地测试
deploy-cli config set server http://localhost:8080/api
deploy-cli config set api-key ""
deploy-cli deploy test-project ./dist
```

---

## 故障排查

### 问题：连接服务器失败
```bash
# 检查配置
deploy-cli config get

# 测试服务器连接
curl http://your-server:8080/api/sites/list
```

### 问题：密钥验证失败
```bash
# 检查密钥是否正确
deploy-cli config get

# 重新设置密钥
deploy-cli config set api-key correct-key
```

### 问题：部署失败
```bash
# 使用全量部署重试
deploy-cli deploy-full my-project ./dist

# 查看详细错误信息
deploy-cli deploy my-project ./dist 2>&1 | tee error.log
```

---

## 提示和技巧

1. **使用别名**（Linux/macOS）：
   ```bash
   alias deploy='deploy-cli'
   alias dp='deploy-cli deploy'
   alias dps='deploy-cli deploy-full'
   ```

2. **Shell 自动完成**（可以添加到你的 shell 配置）：
   ```bash
   # bash completion for deploy-cli (示例)
   _deploy_cli_completion() {
       local cur=${COMP_WORDS[COMP_CWORD]}
       COMPREPLY=($(compgen -W "create delete deploy deploy-full deploy-inc list versions rollback config help" -- $cur))
   }
   complete -F _deploy_cli_completion deploy-cli
   ```

3. **批量部署脚本**：
   ```bash
   #!/bin/bash
   # deploy-all.sh
   deploy-cli config set server http://server1:8080/api
   deploy-cli deploy site1 ./dist1
   deploy-cli deploy site2 ./dist2
   ```

4. **配置切换脚本**：
   ```bash
   # switch-env.sh
   case $1 in
       dev)
           deploy-cli config set server http://localhost:8080/api
           ;;
       test)
           deploy-cli config set server http://test-server:8080/api
           deploy-cli config set api-key test-key
           ;;
       prod)
           deploy-cli config set server https://deploy.example.com/api
           deploy-cli config set api-key prod-key
           ;;
   esac
   echo "已切换到 $1 环境"
   deploy-cli config get
   ```

---

## 命令参数速查表

| 命令 | 参数 | 说明 |
|------|------|------|
| `config set server` | `<url>` | 设置服务器地址 |
| `config set api-key` | `<key>` | 设置 API 密钥 |
| `config get` | - | 查看当前配置 |
| `create` | `<name>` | 创建新网站 |
| `delete` | `<name>` | 删除网站 |
| `deploy` | `<name> <dir> [msg]` | 智能部署 |
| `deploy-full` | `<name> <dir> [msg]` | 全量部署 |
| `deploy-inc` | `<name> <dir> [msg]` | 增量部署 |
| `list` | - | 列出所有网站 |
| `versions` | `<name>` | 查看版本历史 |
| `rollback` | `<name> <hash> [msg]` | 回滚版本 |
| `help` | - | 显示帮助 |
