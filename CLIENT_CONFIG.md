# 客户端配置说明

## 配置文件位置

客户端配置文件位于用户目录下的 `.aideploy` 文件夹中：

- **Windows**: `C:\Users\你的用户名\.aideploy\config.json`
- **Linux/macOS**: `~/.aideploy/config.json`

## 配置文件格式

```json
{
  "server_url": "http://your-server-ip:8080/api",
  "api_key": "your-secret-key"
}
```

### 配置项说明

| 字段 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| server_url | string | 否 | http://localhost:8080/api | 服务端 API 地址 |
| api_key | string | 否 | （空） | API 密钥，需与服务端配置一致 |

## 配置步骤

### 方式一：使用命令行配置（推荐）

这是最简单的方法，无需手动创建文件：

```bash
# 设置服务器地址
deploy-cli config set server http://192.168.1.100:8080/api

# 设置API密钥
deploy-cli config set api-key your-secret-key

# 查看当前配置
deploy-cli config get
```

**命令说明**：
- `config set server <url>` - 设置服务端 API 地址
- `config set api-key <key>` - 设置 API 密钥
- `config get` - 查看当前配置（密钥会部分隐藏以保护安全）

**示例输出**：
```
$ deploy-cli config get

当前配置:
--------------------------------------------------
服务器地址: http://192.168.1.100:8080/api
API密钥:    your****-key
--------------------------------------------------

配置文件位置: C:\Users\YourName\.aideploy\config.json

提示: 使用 'deploy-cli config set <key> <value>' 修改配置
```

### 方式二：手动创建配置文件

如果你更喜欢手动编辑配置文件：

#### 1. 创建配置目录

**Linux/macOS:**
```bash
mkdir -p ~/.aideploy
```

**Windows (CMD):**
```cmd
mkdir %USERPROFILE%\.aideploy
```

**Windows (PowerShell):**
```powershell
New-Item -ItemType Directory -Path "$env:USERPROFILE\.aideploy" -Force
```

#### 2. 创建配置文件

**Linux/macOS:**
```bash
cat > ~/.aideploy/config.json << EOF
{
  "server_url": "http://your-server-ip:8080/api",
  "api_key": "your-secret-key"
}
EOF
```

**Windows:**

使用记事本或其他编辑器创建文件：
```cmd
notepad %USERPROFILE%\.aideploy\config.json
```

然后输入以下内容：
```json
{
  "server_url": "http://192.168.1.100:8080/api",
  "api_key": "your-secret-key"
}
```

### 3. 替换服务端地址

将配置文件中的 `your-server-ip` 替换为实际的服务端 IP 地址或域名。

例如：
- 本地服务器：`http://localhost:8080/api`
- 局域网服务器：`http://192.168.1.100:8080/api`
- 公网服务器：`http://your-domain.com/api`
- HTTPS（如果配置了SSL）：`https://your-domain.com/api`

### 4. 配置 API 密钥（可选）

如果服务端配置了 `api_key`，客户端也需要配置相同的密钥：

```json
{
  "server_url": "http://192.168.1.100:8080/api",
  "api_key": "my-secret-api-key-12345"
}
```

**注意**：
- 如果服务端设置了 `api_key`，客户端必须提供相同的密钥才能访问
- 如果服务端未设置 `api_key` 或为空，客户端可以留空或不配置此字段
- 建议生产环境使用强密钥（随机字符串至少16位）

## 注意事项

1. **路径要求**：`server_url` 必须以 `/api` 结尾
2. **协议支持**：支持 `http://` 和 `https://` 协议
3. **默认值**：如果配置文件不存在，客户端会使用 `http://localhost:8080/api`
4. **端口配置**：确保服务端端口（默认 8080）已开放且可访问
5. **防火墙**：如果服务端在其他机器上，确保防火墙允许对应端口的访问
6. **密钥匹配**：客户端的 `api_key` 必须与服务端配置的 `api_key` 完全一致

## 配置示例

### 示例 1：本地开发环境（无密钥）
```json
{
  "server_url": "http://localhost:8080/api",
  "api_key": ""
}
```

### 示例 2：局域网服务器（有密钥）
```json
{
  "server_url": "http://192.168.1.100:8080/api",
  "api_key": "my-secret-key-12345"
}
```

### 示例 3：公网服务器（生产环境，强密钥）
```json
{
  "server_url": "http://deploy.example.com/api",
  "api_key": "Kj8#mP2$vL9@xR5&nQ3"
}
```

### 示例 4：HTTPS（生产环境推荐）
```json
{
  "server_url": "https://deploy.example.com/api",
  "api_key": "SecureKey-2024-ABC123xyz"
}
```

## 验证配置

配置完成后，可以使用 CLI 工具测试连接：

```bash
# 列出所有网站（测试与服务端的连接）
deploy-cli list
```

如果配置正确，会显示网站列表；如果连接失败，会显示错误信息。

## 修改配置

如需修改服务端地址，直接编辑配置文件即可，无需重新编译客户端：

```bash
# Linux/macOS
vim ~/.aideploy/config.json

# Windows
notepad %USERPROFILE%\.aideploy\config.json
```

修改后保存文件，下次运行客户端工具时会自动使用新配置。
