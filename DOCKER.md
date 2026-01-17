# Docker 部署说明

## 快速启动

### 使用 Docker Compose (推荐)

1. 确保已安装 Docker 和 Docker Compose
2. 在项目根目录运行：

```bash
docker-compose up -d
```

### 使用 Docker 命令

1. 构建镜像：

```bash
docker build -t deploy-server:latest .
```

2. 运行容器：

```bash
docker run -d \
  --name deploy-server \
  -p 80:80 \
  -v $(pwd)/bin/config.json:/app/config/config.json:ro \
  -v $(pwd)/bin/websites:/app/websites:rw \
  --restart unless-stopped \
  deploy-server:latest
```

## 配置说明

### 端口映射

- 容器内部端口：80
- 主机端口：80 (可通过 docker-compose.yml 修改)

### 数据卷

- **配置文件**: `./bin/config.json` -> `/app/config/config.json` (只读)
- **网站目录**: `./bin/websites` -> `/app/websites` (读写)

### 环境变量

- `CONFIG_PATH`: 配置文件路径 (默认: `/app/config/config.json`)

## 管理命令

### 查看日志

```bash
docker-compose logs -f
# 或
docker logs -f deploy-server
```

### 停止服务

```bash
docker-compose down
# 或
docker stop deploy-server
```

### 重启服务

```bash
docker-compose restart
# 或
docker restart deploy-server
```

### 进入容器

```bash
docker exec -it deploy-server sh
```

## 健康检查

容器内置健康检查，每 30 秒检查一次服务状态：

```bash
docker inspect --format='{{.State.Health.Status}}' deploy-server
```

## 故障排查

### 查看容器状态

```bash
docker ps -a
```

### 查看详细日志

```bash
docker logs deploy-server
```

### 重新构建镜像

```bash
docker-compose build --no-cache
```

## 生产环境建议

1. **修改端口**: 如果 80 端口被占用，修改 docker-compose.yml 中的端口映射
2. **资源限制**: 添加 CPU 和内存限制
3. **日志管理**: 配置日志驱动和日志轮转
4. **备份**: 定期备份 `websites` 目录和 `.git` 目录
5. **HTTPS**: 使用反向代理 (如 Nginx) 提供 HTTPS 支持

## 配置反向代理 (Nginx)

示例 Nginx 配置：

```nginx
server {
    listen 443 ssl http2;
    server_name *.localhost;

    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;

    location / {
        proxy_pass http://localhost:80;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```
