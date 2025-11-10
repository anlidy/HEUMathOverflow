@echo off
rem run.cmd - Windows wrapper script
rem Usage: run up --build -d   or run down --volumes

rem Try to use Python script if available, otherwise use batch logic
if exist "%~dp0run.py" (
	python "%~dp0run.py" %*
	exit /b %ERRORLEVEL%
)

rem Fallback to batch implementation
setlocal enabledelayedexpansion

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

rem Check for help
if "%~1"=="-h" goto :help
if "%~1"=="--help" goto :help

rem Get compose subcommand (default: up)
if "%~1"=="" (
	set "CMD=up"
	set "REST_ARGS=-d"
) else (
	set "CMD=%~1"
	shift
	rem Build remaining arguments string manually
	set "REST_ARGS="
	:collect_loop
	if not "%~1"=="" (
		if "!REST_ARGS!"=="" (
			set "REST_ARGS=%~1"
		) else (
			set "REST_ARGS=!REST_ARGS! %~1"
		)
		shift
		goto :collect_loop
	)
)

rem Execute docker compose commands
if /I "%CMD%"=="down" (
	if "!REST_ARGS!"=="" (
		docker compose -p backend -f docker-compose.service.yml down
		if errorlevel 1 goto :error
		docker compose -p backend -f docker-compose.yml down
		if errorlevel 1 goto :error
	) else (
		call docker compose -p backend -f docker-compose.service.yml down !REST_ARGS!
		if errorlevel 1 goto :error
		call docker compose -p backend -f docker-compose.yml down !REST_ARGS!
		if errorlevel 1 goto :error
	)
) else (
	if "!REST_ARGS!"=="" (
		docker compose -p backend -f docker-compose.yml up -d
		if errorlevel 1 goto :error
		docker compose -p backend -f docker-compose.service.yml up -d
		if errorlevel 1 goto :error
	) else (
		call docker compose -p backend -f docker-compose.yml up !REST_ARGS!
		if errorlevel 1 goto :error
		call docker compose -p backend -f docker-compose.service.yml up !REST_ARGS!
		if errorlevel 1 goto :error
	)
)

popd >nul 2>&1
endlocal
exit /b 0

:error
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
