@echo off
echo 编译CLI客户端...
if not exist bin mkdir bin
cd client
go build -tags cli -ldflags="-s -w" -o ..\bin\deploy-cli.exe main.go deployer.go
echo ✓ CLI客户端编译完成: bin\deploy-cli.exe
cd ..
