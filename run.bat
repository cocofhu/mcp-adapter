@echo off
echo Starting MCP Adapter Server...
echo Installing dependencies...
go mod tidy
echo Starting server on http://localhost:8080
set CGO_ENABLED=0
go run ./backend/main.go
pause