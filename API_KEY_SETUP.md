# API 密钥配置指南

## 概述

为了保护服务端安全，系统支持配置 API 密钥进行身份验证。当服务端配置了 API 密钥后，所有客户端请求都必须在请求头中携带正确的密钥才能访问。

## 配置步骤

### 1. 服务端配置

编辑服务端 `config.json` 文件：

```json
{
  "base_domain": "example.com",
  "web_root": "./websites",
  "mode": "subdomain",
  "single_domain": "",
  "port": 8080,
  "enable_versioning": true,
  "api_key": "your-secret-api-key-here"
}
```

**生成强密钥的建议方法**：

**Linux/macOS:**
```bash
# 使用 openssl 生成随机密钥
openssl rand -base64 32

# 或使用 /dev/urandom
head -c 32 /dev/urandom | base64
```

**Windows (PowerShell):**
```powershell
# 生成随机密钥
-guid | Select-Object -ExpandProperty Guid
```

### 2. 客户端配置

编辑客户端 `~/.aideploy/config.json` 文件：

**Windows:** `C:\Users\你的用户名\.aideploy\config.json`
**Linux/macOS:** `~/.aideploy/config.json`

```json
{
  "server_url": "http://your-server-ip:8080/api",
  "api_key": "your-secret-api-key-here"
}
```

**重要**：客户端的 `api_key` 必须与服务端配置的密钥完全一致。

## 工作原理

1. **服务端验证**：
   - 服务端在 `authMiddleware` 中检查每个 API 请求
   - 从请求头 `X-API-Key` 中获取密钥
   - 与配置文件中的 `api_key` 进行比对
   - 如果不匹配，返回 401 Unauthorized 错误

2. **客户端请求**：
   - CLI 和 GUI 客户端会自动从配置文件读取 `api_key`
   - 在所有 HTTP 请求中添加 `X-API-Key` 请求头
   - 如果密钥为空，则不添加该请求头

## 使用场景

### 场景 1：本地开发（无需密钥）

**服务端配置：**
```json
{
  "api_key": ""
}
```

**客户端配置：**
```json
{
  "api_key": ""
}
```

### 场景 2：团队协作（共享密钥）

**服务端配置：**
```json
{
  "api_key": "team-shared-key-2024"
}
```

团队成员的客户端配置：
```json
{
  "server_url": "http://deploy-server:8080/api",
  "api_key": "team-shared-key-2024"
}
```

### 场景 3：生产环境（独立密钥）

**服务端配置：**
```json
{
  "api_key": "Prod-Secure-Key-Kj8#mP2$vL9@xR5"
}
```

**客户端配置：**
```json
{
  "server_url": "https://deploy.example.com/api",
  "api_key": "Prod-Secure-Key-Kj8#mP2$vL9@xR5"
}
```

## 测试验证

### 1. 测试服务端是否启用密钥验证

```bash
# 不带密钥的请求
curl -X GET http://localhost:8080/api/sites/list

# 带密钥的请求
curl -X GET http://localhost:8080/api/sites/list -H "X-API-Key: your-secret-key"
```

**预期结果**：
- 如果服务端配置了密钥，第一个请求返回 `401 Unauthorized`
- 第二个请求（密钥正确）返回正常数据

### 2. 使用客户端测试

```bash
# 列出网站（测试连接）
deploy-cli list
```

**如果密钥配置错误**，会看到：
```
错误: 未授权：无效的API密钥
```

## 安全建议

1. **使用强密钥**：
   - 至少 16 个字符
   - 包含大小写字母、数字和特殊字符
   - 使用随机生成工具创建

2. **定期更换密钥**：
   - 建议每 3-6 个月更换一次
   - 更换时需同步更新所有客户端配置

3. **保护配置文件**：
   - 配置文件包含敏感信息，应设置适当的文件权限
   - 不要将配置文件提交到版本控制系统
   - 生产环境考虑使用环境变量或密钥管理服务

4. **传输安全**：
   - 生产环境建议使用 HTTPS
   - 避免在日志中记录密钥

## 故障排查

### 问题 1：始终返回 401 Unauthorized

**可能原因**：
1. 客户端和服务端的密钥不一致
2. 客户端配置文件路径错误
3. 客户端未正确读取配置

**解决方法**：
1. 检查服务端 `config.json` 中的 `api_key`
2. 检查客户端 `~/.aideploy/config.json` 中的 `api_key`
3. 确保两者完全一致（注意空格、大小写）
4. 使用 `-debug` 模式查看详细的错误信息

### 问题 2：不想使用密钥验证

**解决方法**：
将服务端和客户端配置中的 `api_key` 设置为空字符串：
```json
{
  "api_key": ""
}
```

### 问题 3：密钥在配置文件中明文存储

**建议**：
- 当前版本密钥以明文存储在配置文件中
- 生产环境建议：
  - 设置配置文件权限（仅用户可读）
  - 考虑使用环境变量存储密钥
  - 或使用专业的密钥管理服务

## 更新密钥

### 步骤：

1. **停止服务端**
2. **更新服务端配置**：
   ```bash
   vim config.json
   # 修改 api_key 字段
   ```
3. **重启服务端**
4. **通知所有团队成员更新客户端配置**
5. **测试连接**：
   ```bash
   deploy-cli list
   ```

## 示例配置

### 完整的服务端配置示例

```json
{
  "base_domain": "deploy.example.com",
  "web_root": "/var/www/deploy",
  "mode": "subdomain",
  "single_domain": "",
  "port": 8080,
  "enable_versioning": true,
  "api_key": "Ex@mpL3-K3y-2024-S3cur3"
}
```

### 完整的客户端配置示例

```json
{
  "server_url": "http://deploy.example.com:8080/api",
  "api_key": "Ex@mpL3-K3y-2024-S3cur3"
}
```
