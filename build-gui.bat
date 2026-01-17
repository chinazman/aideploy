@echo off
echo ========================================
echo 编译 Wails GUI 客户端
echo ========================================
echo.

REM 检查 wails 是否安装
where wails >nul 2>nul
if %errorlevel% neq 0 (
    echo 错误: 未找到 Wails CLI
    echo.
    echo 请先安装 Wails:
    echo   go install github.com/wailsapp/wails/v2/cmd/wails@latest
    echo.
    echo 然后确保 %%USERPROFILE%%\go\bin 在你的 PATH 中
    pause
    exit /b 1
)

cd client

echo [1/3] 检查前端依赖...
if not exist frontend\node_modules (
    echo 正在安装前端依赖...
    cd frontend
    call npm install
    cd ..
)

echo.
echo [2/3] 编译 GUI 应用...
wails build

echo.
echo [3/3] 完成!
echo.
if exist build\bin\AI原型部署工具.exe (
    echo ✓ GUI 应用编译成功: build\bin\AI原型部署工具.exe
) else (
    echo ✓ GUI 应用编译完成
)

cd ..
