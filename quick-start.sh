#!/bin/bash

echo "========================================"
echo "  AI原型快速部署工具 - 快速启动"
echo "========================================"
echo ""

# 创建bin目录
mkdir -p bin

# 检查config.json是否存在
if [ ! -f config.json ]; then
    echo "[1/3] 创建默认配置文件..."
    cd server
    go run main.go -init
    cd ..
    echo ""
    echo "✓ 配置文件已创建: config.json"
    echo "  请编辑配置文件后重新运行此脚本"
    echo ""
    exit 0
fi

echo "[1/3] 编译服务端..."
cd server
go build -o ../bin/deploy-server main.go
if [ $? -ne 0 ]; then
    echo "编译失败！请确保已安装Go环境"
    exit 1
fi
cd ..
echo "✓ 服务端编译完成"
echo ""

echo "[2/3] 编译CLI客户端..."
cd client
go build -o ../bin/deploy-cli main.go
if [ $? -ne 0 ]; then
    echo "编译失败！请确保已安装Go环境"
    exit 1
fi
cd ..
echo "✓ CLI客户端编译完成"
echo ""

echo "[3/3] 启动服务..."
echo ""
echo "========================================"
echo "  服务正在启动..."
echo "  API地址: http://localhost:8080"
echo "  按 Ctrl+C 停止服务"
echo "========================================"
echo ""

./bin/deploy-server
