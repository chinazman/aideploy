#!/bin/bash
echo "编译CLI客户端..."
mkdir -p bin
cd client
go build -tags cli -ldflags="-s -w" -o ../bin/deploy-cli main.go deployer.go
echo "✓ CLI客户端编译完成: bin/deploy-cli"
cd ..
