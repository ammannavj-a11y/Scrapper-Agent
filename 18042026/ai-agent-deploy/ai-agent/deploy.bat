@echo off
:: ╔══════════════════════════════════════════════════════╗
:: ║     AI AGENT — QUICK DEPLOY SCRIPT (Windows)         ║
:: ╚══════════════════════════════════════════════════════╝

title AI Agent Deployer

echo.
echo  ╔═══════════════════════════════════════════╗
echo  ║           AI AGENT DEPLOYER               ║
echo  ╚═══════════════════════════════════════════╝
echo.

:: Check Node.js
where node >nul 2>&1
if %errorlevel% neq 0 (
    echo  ERROR: Node.js not found.
    echo  Install from: https://nodejs.org  ^(v16 or higher^)
    pause
    exit /b 1
)
echo  OK  Node.js found.

:: Create .env if missing
if not exist .env (
    if exist .env.example (
        copy .env.example .env >nul
        echo  NOTE: Created .env from .env.example
        echo  Please edit .env and add your ANTHROPIC_API_KEY
        echo.
        echo  Get your key at: https://console.anthropic.com/settings/keys
        echo.
        set /p key="  Enter your API key now (or press Enter to skip): "
        if not "%key%"=="" (
            powershell -Command "(Get-Content .env) -replace 'your_api_key_here', '%key%' | Set-Content .env"
            echo  OK  API key saved to .env
        )
    )
)

:: Install dependencies
echo.
echo  Installing dependencies...
call npm install --production
if %errorlevel% neq 0 (
    echo  ERROR: npm install failed
    pause
    exit /b 1
)
echo  OK  Dependencies installed

:: Open browser and start server
echo.
echo  Starting AI Agent...
echo  Opening browser at: http://localhost:3000
echo  Press Ctrl+C to stop
echo.

:: Open browser after 2s delay
start /b cmd /c "timeout /t 2 /nobreak >nul && start http://localhost:3000"

node server.js
pause
