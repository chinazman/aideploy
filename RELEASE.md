# GitHub Actions 自动发布指南

## 📋 工作流程说明

本项目包含三个 GitHub Actions 工作流：

### 1. Docker CI (.github/workflows/docker-ci.yml)
**触发条件**: 推送到 master/main/develop 分支或 Pull Request

**功能**:
- ✅ Go 代码编译测试
- ✅ Docker 镜像构建测试
- ✅ 容器功能测试
- ❌ **不推送镜像**（仅测试）

### 2. Docker Release (.github/workflows/docker-release.yml)
**触发条件**: 推送版本标签（如 `v1.0.0`）

**功能**:
- 🐳 构建多架构 Docker 镜像 (linux/amd64, linux/arm64)
- 📤 推送到 GitHub Container Registry (ghcr.io)
- 🏷️ 自动添加多个标签

### 3. Create Release (.github/workflows/release.yml)
**触发条件**: 推送版本标签（如 `v1.0.0`）

**功能**:
- 📦 构建多平台二进制文件 (Linux, Windows, macOS)
- 🐳 构建并推送 Docker 镜像
- 📝 自动生成 Release Notes
- 📤 上传构建产物到 GitHub Release

## 🚀 发布新版本

### 方式一：使用 Git 标签（推荐）

1. **更新版本号和变更日志**
   ```bash
   # 编辑 CHANGELOG.md 或在 commit message 中使用规范格式
   git commit -m "feat: 添加新功能"
   git commit -m "fix: 修复bug"
   ```

2. **创建并推送标签**
   ```bash
   # 创建带注释的标签
   git tag -a v1.0.0 -m "Release v1.0.0"

   # 推送标签到 GitHub
   git push origin v1.0.0
   ```

3. **自动构建**
   - 推送标签后，GitHub Actions 自动触发
   - 在 Actions 页面查看构建进度

4. **下载构建产物**
   - 访问项目的 Releases 页面
   - 下载对应平台的二进制文件或 Docker 镜像

### 方式二：通过 GitHub Web 界面

1. 访问项目的 GitHub 页面
2. 点击 "Releases" → "Draft a new release"
3. 选择标签或创建新标签（格式：`v*.*.*`）
4. 填写 Release 标题和描述
5. 点击 "Publish release"

## 📦 可用的构建产物

### Docker 镜像

```bash
# 拉取特定版本
docker pull 你的dockerhub用户名/aideploy:v1.0.0

# 拉取最新版本
docker pull 你的dockerhub用户名/aideploy:latest

# 拉取 ARM64 版本
docker pull 你的dockerhub用户名/aideploy:latest-arm64
```

### 二进制文件

在 Release 页面下载：
- `deploy-server_Linux_x86_64.tar.gz` - Linux AMD64
- `deploy-server_Linux_arm64.tar.gz` - Linux ARM64
- `deploy-server_Windows_x86_64.zip` - Windows 64位
- `deploy-server_Darwin_x86_64.tar.gz` - macOS Intel
- `deploy-server_Darwin_arm64.tar.gz` - macOS Apple Silicon

## 🔐 权限配置

### GitHub Repository 设置

1. **Actions 权限**:
   - Settings → Actions → General
   - ✅ Allow all actions and reusable workflows

2. **Workflow 权限**:
   - Settings → Actions → General → Workflow permissions
   - ✅ Read and write permissions

### Docker Hub 配置

1. **创建 Access Token**:
   - 访问 https://hub.docker.com/settings/security
   - 点击 "New Access Token"
   - 创建一个新 token，命名为 "GitHub Actions"
   - 复制生成的 token（只显示一次）

2. **配置 GitHub Secrets**:
   - 进入 GitHub 仓库
   - Settings → Secrets and variables → Actions
   - 添加以下 secrets:
     - `DOCKERHUB_USERNAME`: 你的 Docker Hub 用户名
     - `DOCKERHUB_TOKEN`: 刚才创建的 Access Token

3. **更新配置文件**:
   - 在 `.github/workflows/docker-release.yml` 中更新 `DOCKERHUB_USERNAME`
   - 在 `.goreleaser.yml` 中设置 `DOCKERHUB_USERNAME` 环境变量

## 📝 Commit 规范

为了生成更好的 Release Notes，建议使用 [Conventional Commits](https://www.conventionalcommits.org/) 格式：

- `feat:` - 新功能
- `fix:` - Bug 修复
- `perf:` - 性能优化
- `docs:` - 文档更新
- `test:` - 测试相关
- `build:` - 构建系统
- `ci:` - CI 配置
- `chore:` - 其他更改

示例：
```bash
git commit -m "feat: 添加用户认证功能"
git commit -m "fix: 修复子域名解析问题"
git commit -m "perf: 优化静态文件缓存"
```

## 🔍 本地测试 Release

在正式发布前，可以使用 GoReleaser 测试：

```bash
# 安装 GoReleaser
go install github.com/goreleaser/goreleaser@latest

# 测试构建（不发布）
goreleaser release --snapshot --clean

# 检查生成的文件
ls dist/
```

## 📊 监控构建

访问项目的 Actions 页面：
```
https://github.com/你的用户名/aideploy/actions
```

查看：
- ✅ 成功的工作流
- ❌ 失败的工作流及日志
- 🔄 正在运行的工作流

## 🛠️ 故障排查

### 构建失败

1. **查看日志**: Actions → 点击失败的工作流 → 查看详细日志
2. **常见问题**:
   - Go 模块下载失败：检查 `go.mod` 文件
   - Docker 构建失败：检查 Dockerfile 语法
   - 测试失败：本地运行 `go test` 确认

### 权限错误

确保 GitHub Token 有足够权限：
1. Settings → Secrets → Actions
2. 检查 `GITHUB_TOKEN` 权限设置

## 📚 参考资料

- [GitHub Actions 文档](https://docs.github.com/en/actions)
- [GoReleaser 文档](https://goreleaser.com/)
- [Docker Build Push Action](https://github.com/docker/build-push-action)
- [Conventional Commits](https://www.conventionalcommits.org/)
