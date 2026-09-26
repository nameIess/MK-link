@echo off
setlocal

go test ./...
if errorlevel 1 exit /b %errorlevel%

go vet ./...
if errorlevel 1 exit /b %errorlevel%

go build -ldflags="-H=windowsgui" -o MK-Link.exe .
if errorlevel 1 exit /b %errorlevel%

echo Built MK-Link.exe
