@echo off
setlocal enabledelayedexpansion
cd /d "%~dp0"
set "PATH=%PATH%;C:\Users\dell\go\bin;%GOPATH%\bin;%LOCALAPPDATA%\Android\Sdk\platform-tools;%ANDROID_HOME%\platform-tools"

echo [SyncRide] Auto-detecting PC LAN IPv4 Address...
for /f "tokens=2 delims=:" %%a in ('ipconfig ^| findstr /c:"IPv4 Address"') do (
    set "IP=%%a"
    set "IP=!IP: =!"
    goto :found_ip
)
:found_ip
echo [SyncRide] Your PC LAN IP is: %IP%

echo [SyncRide] Initializing Docker infrastructure...
docker compose up -d --wait postgres kafka
if errorlevel 1 goto :docker_failed

echo [SyncRide] Launching Ride Service...
start "SyncRide Backend (Ride Service)" cmd /k "cd /d ""%~dp0ride-service"" && go run cmd/main.go"

echo [SyncRide] Waiting for Ride Service on :50051...
powershell -NoProfile -Command "$deadline = (Get-Date).AddSeconds(30); while ((Get-Date) -lt $deadline) { if (Test-NetConnection -ComputerName localhost -Port 50051 -InformationLevel Quiet) { exit 0 }; Start-Sleep -Seconds 1 }; exit 1"
if errorlevel 1 goto :service_failed

echo [SyncRide] Executing Real Ride Booking Test...
go run ride-service/cmd/test_client/main.go

echo.
echo [SyncRide] Backend is running on port :50051 (LAN: %IP%:50051).
goto :eof

:docker_failed
echo [SyncRide] Docker infrastructure failed to become healthy. Check: docker compose logs postgres kafka
exit /b 1

:service_failed
echo [SyncRide] Ride Service did not start. Check the "SyncRide Backend (Ride Service)" window.
exit /b 1
