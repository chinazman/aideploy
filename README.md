# AI原型快速部署工具

一个为产品人员设计的快速部署工具，可以轻松将AI生成的HTML原型发布到服务器上，支持版本管理和回滚功能。

## 功能特性

- 🚀 **快速部署** - 一键上传HTML文件并发布
- 🌐 **两种部署模式**
  - 子域名模式：每个网站使用不同子域名（如 site1.example.com）
  - 路径模式：所有网站共享域名，使用不同路径（如 example.com/site1）
- 📜 **版本管理** - 基于Git的版本控制，每次部署自动提交
- ⏮️ **版本回滚** - 快速恢复到任意历史版本
- 💻 **双端支持** - 图形界面客户端 + 命令行工具

## 项目结构

```
aideploy/
├── server/           # Go服务端
│   ├── main.go      # 服务入口
│   └── deployer.go  # 核心部署逻辑
├── client/          # 客户端
│   ├── main.go      # CLI工具
│   ├── wails.go     # Wails GUI应用
│   └── frontend/    # Vue前端界面
└── cmd/             # 构建脚本
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
cd server
go run main.go -init
```

这会创建 `config.json` 配置文件，编辑它：

```json
{
  "base_domain": "yourdomain.com",     // 基础域名（子域名模式）
  "web_root": "./websites",             // 网站根目录
  "mode": "subdomain",                  // 部署模式: subdomain 或 path
  "single_domain": "",                  // 单域名模式下的域名
  "port": 8080,                         // API服务端口
  "enable_versioning": true             // 是否启用版本控制
}
```

#### 启动服务

```bash
cd server
go run main.go
```

或编译后运行：

```bash
go build -o deploy-server main.go
./deploy-server
```

### 2. 客户端使用

#### 命令行工具

```bash
# 编译CLI工具
cd client
go build -o deploy-cli main.go

# 创建网站
./deploy-cli create my-prototype

# 部署网站
./deploy-cli deploy my-prototype prototype.html

# 查看所有网站
./deploy-cli list

# 查看版本历史
./deploy-cli versions my-prototype

# 回滚到指定版本
./deploy-cli rollback my-prototype abc1234
```

#### 图形界面（需要Wails）

**安装Wails:**

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

**运行开发版本:**

```bash
cd client
wails dev
```

**构建可执行文件:**

```bash
wails build
```

## API接口

服务端提供以下REST API：

### 创建网站
```http
POST /api/sites/create
Content-Type: application/json

{
  "name": "my-prototype"
}
```

### 删除网站
```http
POST /api/sites/delete
Content-Type: application/json

{
  "name": "my-prototype"
}
```

### 部署网站
```http
POST /api/sites/deploy
Content-Type: multipart/form-data

name: my-prototype
file: prototype.html
message: 更新首页设计
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

需要配置通配符DNS：`*.example.com` 指向服务器IP

### 路径模式 (path)

所有网站共享域名，使用不同路径：

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

## Nginx配置示例

### 子域名模式

```nginx
server {
    listen 80;
    server_name *.example.com;

    # 提取子域名
    set $subdomain "default";
    if ($host ~* "^([a-z0-9-]+)\.example\.com$") {
        set $subdomain $1;
    }

    # 设置网站根目录
    root /path/to/websites/$subdomain;

    location / {
        try_files $uri $uri/ /index.html;
    }
}
```

### 路径模式

```nginx
server {
    listen 80;
    server_name example.com;

    location / {
        root /path/to/websites;
        try_files $uri $uri/ /index.html;
    }
}
```

## 开发计划

- [ ] Web界面管理后台
- [ ] 多用户支持
- [ ] 访问统计
- [ ] 自动备份
- [ ] SSL证书自动配置
- [ ] Docker部署支持

## 技术栈

- **服务端**: Go + 标准库
- **客户端**: Go + Wails + Vue.js
- **版本控制**: Git

## 许可证

MIT License

## 贡献

欢迎提交Issue和Pull Request！
