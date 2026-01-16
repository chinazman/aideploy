# 使用指南

## 快速开始

### 1. 启动服务器

**Windows:**
```bash
quick-start.bat
```

**Linux/macOS:**
```bash
chmod +x quick-start.sh
./quick-start.sh
```

首次运行会创建 `config.json` 配置文件，请根据实际情况编辑：

```json
{
  "base_domain": "yourdomain.com",
  "web_root": "./websites",
  "mode": "subdomain",
  "single_domain": "",
  "port": 8080,
  "enable_versioning": true
}
```

配置完成后再次运行启动脚本。

### 2. 使用命令行工具

#### 创建网站
```bash
bin\deploy-cli create my-prototype
```

输出：
```
✓ 网站创建成功!
  域名: my-prototype.yourdomain.com
  路径: F:\work\aideploy\websites\my-prototype
```

#### 部署网站
```bash
bin\deploy-cli deploy my-prototype prototype.html
```

带版本说明：
```bash
bin\deploy-cli deploy my-prototype prototype.html "更新首页设计"
```

#### 查看所有网站
```bash
bin\deploy-cli list
```

输出：
```
网站列表:
--------------------------------------------------
1. my-prototype
2. another-site
--------------------------------------------------
```

#### 查看版本历史
```bash
bin\deploy-cli versions my-prototype
```

输出：
```
网站 'my-prototype' 的版本历史:
--------------------------------------------------------------------------------
1. abc123def456
   提交: 更新首页设计
   作者: Deployer
   日期: 2025-01-16 19:30:00 +0800

2. def456ghi789
   提交: 初始部署
   作者: Deployer
   日期: 2025-01-16 18:00:00 +0800
--------------------------------------------------------------------------------
```

#### 回滚到指定版本
```bash
bin\deploy-cli rollback my-prototype abc123def456
```

带说明：
```bash
bin\deploy-cli rollback my-prototype abc123def456 "回滚到上一个版本"
```

#### 删除网站
```bash
bin\deploy-cli delete my-prototype
```

会提示确认，输入 `y` 确认删除。

## 常见使用场景

### 场景1：快速发布AI生成的原型

1. AI工具生成HTML文件
2. 创建网站：`deploy-cli create prototype-v1`
3. 上传文件：`deploy-cli deploy prototype-v1 index.html`
4. 分享域名给客户查看

### 场景2：迭代修改

1. 修改后的HTML文件
2. 部署新版本：`deploy-cli deploy prototype-v1 index.html "修复导航栏问题"`
3. 如果有问题，可以随时回滚：
   ```bash
   deploy-cli versions prototype-v1  # 查看历史
   deploy-cli rollback prototype-v1 <hash>  # 回滚
   ```

### 场景3：管理多个原型

```bash
# 创建多个网站
deploy-cli create prototype-a
deploy-cli create prototype-b
deploy-cli create prototype-c

# 分别部署
deploy-cli deploy prototype-a design-a.html
deploy-cli deploy prototype-b design-b.html
deploy-cli deploy prototype-c design-c.html

# 查看所有
deploy-cli list
```

## 配置Nginx

### 子域名模式配置

```nginx
server {
    listen 80;
    server_name *.yourdomain.com;

    # 提取子域名作为网站目录
    set $subdomain "default";
    if ($host ~* "^([a-z0-9-]+)\.yourdomain\.com$") {
        set $subdomain $1;
    }

    # 设置根目录
    root /path/to/aideploy/websites/$subdomain;

    # 主页面
    location / {
        try_files $uri $uri/ /index.html;
        index index.html;
    }

    # 静态资源缓存
    location ~* \.(jpg|jpeg|png|gif|ico|css|js)$ {
        expires 1y;
        add_header Cache-Control "public, immutable";
    }
}
```

### 路径模式配置

```nginx
server {
    listen 80;
    server_name yourdomain.com;

    # 设置根目录为所有网站的父目录
    root /path/to/aideploy/websites;

    # 网站路由
    location / {
        # 尝试直接访问文件
        try_files $uri $uri/ /$uri/index.html;
    }

    # 防止访问隐藏文件
    location ~ /\. {
        deny all;
    }
}
```

配置完成后重启Nginx：
```bash
sudo nginx -t          # 测试配置
sudo systemctl restart nginx  # 重启服务
```

## 故障排查

### 1. 端口被占用
```
错误: bind: address already in use
```

解决方法：
- 修改 `config.json` 中的 `port` 为其他端口
- 或停止占用8080端口的程序

### 2. Git未安装
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
2. Nginx配置是否正确
3. DNS是否解析到正确的IP
4. 防火墙是否开放端口

## 高级用法

### 自定义部署消息

```bash
deploy-cli deploy my-site index.html "修复了登录bug，优化了加载速度"
```

### 批量部署脚本

```bash
#!/bin/bash
# deploy-all.sh

sites=("site1" "site2" "site3")

for site in "${sites[@]}"; do
    echo "部署 $site..."
    deploy-cli deploy $site $site/index.html
done
```

### 定期备份

```bash
#!/bin/bash
# backup.sh

DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_DIR="/path/to/backups"

tar -czf $BACKUP_DIR/websites_$DATE.tar.gz websites/
echo "备份完成: websites_$DATE.tar.gz"
```

添加到crontab定期执行：
```bash
# 每天凌晨2点备份
0 2 * * * /path/to/backup.sh
```
