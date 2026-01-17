# AI原型快速部署工具 (v2.0)

一个为产品人员设计的快速部署工具,可以轻松将AI生成的HTML原型发布到服务器上,支持版本管理和回滚功能。

## 🎉 v2.0 重大更新

### 核心改进

1. **无需 Nginx** - Go 服务直接托管静态文件,开箱即用
2. **智能部署系统**
   - **增量部署**: 自动检测文件变更,只上传修改的文件
   - **全量部署**: 一键上传整个网站
3. **目录级部署** - 支持部署整个目录,不再是单个 HTML 文件
4. **自动文件追踪** - 客户端自动记录文件状态,智能选择最优部署方式

## 功能特性

- 🚀 **快速部署** - 支持增量/全量两种部署模式
- 🌐 **两种部署模式**
  - 子域名模式：每个网站使用不同子域名（如 site1.example.com）
  - 路径模式：所有网站共享域名,使用不同路径（如 example.com/site1）
- 📜 **版本管理** - 基于Git的版本控制,每次部署自动提交
- ⏮️ **版本回滚** - 快速恢复到任意历史版本
- 💻 **命令行工具** - 简单易用的 CLI 工具
- 📦 **智能压缩** - 使用 tar.gz 格式压缩传输,节省带宽

## 项目结构

```
aideploy/
├── server/                      # 服务端（Go）
│   ├── main.go                 # 服务入口
│   ├── deployer.go             # 核心部署逻辑
│   └── static.go               # 静态文件托管
│
├── client/                      # 客户端（Go）
│   ├── main.go                 # 命令行工具入口
│   ├── wails.go                # GUI 应用入口（Wails）
│   ├── deployer.go             # 部署器实现
│   └── frontend/               # GUI 前端（Vue 3）
│       ├── src/App.js          # 主应用组件
│       ├── package.json        # 前端依赖
│       └── ...
│
├── build-cli.bat/sh            # CLI 工具编译脚本
├── build-gui.bat/sh            # GUI 工具编译脚本
├── build-server.bat/sh         # 服务端编译脚本
└── README.md                   # 项目文档
```

## 快速开始

### 1. 服务端部署

#### 安装依赖

服务端需要Git环境（用于版本控制功能）：

**Windows:**
```bash
# 下载并安装 Git: https://git-scm.com/download/win
```

**Linux:**
```bash
sudo apt-get install git
```

**macOS:**
```bash
brew install git
```

#### 配置服务器

```bash
# 创建配置文件
go run server/main.go -init
```

这会创建 `config.json` 配置文件,编辑它：

```json
{
  "base_domain": "example.com",      // 基础域名（子域名模式）
  "web_root": "./websites",          // 网站根目录
  "mode": "subdomain",               // 部署模式: subdomain 或 path
  "single_domain": "",               // 单域名模式下的域名
  "port": 8080,                      // 服务端口（HTTP）
  "enable_versioning": true          // 是否启用版本控制
}
```

#### 启动服务

```bash
# 方式1: 直接运行
go run server/main.go

# 方式2: 编译后运行
go build -o deploy-server server/main.go
./deploy-server

# Windows
deploy-server.exe
```

服务启动后会显示：

```
服务器启动在 http://localhost:8080
部署模式: subdomain
基础域名: example.com
访问格式: http://site-name.example.com
网站目录: /path/to/websites
```

### 2. 客户端使用

项目提供两种客户端：**CLI 命令行工具** 和 **GUI 图形界面工具**

#### 方式一：CLI 命令行工具

**编译 CLI 工具**

```bash
# 使用编译脚本（推荐）
./build-cli.bat    # Windows
./build-cli.sh     # Linux/macOS

# 或手动编译
go build -o deploy-cli client/main.go

# Windows
go build -o deploy-cli.exe client/main.go
```

#### 方式二：GUI 图形界面工具

**安装 Wails CLI**

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

**编译 GUI 应用**

```bash
# 使用编译脚本（推荐）
./build-gui.bat    # Windows
./build-gui.sh     # Linux/macOS

# 或手动编译
cd client
wails build
```

编译完成后：
- **Windows**: `client/build/bin/AI原型部署工具.exe`
- **Linux/macOS**: `client/build/bin/AI原型部署工具`

**GUI 功能特性**

- 📁 创建和删除网站
- 📤 拖拽或选择文件进行部署
- 📜 查看版本历史
- ↩️ 一键回滚到任意版本
- 🎨 美观的图形界面

**使用 GUI 工具**

1. 启动应用：`AI原型部署工具.exe` (或对应平台的可执行文件)
2. 在界面中：
   - 输入网站名称，点击"创建网站"
   - 选择已创建的网站
   - 选择要部署的 HTML 文件或目录
   - 填写版本说明（可选）
   - 点击"部署"按钮

#### CLI 工具基本用法

```bash
# 创建网站
deploy-cli create my-prototype

# 部署网站（智能选择增量或全量）
deploy-cli deploy my-prototype ./dist

# 全量部署
deploy-cli deploy-full my-prototype ./dist

# 增量部署
deploy-cli deploy-inc my-prototype ./dist

# 查看所有网站
deploy-cli list

# 查看版本历史
deploy-cli versions my-prototype

# 回滚到指定版本
deploy-cli rollback my-prototype abc1234
```

## 部署模式说明

### 子域名模式 (subdomain)

每个网站使用独立的子域名：

```
my-prototype.example.com  -> websites/my-prototype/
another-site.example.com  -> websites/another-site/
```

配置：

```json
{
  "mode": "subdomain",
  "base_domain": "example.com"
}
```

**DNS 配置**：
需要配置通配符 DNS：`*.example.com` 指向服务器 IP

**访问方式**：
直接访问 `http://my-prototype.example.com`

### 路径模式 (path)

所有网站共享域名,使用不同路径：

```
example.com/my-prototype  -> websites/my-prototype/
example.com/another-site  -> websites/another-site/
```

配置：

```json
{
  "mode": "path",
  "single_domain": "example.com"
}
```

**访问方式**：
访问 `http://example.com/my-prototype`

## 部署方式详解

### 1. 智能部署 (推荐)

```bash
deploy-cli deploy my-prototype ./dist
```

- 自动检测是否有跟踪信息
- 首次部署使用全量模式
- 后续部署自动使用增量模式
- 只传输变更的文件,节省时间和带宽

### 2. 全量部署

```bash
deploy-cli deploy-full my-prototype ./dist
```

- 打包整个目录
- 上传所有文件
- 适用于首次部署或大量变更

### 3. 增量部署

```bash
deploy-cli deploy-inc my-prototype ./dist
```

- 只打包变更的文件
- 快速上传
- 适用于小幅度修改

## 文件追踪机制

客户端会在 `~/.aideploy/tracking/` 目录下为每个网站维护一个跟踪文件：

```json
{
  "site_name": "my-prototype",
  "last_sync": "2025-01-16T20:30:00Z",
  "files": [
    {
      "path": "index.html",
      "hash": "5d41402abc4b2a76b9719d911017c592",
      "size": 1024,
      "mod_time": "2025-01-16T20:00:00Z",
      "last_deployed": "2025-01-16T20:30:00Z"
    }
  ]
}
```

- 自动记录文件哈希值
- 检测文件变更
- 智能选择部署方式

## API 接口

服务端提供以下 REST API：

### 创建网站
```http
POST /api/sites/create
Content-Type: application/json

{
  "name": "my-prototype"
}
```

### 全量部署
```http
POST /api/sites/deploy-full
Content-Type: multipart/form-data

name: my-prototype
package: <tar.gz file>
message: 全量部署
```

### 增量部署
```http
POST /api/sites/deploy-incremental
Content-Type: multipart/form-data

name: my-prototype
package: <tar.gz file>
message: 增量部署
```

### 列出网站
```http
GET /api/sites/list
```

### 查看版本
```http
GET /api/sites/versions?name=my-prototype
```

### 回滚版本
```http
POST /api/sites/rollback
Content-Type: application/json

{
  "name": "my-prototype",
  "hash": "abc1234",
  "message": "回滚版本"
}
```

## 常见使用场景

### 场景1：AI 生成原型快速发布

```bash
# 1. AI 工具生成了 HTML 文件在 dist 目录
# 2. 创建网站
deploy-cli create prototype-v1

# 3. 部署
deploy-cli deploy prototype-v1 ./dist

# 4. 访问
# http://prototype-v1.example.com
```

### 场景2：迭代修改

```bash
# 修改了部分文件
# 重新部署（自动增量）
deploy-cli deploy prototype-v1 ./dist "修复导航栏问题"

# 如果有问题,回滚
deploy-cli versions prototype-v1
deploy-cli rollback prototype-v1 abc123
```

### 场景3：管理多个原型

```bash
# 创建多个网站
deploy-cli create prototype-a
deploy-cli create prototype-b
deploy-cli create prototype-c

# 分别部署
deploy-cli deploy prototype-a ./dist-a
deploy-cli deploy prototype-b ./dist-b
deploy-cli deploy prototype-c ./dist-c

# 查看所有
deploy-cli list
```

## 技术栈

- **服务端**: Go 1.21+ (标准库)
- **客户端**: Go 1.21+
- **版本控制**: Git
- **压缩**: tar.gz

## 安全建议

1. **生产环境部署**
   - 使用反向代理（如 Nginx）处理 HTTPS
   - 添加认证中间件
   - 限制文件大小
   - 验证文件类型

2. **权限控制**
   - 限制网站命名规则
   - 限制可部署文件类型
   - 定期备份

3. **防火墙配置**
   - 只开放必要的端口
   - 使用 HTTPS

## 性能优化

- **静态文件缓存**: 自动为 JS、CSS、图片等资源设置 1 年缓存
- **ETag 支持**: 自动生成 ETag,减少带宽消耗
- **增量部署**: 只传输变更文件
- **Gzip 压缩**: 部署包使用 gzip 压缩

## 故障排查

### 1. 端口被占用
```
错误: bind: address already in use
```

解决方法：
- 修改 `config.json` 中的 `port` 为其他端口
- 或停止占用该端口的程序

### 2. Git 未安装
```
错误: git: command not found
```

解决方法：
- Windows: 从 https://git-scm.com 下载安装
- Linux: `sudo apt-get install git`
- macOS: `brew install git`

### 3. 权限问题
```
错误: permission denied
```

解决方法：
- Linux/macOS: 使用 `chmod` 设置正确的权限
- Windows: 以管理员身份运行

### 4. 部署后无法访问

检查项：
1. 服务器是否正在运行
2. DNS 是否解析到正确的 IP
3. 防火墙是否开放端口
4. 配置文件中的域名是否正确

## 开发计划

- [ ] Web 管理后台
- [ ] 用户认证系统
- [ ] 访问统计
- [ ] 自动备份
- [ ] Docker 支持
- [ ] HTTPS 支持
- [ ] 多服务器支持

## 许可证

MIT License

## 贡献

欢迎提交 Issue 和 Pull Request！
