
@echo off
rem run.cmd - Windows equivalent of backend/bin/run (bash)
rem Usage: run up --build -d   or run down --volumes

setlocal

rem Change to the script directory then to ../docker
pushd "%~dp0" >nul 2>&1
cd ..\docker || (
	echo Failed to change to docker directory
	popd >nul 2>&1
	endlocal
	exit /b 1
)

rem If .env missing, copy from .env.example
if not exist ".env" (
	if exist ".env.example" (
		copy /Y ".env.example" ".env" >nul
	)
)

:check_help
if "%~1"=="-h" goto :help
if "%~1"=="--help" goto :help

docker compose -p backend -f docker-compose.service.yml  %*

popd >nul 2>&1
endlocal
exit /b %ERRORLEVEL%

:help
echo Usage: run [compose-subcommand] [args...]
echo.
echo Examples:
echo   run up -d              ^> detached
echo   run up --build -d      ^> build then detached
echo   run down --volumes     ^> stop and remove volumes
popd >nul 2>&1
endlocal
exit /b 0
