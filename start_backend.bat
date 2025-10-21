@echo off
echo Starting AI Doc System Backend...

REM Set environment variables
set DATABASE_URL=postgres://postgres:password123@localhost:5432/ai_doc_system?sslmode=disable
set DB_HOST=localhost
set DB_PORT=5432
set DB_USER=postgres
set DB_PASSWORD=password123
set DB_NAME=ai_doc_system
set JWT_SECRET=your-super-secret-jwt-key-change-in-production
set PORT=8080
set GIN_MODE=release

REM Change to backend directory
cd /d "%~dp0backend"

REM Check if Go is available
go version >nul 2>&1
if %errorlevel% neq 0 (
    echo Error: Go is not installed or not in PATH
    echo Please install Go from https://golang.org/dl/
    pause
    exit /b 1
)

REM Run the backend
echo Starting backend server on port 8080...
go run cmd/main.go

pause