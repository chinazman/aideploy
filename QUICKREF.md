# AI原型部署工具 - 快速参考

## 🚀 快速开始

### 服务端

```bash
# 1. 初始化配置
go run main.go -init

# 2. 编辑 config.json 配置文件

# 3. 启动服务器
go run main.go
```

### 客户端

#### CLI 命令行工具

```bash
# 编译
./build-cli.bat    # Windows
./build-cli.sh     # Linux/macOS

# 使用
./bin/deploy-cli create my-prototype
./bin/deploy-cli deploy my-prototype ./dist
```

#### GUI 图形界面工具

```bash
# 前置条件：安装 Wails
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# 编译
./build-gui.bat    # Windows
./build-gui.sh     # Linux/macOS

# 运行
./client/build/bin/AI原型部署工具.exe  # Windows
./client/build/bin/AI原型部署工具      # Linux/macOS
```

## 📋 常用命令

### CLI 工具命令

| 命令 | 说明 | 示例 |
|------|------|------|
| `create` | 创建网站 | `deploy-cli create my-site` |
| `deploy` | 智能部署 | `deploy-cli deploy my-site ./dist` |
| `deploy-full` | 全量部署 | `deploy-cli deploy-full my-site ./dist` |
| `deploy-inc` | 增量部署 | `deploy-cli deploy-inc my-site ./dist` |
| `list` | 列出网站 | `deploy-cli list` |
| `versions` | 查看版本 | `deploy-cli versions my-site` |
| `rollback` | 回滚版本 | `deploy-cli rollback my-site abc123` |
| `delete` | 删除网站 | `deploy-cli delete my-site` |

### GUI 工具功能

- ✅ 创建/删除网站
- ✅ 文件部署（支持拖拽）
- ✅ 查看版本历史
- ✅ 版本回滚
- ✅ 实时状态显示

## 🔧 配置文件

`config.json` 示例：

```json
{
  "base_domain": "example.com",
  "web_root": "./websites",
  "mode": "subdomain",
  "single_domain": "",
  "port": 8080,
  "enable_versioning": true
}
```

### 部署模式

**子域名模式** (subdomain)
- 每个网站独立子域名
- 需要配置通配符 DNS: `*.example.com`

**路径模式** (path)
- 所有网站共享域名
- 使用不同路径访问

## 📁 项目结构

```
aideploy/
├── server/          # 服务端
├── client/          # 客户端
│   └── frontend/   # GUI 前端
├── bin/             # 编译输出
└── websites/        # 部署的网站
```

## 🛠️ 编译脚本

| 脚本 | 说明 |
|------|------|
| `build-server.bat/sh` | 编译服务端 |
| `build-cli.bat/sh` | 编译 CLI 工具 |
| `build-gui.bat/sh` | 编译 GUI 工具 |

## 📞 获取帮助

- CLI: `deploy-cli help`
- README: [README.md](README.md)
- Issues: GitHub Issues
