@echo off
chcp 65001
cls

set app_name=jvms
set app_ver=3.0.5

rem 获取当前时间
set "hour=%time:~0,2%"
if "%hour:~0,1%" == " " set "hour=0%hour:~1,1%"

set "year=%date:~0,4%"
set "month=%date:~5,2%"
set "day=%date:~8,2%"

set "minute=%time:~3,2%"
set "second=%time:~6,2%"

set "date_text=%year%-%month%-%day%(%hour%:%minute%:%second%)"
set "date_version=%year%%month%%day%_%hour%%minute%%second%

:: 发布正式版本时，去掉版本号中的时间信息
set BuildTime=%date_text%
echo 编译时间：%date_text%

:: 清理残留的编译进程
taskkill /f /im go.exe    >nul 2>nul
taskkill /f /im compile.exe    >nul 2>nul
taskkill /f /im asm.exe        >nul 2>nul
taskkill /f /im link.exe       >nul 2>nul
taskkill /f /im git.exe        >nul 2>nul

:: 编译程序
SET GO111MODULE=on
SET CGO_ENABLED=0
SET GOOS=windows
SET GOARCH=amd64
taskkill /f /im %app_name%.exe  >nul 2>nul
del %app_name%.exe               >nul 2>nul
del %app_name%*.gz               >nul 2>nul
del %app_name%*.zip              >nul 2>nul
attrib -H *.old                  >nul 2>nul
del *.exe.old                    >nul 2>nul

echo =============================================================
echo 1 - 编译 Windows 可执行程序
echo =============================================================

go build -o %app_name%.exe -ldflags "-X main.AppVersion=%app_ver% -X main.BuildTime=%BuildTime%" .
if errorlevel 1 (
    echo 编译失败，请检查错误信息。
    exit /b 1
)

:: 获取版本号信息
:: echo %app_name%.exe version
%app_name%.exe version >nul 2>nul
if errorlevel 1 (
    echo [E] 获取版本失败，请确认是否有 version 参数。
    exit /b 1
)

for /f "delims=" %%a in ('%app_name%.exe version') do @set "app_version=%%a"
echo 当前版本：%app_version%


echo 2 - 运行程序
echo =============================================================
del jdkdlindex.json   >nul 2>nul

rem %app_name% rls -t lzu -a
%app_name% -l=7 version
