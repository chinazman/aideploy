# AI原型快速部署工具 - 项目概览

## 📋 项目说明

这是一个专为产品人员设计的AI原型快速发布工具，让不懂服务器操作的产品人员也能轻松部署AI生成的HTML原型。

## 🏗️ 项目结构

```
aideploy/
├── server/                      # 服务端（Go）
│   ├── main.go                 # 服务入口，处理配置和启动
│   └── deployer.go             # 核心部署逻辑（1448行）
│       ├── 创建/删除网站
│       ├── 部署HTML文件
│       ├── Git版本管理
│       ├── 版本回滚
│       └── REST API接口
│
├── client/                      # 客户端（Go + Wails + Vue）
│   ├── main.go                 # 命令行工具（400行）
│   │   ├── create - 创建网站
│   │   ├── delete - 删除网站
│   │   ├── deploy - 部署文件
│   │   ├── list - 列出网站
│   │   ├── versions - 查看版本
│   │   └── rollback - 版本回滚
│   │
│   ├── wails.go                # Wails GUI应用后端
│   │   ├── Go绑定到前端
│   │   ├── API调用封装
│   │   └── 文件上传处理
│   │
│   ├── wails.json              # Wails配置
│   └── frontend/               # Vue.js前端界面
│       ├── index.html
│       ├── package.json
│       ├── vite.config.js
│       └── src/
│           ├── main.js
│           └── App.js          # 主应用组件（含样式）
│
├── cmd/                         # 构建脚本目录（预留）
│
├── config.example.json          # 配置文件示例
├── go.mod                       # Go模块定义
├── README.md                    # 项目文档
├── USAGE.md                     # 使用指南
│
├── build-server.sh/bat          # 编译服务端脚本
├── build-cli.sh/bat             # 编译客户端脚本
└── quick-start.sh/bat           # 快速启动脚本（编译+运行）
```

## 🎯 核心功能

### 服务端功能（[server/deployer.go](server/deployer.go)）

1. **网站管理**
   - 创建网站（自动初始化Git仓库）
   - 删除网站（支持确认）
   - 列出所有网站

2. **部署功能**
   - 上传HTML文件
   - 自动提取HTML中的资源（base64图片等）
   - Git自动提交

3. **版本管理**
   - 查看Git提交历史（最近20条）
   - 回滚到任意历史版本
   - 每次部署自动创建版本

4. **两种部署模式**
   - **子域名模式**: `site.example.com` → `websites/site/`
   - **路径模式**: `example.com/site` → `websites/site/`

5. **REST API**
   - `/api/sites/create` - 创建网站
   - `/api/sites/delete` - 删除网站
   - `/api/sites/deploy` - 部署文件
   - `/api/sites/list` - 列出网站
   - `/api/sites/versions` - 查看版本
   - `/api/sites/rollback` - 版本回滚

### 客户端功能

#### 命令行工具（[client/main.go](client/main.go)）
- 所有操作的CLI封装
- 交互式确认（删除、回滚）
- 友好的错误提示

#### GUI应用（[client/wails.go](client/wails.go) + [client/frontend/src/App.js](client/frontend/src/App.js)）
- Vue.js + Wails桌面应用
- 可视化操作界面
- 实时反馈
- 版本历史可视化

## 🚀 快速开始

### 第一次使用

1. **启动服务**
   ```bash
   # Windows
   quick-start.bat

   # Linux/macOS
   ./quick-start.sh
   ```

2. **编辑配置**（首次运行自动生成）
   ```json
   {
     "base_domain": "yourdomain.com",
     "web_root": "./websites",
     "mode": "subdomain",
     "port": 8080,
     "enable_versioning": true
   }
   ```

3. **开始部署**
   ```bash
   bin/deploy-cli create my-prototype
   bin/deploy-cli deploy my-prototype prototype.html
   ```

### 编译项目

```bash
# 编译服务端
./build-server.bat  # Windows
./build-server.sh   # Linux/macOS

# 编译CLI客户端
./build-cli.bat     # Windows
./build-cli.sh      # Linux/macOS

# 编译GUI应用（需要Wails）
cd client
wails build
```

## 📊 技术架构

### 服务端
- **语言**: Go 1.21+
- **框架**: 标准库 net/http
- **版本控制**: Git（系统依赖）
- **数据存储**: 文件系统 + Git

### 客户端
- **CLI**: Go标准库
- **GUI**: Wails v2 + Vue.js 3
- **构建**: Vite 4

### 通信
- **协议**: HTTP/REST
- **数据格式**: JSON
- **文件上传**: Multipart Form Data

## 🔧 配置说明

### 部署模式选择

**子域名模式**（适合多独立站点）
```json
{
  "mode": "subdomain",
  "base_domain": "example.com"
}
```
需要配置DNS: `*.example.com` → 服务器IP

**路径模式**（适合单域名多项目）
```json
{
  "mode": "path",
  "single_domain": "example.com"
}
```
所有网站共享一个域名

### 其他配置项

- `web_root`: 网站存储目录
- `port`: API服务端口（默认8080）
- `enable_versioning`: 是否启用Git版本控制

## 📝 API示例

### 创建网站
```bash
curl -X POST http://localhost:8080/api/sites/create \
  -H "Content-Type: application/json" \
  -d '{"name":"my-prototype"}'
```

### 部署文件
```bash
curl -X POST http://localhost:8080/api/sites/deploy \
  -F "name=my-prototype" \
  -F "file=@prototype.html" \
  -F "message=更新首页"
```

### 获取版本列表
```bash
curl http://localhost:8080/api/sites/versions?name=my-prototype
```

## 🔐 安全建议

1. **生产环境部署**
   - 添加认证中间件
   - 使用HTTPS
   - 限制文件大小
   - 验证文件类型

2. **权限控制**
   - 限制网站命名规则
   - 限制可部署文件类型
   - 定期备份

## 🛠️ 扩展功能建议

- [ ] Web管理后台
- [ ] 用户认证系统
- [ ] 访问统计
- [ ] 自动备份
- [ ] Let's Encrypt自动SSL
- [ ] Docker支持
- [ ] 数据库后端
- [ ] 多服务器支持

## 📄 文件说明

### 核心文件

| 文件 | 说明 | 行数 |
|------|------|------|
| [server/deployer.go](server/deployer.go) | 服务端核心逻辑 | ~450 |
| [server/main.go](server/main.go) | 服务入口 | ~100 |
| [client/main.go](client/main.go) | CLI工具 | ~400 |
| [client/wails.go](client/wails.go) | GUI后端 | ~200 |
| [client/frontend/src/App.js](client/frontend/src/App.js) | Vue前端 | ~500 |

### 配置文件

| 文件 | 说明 |
|------|------|
| [config.example.json](config.example.json) | 配置模板 |
| [go.mod](go.mod) | Go依赖 |
| [wails.json](client/wails.json) | Wails配置 |
| [package.json](client/frontend/package.json) | NPM依赖 |

### 文档文件

| 文件 | 说明 |
|------|------|
| [README.md](README.md) | 项目介绍和快速开始 |
| [USAGE.md](USAGE.md) | 详细使用指南 |
| [PROJECT.md](PROJECT.md) | 本文件，项目概览 |

## 📞 支持

如有问题，请查看：
1. [USAGE.md](USAGE.md) - 使用指南和故障排查
2. [README.md](README.md) - 项目文档
3. 提交Issue

## 📜 许可证

MIT License

---

**版本**: 1.0.0
**最后更新**: 2025-01-16
