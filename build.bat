@echo off
cd /d "%~dp0"
go mod tidy
if errorlevel 1 exit /b 1
go build -o chatssh.exe ./cmd/chatssh
if errorlevel 1 exit /b 1
go test ./...
exit /b %errorlevel%
