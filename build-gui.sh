#!/bin/bash

# 确保 PATH 包含 Go bin 目录
export PATH=$PATH:~/go/bin:$GOPATH/bin

echo "========================================"
echo "编译 Wails GUI 客户端"
echo "========================================"
echo ""

cd client

echo "[1/3] 检查前端依赖..."
if [ ! -d "frontend/node_modules" ]; then
    echo "正在安装前端依赖..."
    cd frontend
    npm install
    cd ..
fi

echo ""
echo "[2/3] 编译 GUI 应用..."
wails build

echo ""
echo "[3/3] 完成!"
echo ""
if [ -f "build/bin/AI原型部署工具" ]; then
    echo "✓ GUI 应用编译成功: build/bin/AI原型部署工具"
else
    echo "✓ GUI 应用编译完成"
fi

cd ..
