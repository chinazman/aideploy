@echo off
echo 编译服务端...
if not exist bin mkdir bin
go build -o bin\deploy-server.exe main.go
echo ✓ 服务端编译完成: bin\deploy-server.exe
