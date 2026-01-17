@echo off
echo 编译服务端...
if not exist bin mkdir bin
go build -ldflags="-s -w" -o bin\deploy-server.exe main.go
echo ✓ 服务端编译完成: bin\deploy-server.exe
echo.
echo 启动服务端:
echo   .\bin\deploy-server.exe
echo   或
echo   .\bin\deploy-server.exe -config config.json
