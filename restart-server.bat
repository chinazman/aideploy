@echo off
echo 重启部署服务器...

REM 停止现有进程
taskkill /F /IM deploy-server.exe 2>nul
timeout /t 1 /nobreak >nul

REM 启动服务器
echo 启动服务器...
cd bin
start deploy-server.exe

echo 服务器已重启
echo 配置: path模式
echo 访问: http://localhost:8080/web1/
pause
