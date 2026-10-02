@echo off
rsrc -ico ui/icon.ico -o mklink.syso
go build -ldflags="-s -w" -o mklink.exe
echo Build complete.