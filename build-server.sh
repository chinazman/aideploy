#!/bin/bash

echo "编译服务端..."
mkdir -p bin
go build -o bin/deploy-server main.go
echo "✓ 服务端编译完成: bin/deploy-server"
