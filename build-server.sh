#!/bin/bash

echo "编译服务端..."
mkdir -p bin
go build -o bin/deploy-server -ldflags="-s -w" .
echo "✓ 服务端编译完成: bin/deploy-server"
echo ""
echo "启动服务端:"
echo "  ./bin/deploy-server"
echo "  或"
echo "  ./bin/deploy-server -config config.json"
