#!/bin/bash

echo "重启部署服务器..."

# 停止现有进程
pkill -f deploy-server 2>/dev/null
sleep 1

# 启动服务器
echo "启动服务器..."
cd bin
./deploy-server.exe &

echo "服务器已重启"
echo "配置: path模式"
echo "访问: http://localhost:8080/web1/"
