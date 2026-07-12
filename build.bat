@echo off
REM ============================================================
REM PlaylistAggregator Lite 构建脚本（Windows）
REM 流程：构建前端 -> 编译后端 -> UPX 最小压缩
REM 目的：固化"编译后必须 UPX 压缩"的步骤，避免发布未压缩的大体积 exe
REM 依赖：upx（UPX 4.2.4，需在 PATH 中，例如 C:\Users\Administrator\.local\bin\upx.exe）
REM 注意：运行前请先停止 8081 服务（go build 会覆盖正在运行的 gomusic.exe）
REM ============================================================
setlocal
set ROOT=%~dp0
cd /d "%ROOT%"

REM 1. 构建前端（输出到 static/dist，由 main.go 的 //go:embed static/dist 嵌入）
echo [1/3] Building frontend...
cd static
call npm install
if errorlevel 1 ( echo npm install failed & exit /b 1 )
call npm run build
if errorlevel 1 ( echo npm run build failed & exit /b 1 )
cd ..

REM 2. 编译后端（输出 gomusic.exe）
echo [2/3] Building backend...
go build -o gomusic.exe
if errorlevel 1 ( echo go build failed & exit /b 1 )

REM 3. UPX 最小压缩（--best --lzma）
echo [3/3] Compressing with UPX (--best --lzma)...
where upx >nul 2>&1
if errorlevel 1 (
  echo [WARN] upx not found in PATH, skip compression.
  echo         Install UPX 4.2.4 and ensure upx.exe is on PATH (e.g. C:\Users\Administrator\.local\bin).
  goto :done
)
upx --best --lzma gomusic.exe

:done
echo Build complete: gomusic.exe
endlocal
