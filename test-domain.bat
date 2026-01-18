@echo off
echo ==========================================
echo 域名访问测试脚本
echo ==========================================
echo.

set /p DOMAIN="请输入要测试的域名（例如: jydemo.localhost 或 jydemo.example.com）: "
echo.

echo [测试 1] 域名解析测试
echo -----------------------------------
echo 正在测试域名: %DOMAIN%
nslookup %DOMAIN%
echo.

echo [测试 2] HTTP 请求测试
echo -----------------------------------
echo 正在测试 HTTP 连接...
curl -v http://%DOMAIN%/
echo.

echo [测试 3] 查看 Host 头
echo -----------------------------------
echo 正在检查服务器接收到的 Host 头...
curl -I http://%DOMAIN%/
echo.

echo ==========================================
echo 测试完成！
echo ==========================================
echo.
echo 如果测试失败，请检查:
echo 1. 服务器是否正在运行
echo 2. config.json 中的 base_domain 配置是否正确
echo 3. DNS 是否正确配置（对于真实域名）
echo 4. 防火墙是否允许端口 80 的访问
echo ==========================================
pause
