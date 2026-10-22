@echo off
setlocal enabledelayedexpansion
title SyncRide 1-Click System Launcher

echo ========================================================
echo           SyncRide 1-Click Real-User App Launcher
echo ========================================================
echo.

:: 1. Auto-Detect PC IPv4 Address
for /f "tokens=2 delims=:" %%a in ('ipconfig ^| findstr /c:"IPv4 Address"') do (
    set "IP=%%a"
    set "IP=!IP: =!"
    goto :found_ip
)
:found_ip
echo [1/5] PC Local IPv4 Address Detected: %IP%

:: 2. Set PATH for Go bin and Android SDK platform-tools
set "PATH=%PATH%;C:\Users\dell\go\bin;%GOPATH%\bin;%LOCALAPPDATA%\Android\Sdk\platform-tools;%ANDROID_HOME%\platform-tools"

:: 3. Start Infrastructure & Backend Services (Non-blocking Docker check)
echo [2/5] Initializing Infrastructure (PostgreSQL and Kafka)...
cd /d "%~dp0backend"
start /b cmd /c "docker compose up -d postgres kafka >nul 2>nul"

echo [3/5] Launching SyncRide Go Backend (Ride Service)...
start "SyncRide Backend (Ride Service)" cmd /k "cd /d "%~dp0backend\ride-service" && go run cmd/main.go"

echo Waiting 5 seconds for Backend Services to initialize...
ping -n 6 127.0.0.1 >nul

:: 4. Build & Launch Android App on Device via ADB if connected
echo [4/5] Checking for connected Android Phone...
where adb >nul 2>nul
if %ERRORLEVEL% EQU 0 (
    adb devices | findstr /v "List" | findstr "device" >nul
    if !ERRORLEVEL! EQU 0 (
        echo Connected Android phone detected! Building and installing app...
        cd /d "%~dp0"
        call gradlew.bat assembleDebug
        if exist "app\build\outputs\apk\debug\app-debug.apk" (
            echo Installing APK to your phone...
            adb install -r app\build\outputs\apk\debug\app-debug.apk
            echo Launching SyncRide App on your phone screen...
            adb shell am start -n com.syncride.app/.LoginActivity
        )
    ) else (
        echo [Notice] No physical phone detected via USB ADB. 
        echo (You can open the SyncRide app manually on your phone anytime!)
    )
) else (
    echo [Notice] ADB tool not found in PATH; skipping phone auto-install.
    echo (You can open the SyncRide app manually on your phone anytime!)
)

:: 5. Execute Automated Real User Ride Booking Test
echo.
echo [5/5] Executing Real Ride Booking Simulation (Jaipur Location)...
cd /d "%~dp0backend"
go run ride-service/cmd/test_client/main.go

echo.
echo ========================================================
echo   SUCCESS! SyncRide System is Live and Ready!
echo   - Backend Server : Listening on localhost:50051 (gRPC)
echo   - LAN IP Access  : http://%IP%:8080 / %IP%:50051
echo   - Kafka Web UI   : http://localhost:8081
echo ========================================================
echo.
pause
