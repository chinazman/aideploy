@echo off
echo ========================================
echo   测试 AI原型快速部署工具 v2.0
echo ========================================
echo.

REM 检查bin目录
if not exist bin (
    echo [错误] 请先运行 quick-start.bat 编译项目
    pause
    exit /b 1
)

echo [1/5] 创建配置文件...
if not exist config.json (
    bin\deploy-server.exe -init
    echo.
    echo ✓ 配置文件已创建
    echo.
)

echo [2/5] 编译完成检查...
if exist bin\deploy-server.exe (
    echo ✓ 服务端: bin\deploy-server.exe
)
if exist bin\deploy-cli.exe (
    echo ✓ 客户端: bin\deploy-cli.exe
)
echo.

echo [3/5] 测试帮助信息...
bin\deploy-cli.exe help
echo.

echo [4/5] 当前配置:
if exist config.json (
    type config.json
) else (
    echo [未配置]
)
echo.

echo [5/5] 准备就绪!
echo.
echo ========================================
echo   下一步操作:
echo   1. 编辑 config.json 配置文件
echo   2. 运行 quick-start.bat 启动服务
echo   3. 在另一个终端运行部署命令
echo ========================================
echo.

echo 常用命令:
echo   创建网站:   bin\deploy-cli.exe create my-site
echo   部署网站:   bin\deploy-cli.exe deploy my-site test-site
echo   全量部署:   bin\deploy-cli.exe deploy-full my-site test-site
echo   增量部署:   bin\deploy-cli.exe deploy-inc my-site test-site
echo   查看列表:   bin\deploy-cli.exe list
echo   查看版本:   bin\deploy-cli.exe versions my-site
echo.

pause
