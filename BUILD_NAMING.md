# 编译输出文件命名规范

## 统一命名规范

所有编译输出的可执行文件应遵循以下命名规范：

### 服务端
- **文件名**: `deploy-server.exe` (Windows) / `deploy-server` (Linux/Mac)
- **位置**: `bin/deploy-server.exe`
- **大小**: 约 6-7 MB (使用 `-ldflags="-s -w"` 压缩)
- **编译命令**:
  ```bash
  # Windows (build-server.bat)
  go build -ldflags="-s -w" -o bin/deploy-server.exe main.go

  # Linux/Mac (build-server.sh)
  go build -ldflags="-s -w" -o bin/deploy-server main.go

  # Makefile
  make server
  ```

### CLI客户端
- **文件名**: `deploy-cli.exe` (Windows) / `deploy-cli` (Linux/Mac)
- **位置**: `bin/deploy-cli.exe`
- **大小**: 约 6 MB (使用 `-ldflags="-s -w"` 压缩)
- **编译命令**:
  ```bash
  # Windows (build-cli.bat)
  cd client && go build -tags cli -ldflags="-s -w" -o ../bin/deploy-cli.exe main.go deployer.go

  # Linux/Mac (build-cli.sh)
  cd client && go build -tags cli -ldflags="-s -w" -o ../bin/deploy-cli main.go deployer.go

  # Makefile
  make client
  ```

### GUI客户端
- **文件名**: `AI原型部署工具.exe`
- **位置**: `client/build/bin/AI原型部署工具.exe`
- **编译命令**:
  ```bash
  # Windows (build-gui.bat)
  cd client && wails build

  # Linux/Mac (build-gui.sh)
  cd client && wails build

  # Makefile
  make gui
  ```

## 命名规则说明

### 前缀规范
- `deploy-` 前缀：表示这是部署工具相关的可执行文件
- 保持简洁明了，便于识别和使用

### 文件名对照表

| 组件 | Windows | Linux/Mac | 说明 |
|------|---------|-----------|------|
| 服务端 | `deploy-server.exe` | `deploy-server` | 部署服务器 |
| CLI客户端 | `deploy-cli.exe` | `deploy-cli` | 命令行客户端 |
| GUI客户端 | `AI原型部署工具.exe` | `AI原型部署工具` | 图形界面客户端 |

## 编译脚本规范

### 确保使用统一命名的脚本

1. **build-server.bat** ✅
   ```bat
   go build -ldflags="-s -w" -o bin\deploy-server.exe main.go
   ```

2. **build-server.sh** ✅
   ```bash
   go build -ldflags="-s -w" -o bin/deploy-server main.go
   ```

3. **Makefile** ✅
   ```makefile
   server:
       @go build -ldflags="-s -w" -o bin/deploy-server main.go
   ```

4. **quick-start.bat** ✅
   ```bat
   go build -o bin/deploy-server.exe main.go
   ```

5. **restart-server.bat** ✅
   ```bat
   taskkill /F /IM deploy-server.exe 2>nul
   start deploy-server.exe
   ```

6. **restart-server.sh** ✅
   ```bash
   ./deploy-server.exe &
   ```

7. **test.bat** ✅
   ```bat
   bin\deploy-server.exe -init
   if exist bin\deploy-server.exe (
       echo ✓ 服务端: bin\deploy-server.exe
   )
   ```

## 旧的命名（已废弃）

以下命名已不再使用，仅作参考：

| 旧名称 | 新名称 | 说明 |
|--------|--------|------|
| `server.exe` | `deploy-server.exe` | 服务端 |
| `aideploy-cli.exe` | `deploy-cli.exe` | CLI客户端 |

## 注意事项

1. **编译优化参数**：使用 `-ldflags="-s -w"` 减小文件体积
   ```bash
   # -s: 去除符号表
   # -w: 去除调试信息
   # 可以将文件大小从 ~15MB 减少到 ~6MB
   ```

2. **手动编译时**：请确保使用 `-o` 参数指定正确的输出文件名
   ```bash
   # 正确 ✅
   go build -ldflags="-s -w" -o bin/deploy-server.exe main.go

   # 错误 ❌
   go build -o bin/server.exe
   go build  # 会使用默认名称，且可能不兼容
   ```

3. **编译位置**：服务端必须从根目录编译（因为main.go在根目录）
   ```bash
   # 正确 ✅
   go build -ldflags="-s -w" -o bin/deploy-server.exe main.go

   # 错误 ❌
   cd server
   go build -o ../bin/deploy-server.exe main.go  # server目录下没有main.go
   ```

4. **清理旧文件**：如果bin目录中有旧名称的文件，建议删除
   ```bash
   # Windows
   del bin\server.exe
   del bin\aideploy-cli.exe
   del bin\deploy-server-new.exe

   # Linux/Mac
   rm bin/server
   rm bin/aideploy-cli
   rm bin/deploy-server-new
   ```

5. **文档一致性**：所有文档和脚本中应统一使用新名称
   - README.md
   - 脚本文件
   - 配置文件
   - 帮助文档

## 常见问题

### 问题1：编译的程序无法运行

**症状**：提示"该版本的程序与你的Windows版本不兼容"

**原因**：编译参数不正确或编译位置错误

**解决**：
```bash
# 确保使用正确的编译命令
go build -ldflags="-s -w" -o bin/deploy-server.exe main.go

# 检查编译目标
go env GOOS GOARCH
# 应该显示: windows amd64
```

### 问题2：文件体积过大

**症状**：编译的程序超过15MB

**解决**：使用优化参数
```bash
go build -ldflags="-s -w" -o bin/deploy-server.exe main.go
```

### 问题3：编译找不到main.go

**症状**：`no required module provides package main.go`

**原因**：在错误的目录编译

**解决**：
```bash
# 服务端必须从根目录编译
cd /path/to/aideploy
go build -ldflags="-s -w" -o bin/deploy-server.exe main.go
```

## 更新历史

- 2025-01-17: 统一服务端命名为 `deploy-server.exe`，CLI客户端命名为 `deploy-cli.exe`
- 所有构建脚本已更新为使用新名称
- Makefile 已更新为使用新名称
