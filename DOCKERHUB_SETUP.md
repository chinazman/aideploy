# Docker Hub 配置指南

本文档说明如何配置 Docker Hub 以实现自动发布。

## 📋 前置要求

1. 已注册 Docker Hub 账号：https://hub.docker.com/
2. 有 GitHub 仓库的 admin 权限

## 🔧 配置步骤

### 1. 创建 Docker Hub Access Token

1. **登录 Docker Hub**
   - 访问：https://hub.docker.com/
   - 使用你的账号登录

2. **创建 Access Token**
   - 点击右上角头像 → **Account Settings**
   - 选择左侧菜单 **Security**
   - 点击 **"New Access Token"** 按钮
   - 填写信息：
     - **Access Token Description**: `GitHub Actions - aideploy`
     - **Access permissions**: 选择 **Read & Write**（需要推送权限）
   - 点击 **Generate** 按钮
   - **重要**：立即复制生成的 token（格式如：`dckr_pat_XXXXX`）
   - 这个 token 只会显示一次！

### 2. 配置 GitHub Secrets

1. **进入仓库设置**
   - 访问你的 GitHub 仓库
   - 点击 **Settings** 标签页

2. **添加 Secrets**
   - 左侧菜单选择 **Secrets and variables** → **Actions**
   - 点击 **"New repository secret"** 按钮
   - 添加两个 secrets：

   **Secret 1: DOCKERHUB_USERNAME**
   - Name: `DOCKERHUB_USERNAME`
   - Value: 你的 Docker Hub 用户名
   - 点击 **Add secret**

   **Secret 2: DOCKERHUB_TOKEN**
   - Name: `DOCKERHUB_TOKEN`
   - Value: 刚才复制的 Access Token
   - 点击 **Add secret**

### 3. 更新配置文件

#### 3.1 更新 `.github/workflows/docker-release.yml`

找到第 10 行，修改 Docker Hub 用户名：

```yaml
env:
  DOCKERHUB_USERNAME: your-dockerhub-username  # 改成你的用户名
```

例如：
```yaml
env:
  DOCKERHUB_USERNAME: johndoe  # 你的实际用户名
```

#### 3.2 更新 `.goreleaser.yml`

在 `.goreleaser.yml` 中，镜像名称已经配置为使用环境变量：
```yaml
docker.io/{{ .Env.DOCKERHUB_USERNAME }}/aideploy:{{ .Version }}
```

这会自动读取 GitHub Actions 中设置的环境变量。

### 4. 验证配置

创建一个测试标签来验证配置是否正确：

```bash
# 创建测试标签
git tag v1.0.0-test -m "Test release"

# 推送标签
git push origin v1.0.0-test
```

然后在 GitHub Actions 页面查看构建是否成功。

## 🎯 Docker Hub 镜像命名规则

配置完成后，发布的镜像格式为：

```
docker.io/你的用户名/aideploy:版本号
docker.io/你的用户名/aideploy:latest
docker.io/你的用户名/aideploy:版本号-arm64
docker.io/你的用户名/aideploy:latest-arm64
```

示例：
```bash
docker pull johndoe/aideploy:v1.0.0
docker pull johndoe/aideploy:latest
docker pull johndoe/aideploy:latest-arm64
```

## 🔒 安全建议

1. **定期轮换 Token**:
   - 建议每 6 个月更新一次 Access Token
   - 删除不再使用的 token

2. **限制权限**:
   - 只授予必要的权限（Read & Write）
   - 不要使用管理员 token

3. **监控活动**:
   - 定期检查 Docker Hub 的活动日志
   - 确认只有预期的推送操作

4. **保护 Secrets**:
   - 不要在代码中硬编码 token
   - 不要将 token 提交到 git
   - 定期审查 GitHub Secrets

## 🛠️ 故障排查

### 问题 1: 推送失败 - "unauthorized: authentication required"

**原因**: Token 无效或过期

**解决方法**:
1. 检查 `DOCKERHUB_TOKEN` secret 是否正确
2. 重新创建 Access Token
3. 更新 GitHub secret

### 问题 2: 推送失败 - "denied: requested access to the resource is denied"

**原因**: Token 权限不足

**解决方法**:
1. 确认 token 有 **Read & Write** 权限
2. 在 Docker Hub 安全设置中重新生成 token

### 问题 3: 镜像名称错误

**原因**: `DOCKERHUB_USERNAME` 配置不正确

**解决方法**:
1. 检查 `.github/workflows/docker-release.yml` 中的用户名
2. 确保与 Docker Hub 用户名完全一致（区分大小写）

### 问题 4: 构建成功但镜像未推送

**原因**: 标签格式不符合 `v*.*.*` 格式

**解决方法**:
使用正确的标签格式：
```bash
git tag v1.0.0        # ✅ 正确
git tag v1.0.0-beta   # ✅ 正确
git tag 1.0.0         # ❌ 错误（缺少 v 前缀）
git tag version-1.0.0 # ❌ 错误
```

## 📚 参考资料

- [Docker Hub 官方文档](https://docs.docker.com/docker-hub/)
- [Docker Hub Access Tokens](https://docs.docker.com/security/for-developers/access-tokens/)
- [GitHub Actions Docker Login](https://github.com/docker/login-action)
