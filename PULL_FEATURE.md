# 从服务器覆盖本地功能说明

## 功能概述

"从服务器覆盖本地"功能允许您将服务器上的网站文件下载到本地目录，实现：
- 在不同机器间同步代码
- 恢复本地文件到服务器最新版本
- 获取服务器的完整副本

## 使用场景

### 1. 多机协作
```bash
# 在机器A上部署到服务器
deploy-cli deploy my-project ./dist

# 在机器B上从服务器获取最新文件
deploy-cli pull my-project
```

### 2. 恢复本地文件
```bash
# 本地文件出现问题时，从服务器恢复
deploy-cli pull my-project
```

### 3. 首次同步
```bash
# 新机器首次加入项目，获取完整代码
deploy-cli pull my-project
```

## 工作原理

### CLI方式

```bash
# 基本用法
deploy-cli pull <website-name>

# 示例
deploy-cli pull my-prototype

# 如果已配置目录，可以省略网站名称自动匹配
cd /path/to/bound/directory
deploy-cli pull
```

### GUI方式

在网站列表中，已绑定的网站会显示 **⬇️ 下载** 按钮：
1. 点击"下载"按钮
2. 确认操作提示
3. 系统自动下载并覆盖本地目录

## 功能细节

### 文件处理

- ✅ **保留**：隐藏文件和目录（以`.`开头）
- ❌ **清空**：所有其他文件和目录
- ✅ **下载**：服务器上的所有网站文件
- ❌ **排除**：`.git`目录不会被下载

### 跟踪信息更新

下载完成后，系统会：
1. 扫描下载的文件
2. 更新本地跟踪信息
3. 确保后续增量部署正常工作

### 安全保护

- 操作前需要确认
- 显示将被覆盖的目录路径
- 保留隐藏文件避免丢失配置

## 完整工作流程示例

### 场景：团队协作

#### 开发者A（创建并部署）
```bash
# 1. 创建网站
deploy-cli create team-project

# 2. 配置目录
deploy-cli config set site team-project ./dist

# 3. 首次部署
deploy-cli deploy team-project "初始版本"
```

#### 开发者B（获取并修改）
```bash
# 1. 配置服务器连接
deploy-cli config set server http://server:8080/api

# 2. 从服务器获取文件
deploy-cli pull team-project

# 3. 本地修改后部署
deploy-cli deploy team-project "添加新功能"
```

#### 开发者A（获取更新）
```bash
# 从服务器获取开发者B的修改
deploy-cli pull team-project
```

## 注意事项

### 1. 目录覆盖警告
⚠️ **重要**：此操作会清空本地目录（隐藏文件除外），请确保：
- 已提交本地修改到版本控制系统
- 确认不需要保留本地文件
- 或者先备份本地文件

### 2. 网络要求
- 需要稳定的网络连接
- 大型网站下载可能需要较长时间
- 确保服务器地址配置正确

### 3. 权限要求
- 本地目录需要写入权限
- 服务器API需要访问权限（API密钥）

### 4. 跟踪同步
下载后跟踪信息会自动更新，下次部署将使用增量模式

## 故障排查

### 问题1：下载失败
```bash
# 检查服务器连接
deploy-cli config get

# 测试服务器可用性
curl http://your-server:8080/api/sites/list
```

### 问题2：权限错误
```bash
# 检查目录权限
ls -la ./dist

# 确保有写入权限
chmod u+w ./dist
```

### 问题3：文件冲突
```bash
# 备份本地文件
cp -r ./dist ./dist.backup

# 再执行下载
deploy-cli pull my-project
```

## 与其他功能的配合

### 与版本控制结合
```bash
# 1. 从服务器下载
deploy-cli pull my-project

# 2. 查看版本历史
deploy-cli versions my-project

# 3. 如需回滚，回滚到指定版本
deploy-cli rollback my-project abc1234
```

### 与增量部署配合
```bash
# 1. 从服务器拉取最新版本
deploy-cli pull my-project

# 2. 本地修改文件
# ... 进行编辑 ...

# 3. 增量部署（只上传修改的文件）
deploy-cli deploy my-project "本地修改"
```

## API说明（供开发者参考）

### 服务端API

**请求**：
```
GET /api/sites/export?name={website-name}
Headers:
  X-API-Key: {your-api-key}  // 如果配置了密钥
```

**响应**：
```
Content-Type: application/x-gzip
Content-Disposition: attachment; filename={website-name}.tar.gz
Body: tar.gz格式的网站文件包
```

### 客户端实现

核心方法：`PullFromServer(sitePath string) error`

功能：
1. 检查本地目录是否存在
2. 请求服务器导出API
3. 清空本地目录（保留隐藏文件）
4. 解压下载的文件
5. 更新跟踪信息

## 总结

"从服务器覆盖本地"功能是一个强大的同步工具，特别适合：
- 团队协作场景
- 多机开发环境
- 代码恢复需求

合理使用此功能可以大大提高团队协作效率，确保代码同步的一致性。
