@echo off
chcp 65001 >nul 2>&1
setlocal EnableDelayedExpansion
::
:: WORNG Installer for Windows (CMD)
:: Downloads and installs the WORNG interpreter from GitHub Releases.
::
:: Usage:
::   install.bat                          install latest version
::   install.bat --version 0.1.0          install specific version
::   install.bat --no-modify-path         skip PATH modification
::   install.bat --help                   show this help
::

set "REPO=KashifKhn/worng"
set "BINARY_NAME=worng"
set "DOCS_URL=https://github.com/KashifKhn/worng"

set "REQUESTED_VERSION="
set "NO_MODIFY_PATH=0"
set "INSTALL_DIR="

:: ---------------------------------------------------------------------------
:: Parse arguments
:: ---------------------------------------------------------------------------
:parse_args
if "%~1"=="" goto :args_done
if /i "%~1"=="--help"           goto :show_help
if /i "%~1"=="-h"               goto :show_help
if /i "%~1"=="--version"        goto :arg_version
if /i "%~1"=="-v"               goto :arg_version
if /i "%~1"=="--no-modify-path" (
    set "NO_MODIFY_PATH=1"
    shift
    goto :parse_args
)
echo   [!] Unknown option: %~1
shift
goto :parse_args

:arg_version
if "%~2"=="" (
    echo   [x] --version requires a version argument
    exit /b 1
)
set "REQUESTED_VERSION=%~2"
:: Strip leading 'v' if present
if "!REQUESTED_VERSION:~0,1!"=="v" set "REQUESTED_VERSION=!REQUESTED_VERSION:~1!"
shift
shift
goto :parse_args

:args_done

:: ---------------------------------------------------------------------------
:: Logo
:: ---------------------------------------------------------------------------
:show_logo
echo.
echo   ██╗    ██╗ ██████╗ ██████╗ ███╗   ██╗ ██████╗
echo   ██║    ██║██╔═══██╗██╔══██╗████╗  ██║██╔════╝
echo   ██║ █╗ ██║██║   ██║██████╔╝██╔██╗ ██║██║  ███╗
echo   ██║███╗██║██║   ██║██╔══██╗██║╚██╗██║██║   ██║
echo   ╚███╔███╔╝╚██████╔╝██║  ██║██║ ╚████║╚██████╔╝
echo    ╚══╝╚══╝  ╚═════╝ ╚═╝  ╚═╝╚═╝  ╚═══╝ ╚═════╝
echo.
echo         The esoteric programming language
echo           if it looks right, it's wrong
echo.
goto :main

:show_help
echo WORNG Installer for Windows (CMD)
echo.
echo Usage: install.bat [options]
echo.
echo Options:
echo   -h, --help              Show this help message
echo   -v, --version ^<ver^>     Install a specific version (e.g., 0.1.0)
echo       --no-modify-path    Don't add install directory to PATH
echo.
echo Examples:
echo   install.bat
echo   install.bat --version 0.1.0
echo   install.bat --no-modify-path
exit /b 0

:: ---------------------------------------------------------------------------
:: Main
:: ---------------------------------------------------------------------------
:main

:: Determine arch  (AMD64 / ARM64 / x86)
set "ARCH=amd64"
if /i "%PROCESSOR_ARCHITECTURE%"=="ARM64"   set "ARCH=arm64"
if /i "%PROCESSOR_ARCHITEW6432%"=="AMD64"   set "ARCH=amd64"
if /i "%PROCESSOR_ARCHITEW6432%"=="ARM64"   set "ARCH=arm64"
if /i "%PROCESSOR_ARCHITECTURE%"=="x86" (
    if not defined PROCESSOR_ARCHITEW6432   set "ARCH=386"
)

:: Resolve version
if not "!REQUESTED_VERSION!"=="" (
    set "VERSION=!REQUESTED_VERSION!"
    echo   --^> Installing version: v!VERSION!
) else (
    <nul set /p ="  Fetching latest version..."
    call :get_latest_version VERSION
    :: Clear the line
    echo.
    if "!VERSION!"=="" (
        echo   [x] Could not determine latest version
        echo   --^> Check releases: https://github.com/%REPO%/releases
        exit /b 1
    )
    echo   [+] Latest version: v!VERSION!
)

:: Resolve install dir
if "!INSTALL_DIR!"=="" (
    if defined LOCALAPPDATA (
        set "INSTALL_DIR=!LOCALAPPDATA!\worng\bin"
    ) else (
        set "INSTALL_DIR=%USERPROFILE%\.local\worng\bin"
    )
)

:: Check existing installation
call :get_installed_version EXISTING_VERSION
if not "!EXISTING_VERSION!"=="" (
    if "!EXISTING_VERSION!"=="!VERSION!" (
        echo   --^> Version v!VERSION! is already installed
        exit /b 0
    ) else (
        echo   --^> Upgrading from v!EXISTING_VERSION! to v!VERSION!
    )
)

echo.
echo   --^> Platform: windows/!ARCH!
echo.

:: Build archive name — matches release.yml: worng_<version>_windows_<arch>.zip
set "ARCHIVE_BASE=%BINARY_NAME%_!VERSION!_windows_!ARCH!"
set "ARCHIVE=!ARCHIVE_BASE!.zip"
set "DOWNLOAD_URL=https://github.com/%REPO%/releases/download/v!VERSION!/!ARCHIVE!"

:: Create temp directory
set "TMP_DIR=%TEMP%\worng_install_%RANDOM%"
mkdir "!TMP_DIR!" 2>nul

:: Download
echo   --^> Downloading %BINARY_NAME% v!VERSION!...
powershell -NoProfile -NonInteractive -Command ^
    "try { Invoke-WebRequest -Uri '!DOWNLOAD_URL!' -OutFile '!TMP_DIR!\!ARCHIVE!' -UseBasicParsing -ErrorAction Stop } catch { Write-Error $_.Exception.Message; exit 1 }"
if %ERRORLEVEL% neq 0 (
    echo   [x] Failed to download %BINARY_NAME% v!VERSION!
    echo   --^> Check releases: https://github.com/%REPO%/releases
    rmdir /s /q "!TMP_DIR!" 2>nul
    exit /b 1
)

:: Extract zip
powershell -NoProfile -NonInteractive -Command ^
    "try { Expand-Archive -Path '!TMP_DIR!\!ARCHIVE!' -DestinationPath '!TMP_DIR!' -Force -ErrorAction Stop } catch { Write-Error $_.Exception.Message; exit 1 }"
if %ERRORLEVEL% neq 0 (
    echo   [x] Failed to extract archive
    rmdir /s /q "!TMP_DIR!" 2>nul
    exit /b 1
)

set "EXTRACTED_BIN=!TMP_DIR!\%BINARY_NAME%.exe"
if not exist "!EXTRACTED_BIN!" (
    echo   [x] Binary not found in archive: %BINARY_NAME%.exe
    rmdir /s /q "!TMP_DIR!" 2>nul
    exit /b 1
)

:: Create install dir and copy binary
if not exist "!INSTALL_DIR!" (
    mkdir "!INSTALL_DIR!" 2>nul
    if %ERRORLEVEL% neq 0 (
        echo   [x] Cannot create install directory: !INSTALL_DIR!
        rmdir /s /q "!TMP_DIR!" 2>nul
        exit /b 1
    )
)

set "DEST_BIN=!INSTALL_DIR!\%BINARY_NAME%.exe"
copy /y "!EXTRACTED_BIN!" "!DEST_BIN!" >nul
if %ERRORLEVEL% neq 0 (
    echo   [x] Failed to copy binary to !DEST_BIN!
    rmdir /s /q "!TMP_DIR!" 2>nul
    exit /b 1
)
echo   [+] Installed to !DEST_BIN!

:: Cleanup temp
rmdir /s /q "!TMP_DIR!" 2>nul

:: Modify PATH
if "!NO_MODIFY_PATH!"=="0" (
    call :add_to_path "!INSTALL_DIR!"
)

echo.
echo   --^> Run "worng --help" to get started
echo   --^> Docs: %DOCS_URL%
echo.
exit /b 0

:: ---------------------------------------------------------------------------
:: Subroutine: get_latest_version
::   Sets %~1 to the latest release version string (no leading v), or empty.
:: ---------------------------------------------------------------------------
:get_latest_version
set "%~1="
set "_VER_TMP="
for /f "usebackq tokens=*" %%L in (
    `powershell -NoProfile -NonInteractive -Command ^
        "try { $r=(Invoke-RestMethod 'https://api.github.com/repos/%REPO%/releases/latest' -Headers @{'User-Agent'='worng-installer'} -ErrorAction Stop).tag_name -replace '^v',''; Write-Output $r } catch { Write-Output '' }"` ^
) do set "_VER_TMP=%%L"
set "%~1=!_VER_TMP!"
exit /b 0

:: ---------------------------------------------------------------------------
:: Subroutine: get_installed_version
::   Sets %~1 to the installed version string, or empty.
:: ---------------------------------------------------------------------------
:get_installed_version
set "%~1="
set "_IV_TMP="

:: Try binary in PATH first
where %BINARY_NAME%.exe >nul 2>&1
if %ERRORLEVEL%==0 (
    for /f "usebackq tokens=*" %%V in (
        `%BINARY_NAME%.exe version 2^>nul` ^
    ) do (
        set "_IV_LINE=%%V"
        for /f "tokens=*" %%M in (
            `powershell -NoProfile -NonInteractive -Command ^
                "if ('!_IV_LINE!' -match 'v?(\d+\.\d+\.\d+)') { $Matches[1] } else { '' }"` ^
        ) do set "_IV_TMP=%%M"
    )
)

:: Try binary in install dir
if "!_IV_TMP!"=="" (
    if exist "!INSTALL_DIR!\%BINARY_NAME%.exe" (
        for /f "usebackq tokens=*" %%V in (
            `"!INSTALL_DIR!\%BINARY_NAME%.exe" version 2^>nul` ^
        ) do (
            set "_IV_LINE=%%V"
            for /f "tokens=*" %%M in (
                `powershell -NoProfile -NonInteractive -Command ^
                    "if ('!_IV_LINE!' -match 'v?(\d+\.\d+\.\d+)') { $Matches[1] } else { '' }"` ^
            ) do set "_IV_TMP=%%M"
        )
    )
)

set "%~1=!_IV_TMP!"
exit /b 0

:: ---------------------------------------------------------------------------
:: Subroutine: add_to_path
::   Adds %~1 to the user PATH registry key (permanent) and current session.
:: ---------------------------------------------------------------------------
:add_to_path
set "_NEW_DIR=%~1"

:: Read current user PATH from registry
for /f "usebackq skip=2 tokens=3*" %%A in (
    `reg query "HKCU\Environment" /v PATH 2^>nul` ^
) do set "_CUR_PATH=%%A %%B"

:: Check if already present
echo !_CUR_PATH! | findstr /i /c:"!_NEW_DIR!" >nul 2>&1
if %ERRORLEVEL%==0 (
    exit /b 0
)

:: Append
if "!_CUR_PATH!"=="" (
    set "_NEW_PATH=!_NEW_DIR!"
) else (
    :: Trim trailing space that the double-token trick introduces
    set "_CUR_PATH=!_CUR_PATH: =!"
    set "_NEW_PATH=!_CUR_PATH!;!_NEW_DIR!"
)

reg add "HKCU\Environment" /v PATH /t REG_EXPAND_SZ /d "!_NEW_PATH!" /f >nul
if %ERRORLEVEL%==0 (
    set "PATH=!PATH!;!_NEW_DIR!"
    echo   [+] Added to user PATH: !_NEW_DIR!
    echo   [!] Restart your terminal for PATH changes to take effect
) else (
    echo   [!] Could not update PATH automatically
    echo   [!] Add manually: !_NEW_DIR!
)
exit /b 0
