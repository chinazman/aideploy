# 多阶段构建 - 构建阶段
FROM golang:1.21-alpine AS builder

# 设置工作目录
WORKDIR /build

# 复制 go.mod 和 go.sum (如果存在)
COPY go.mod ./
COPY go.sum* ./

# 下载依赖
RUN go mod download || true

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s' \
    -o deploy-server \
    main.go

# 运行阶段 - 使用最小的基础镜像
FROM alpine:latest

# 安装必要的运行时工具
RUN apk --no-cache add ca-certificates git curl

# 创建非 root 用户
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

# 配置 git (在切换用户前以 root 配置)
RUN git config --system user.name "Deploy Server" && \
    git config --system user.email "deployer@localhost"

# 设置工作目录
WORKDIR /app

# 从构建阶段复制可执行文件
COPY --from=builder /build/deploy-server .

# 创建配置文件目录和默认配置
RUN mkdir -p /app/config && \
    mkdir -p /app/websites && \
    echo '{"base_domain":"localhost","web_root":"/app/websites","mode":"subdomain","single_domain":"localhost","port":80,"enable_versioning":true,"api_key":""}' > /app/config/config.json && \
    chown -R appuser:appuser /app

# 切换到非 root 用户
USER appuser

# 暴露端口
EXPOSE 80

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
    CMD curl -f http://localhost/api/sites/list || exit 1

# 启动应用
CMD ["./deploy-server", "-config", "config/config.json"]
