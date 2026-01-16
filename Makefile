.PHONY: all clean server client run help

# 默认目标
all: server client

# 编译服务端
server:
	@echo "编译服务端..."
	@cd server && go build -o ../bin/deploy-server main.go
	@echo "✓ 服务端编译完成"

# 编译客户端
client:
	@echo "编译CLI客户端..."
	@cd client && go build -o ../bin/deploy-cli main.go
	@echo "✓ CLI客户端编译完成"

# 编译GUI应用（需要Wails）
gui:
	@echo "编译GUI应用..."
	@cd client && wails build
	@echo "✓ GUI应用编译完成"

# 运行服务端
run: server
	@echo "启动服务..."
	@./bin/deploy-server

# 初始化配置
init:
	@cd server && go run main.go -init

# 清理编译文件
clean:
	@echo "清理编译文件..."
	@rm -rf bin/
	@rm -f server/deploy-server
	@rm -f client/deploy-cli
	@echo "✓ 清理完成"

# 安装依赖
deps:
	@echo "下载Go依赖..."
	@go mod download
	@echo "✓ 依赖下载完成"

# 格式化代码
fmt:
	@echo "格式化代码..."
	@go fmt ./...
	@echo "✓ 代码格式化完成"

# 运行测试
test:
	@echo "运行测试..."
	@go test ./...

# 显示帮助
help:
	@echo "AI原型快速部署工具 - Make命令"
	@echo ""
	@echo "使用方法:"
	@echo "  make          - 编译服务端和CLI客户端"
	@echo "  make server   - 编译服务端"
	@echo "  make client   - 编译CLI客户端"
	@echo "  make gui      - 编译GUI应用（需要Wails）"
	@echo "  make run      - 编译并运行服务端"
	@echo "  make init     - 创建配置文件"
	@echo "  make clean    - 清理编译文件"
	@echo "  make deps     - 下载依赖"
	@echo "  make fmt      - 格式化代码"
	@echo "  make test     - 运行测试"
	@echo "  make help     - 显示此帮助信息"
