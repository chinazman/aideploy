# 网站目录映射功能使用指南

## 功能概述

网站目录映射功能允许你预先配置每个网站的本地发布目录，之后部署时无需每次都指定路径，只需提供网站名称即可。

## 使用场景

这个功能特别适合：
- 经常部署同一项目的固定目录
- 管理多个不同项目
- 简化日常部署流程

## 配置步骤

### 1. 首次配置

```bash
# 配置网站发布目录
deploy-cli config set site my-prototype ./dist

# 查看配置
deploy-cli config get
```

**输出示例**：
```
✓ 网站 'my-prototype' 的发布目录已设置为: F:\projects\my-prototype\dist
  现在可以使用 'deploy-cli deploy my-prototype' 直接部署
```

### 2. 日常使用

配置完成后，部署变得非常简单：

```bash
# 直接部署（无需指定路径）
deploy-cli deploy my-prototype

# 带说明的部署
deploy-cli deploy my-prototype "更新了首页样式"

# 全量部署
deploy-cli deploy-full my-prototype

# 增量部署
deploy-cli deploy-inc my-prototype
```

### 3. 管理多个项目

```bash
# 配置多个项目
deploy-cli config set site project-a ./projects/a/dist
deploy-cli config set site project-b ./projects/b/dist
deploy-cli config set site project-c ./projects/c/dist

# 查看所有配置
deploy-cli config get
```

**输出**：
```
当前配置:
------------------------------------------------------------
服务器地址: http://192.168.1.100:8080/api
API密钥:    my-secret-key

网站发布目录:
  project-a             -> F:\projects\a\dist
  project-b             -> F:\projects\b\dist
  project-c             -> F:\projects\c\dist
------------------------------------------------------------
```

快速部署任何项目：
```bash
deploy-cli deploy project-a
deploy-cli deploy project-b
deploy-cli deploy project-c
```

## 高级用法

### 临时覆盖配置

即使配置了默认目录，你仍然可以临时指定其他目录：

```bash
# 使用配置的目录
deploy-cli deploy my-prototype

# 临时使用其他目录
deploy-cli deploy my-prototype ./test-dist
```

### 移除配置

```bash
# 移除某个网站的目录配置
deploy-cli config remove site my-prototype

# 再次部署时需要指定路径
deploy-cli deploy my-prototype ./dist
```

### 路径自动转换

配置的路径会自动转换为绝对路径：

```bash
# 配置相对路径
deploy-cli config set site my-site ./dist

# 自动转换为绝对路径保存
✓ 网站 'my-site' 的发布目录已设置为: F:\work\my-project\dist
```

这意味即使你从不同的目录执行命令，也能正常工作。

## 完整工作流程示例

### 场景：前端开发者日常部署

#### 1. 项目初始化（只需一次）

```bash
# 1. 在服务端创建网站
deploy-cli create my-project

# 2. 配置本地发布目录
deploy-cli config set site my-project ./dist

# 3. 首次部署
deploy-cli deploy my-project "初始化项目"
```

#### 2. 日常开发迭代

```bash
# 修改代码...
# 构建项目...

# 部署更新（一条命令搞定）
deploy-cli deploy my-project "修复导航栏bug"
```

#### 3. 多项目切换

```bash
# 项目A的更新
npm run build:project-a
deploy-cli deploy project-a

# 项目B的更新
npm run build:project-b
deploy-cli deploy project-b

# 项目C的更新
npm run build:project-c
deploy-cli deploy project-c
```

## 常见问题

### Q: 如果目录不存在怎么办？

A: 系统会提示确认是否保存路径：

```bash
$ deploy-cli config set site my-site ./dist
警告: 目录不存在: ./dist
是否仍然保存此路径？(y/N):
```

- 输入 `y` 仍然保存（适合尚未构建的项目）
- 输入 `N` 或直接回车取消操作

### Q: 可以修改已配置的目录吗？

A: 可以，重新执行 `config set site` 命令即可覆盖：

```bash
# 原配置
deploy-cli config set site my-site ./old-dist

# 修改为新路径
deploy-cli config set site my-site ./new-dist
```

### Q: 配置文件在哪里？

A: 配置文件位于：
- **Windows**: `C:\Users\你的用户名\.aideploy\config.json`
- **Linux/macOS**: `~/.aideploy/config.json`

你也可以通过以下命令查看：
```bash
deploy-cli config get
```

### Q: 如何清空所有配置？

A: 直接删除配置文件即可：

```bash
# Windows
del %USERPROFILE%\.aideploy\config.json

# Linux/macOS
rm ~/.aideploy/config.json
```

或者手动删除文件后重新配置。

### Q: 团队协作怎么办？

A: 配置文件是本地的，每个开发者需要单独配置。建议：

1. 将配置命令加入团队文档
2. 新成员入职时执行配置脚本

示例配置脚本：
```bash
#!/bin/bash
# setup.sh - 新成员配置脚本

deploy-cli config set server http://deploy-server:8080/api
deploy-cli config set api-key team-secret-key
deploy-cli config set site project-a ./projects/a/dist
deploy-cli config set site project-b ./projects/b/dist

echo "✓ 配置完成！"
deploy-cli config get
```

## 最佳实践

### 1. 项目结构建议

```
my-website/
├── src/           # 源代码
├── dist/          # 构建输出（配置这个目录）
├── package.json
└── build.sh       # 构建脚本
```

配置命令：
```bash
deploy-cli config set site my-website ./dist
```

### 2. 构建脚本集成

在项目的 `package.json` 中添加部署命令：

```json
{
  "scripts": {
    "build": "vite build",
    "deploy": "npm run build && deploy-cli deploy my-website",
    "deploy:prod": "npm run build && deploy-cli deploy my-website 生产环境部署"
  }
}
```

使用：
```bash
npm run deploy
```

### 3. 多环境配置

```bash
# 开发环境
deploy-cli config set site project-dev ./dist-dev

# 测试环境
deploy-cli config set site project-test ./dist-test

# 生产环境
deploy-cli config set site project-prod ./dist-prod
```

快速切换环境部署：
```bash
deploy-cli deploy project-dev
deploy-cli deploy project-test
deploy-cli deploy project-prod
```

### 4. 版本控制

**建议**：将配置命令加入项目的 README.md：

```markdown
## 部署

首次使用请先配置：
```bash
deploy-cli config set site my-project ./dist
```

日常部署：
```bash
npm run build
deploy-cli deploy my-project
```
```

**不要**：将配置文件提交到 Git（添加到 `.gitignore`）

## 配置文件格式

配置文件（`~/.aideploy/config.json`）示例：

```json
{
  "server_url": "http://192.168.1.100:8080/api",
  "api_key": "my-secret-key",
  "site_paths": {
    "my-prototype": "F:\\projects\\my-prototype\\dist",
    "project-a": "/home/user/projects/a/dist",
    "project-b": "/home/user/projects/b/dist"
  }
}
```

## 总结

使用网站目录映射功能后，部署流程从：

```bash
# 之前：每次都要输入完整路径
deploy-cli deploy my-prototype /path/to/my-prototype/dist
```

简化为：

```bash
# 之后：一条命令搞定
deploy-cli deploy my-prototype
```

大大提高了日常部署效率！
