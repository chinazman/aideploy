#!/bin/bash

echo "编译CLI客户端..."
cd client
go build -o ../bin/deploy-cli main.go
echo "✓ CLI客户端编译完成: bin/deploy-cli"
