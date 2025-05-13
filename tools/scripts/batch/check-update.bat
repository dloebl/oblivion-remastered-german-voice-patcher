@echo off
setlocal enabledelayedexpansion

:: Config
set "RELEASE_API_URL=https://api.github.com/repos/dloebl/oblivion-remastered-german-voice-patcher/releases/latest"
set "RELEASE_JSON=%~dp0..\..\..\release.json"
set "RELEASE_ZIP=latest_release.zip"
set "VERSION_FILE=%~dp0..\..\..\config\version.txt"
set "EXTRACT_DIR=%~dp0..\..\..\..\update_temp"
set "LOG_FILE=%~dp0..\..\..\logs\updater.txt"

if not exist "%VERSION_FILE%" (
    echo ERROR: Version file not found at %VERSION_FILE%. >> %LOG_FILE%
    pause
    exit /b 1
)

:: Check if leftover temp folder exists
:: TODO: Change to respect setting
:: if exist "%EXTRACT_DIR%" rd /s /q "%EXTRACT_DIR%"
mkdir "%EXTRACT_DIR%"

if exist "%LOG_FILE%" rd /s /q "%LOG_FILE%"

echo Check for new version...
curl -s -L --max-time 30 "%RELEASE_API_URL%" -o "%RELEASE_JSON%"

if %ERRORLEVEL% NEQ 0 (
    echo ERROR: Could not fetch GitHub release info. >> %LOG_FILE%
    pause
    exit /b 1
)

if not exist "%RELEASE_JSON%" (
    echo ERROR: The release JSON file was not downloaded. >> %LOG_FILE%
    pause
    exit /b 1
)

:: Get remote version
set "REMOTE_VERSION="
for /f "tokens=2 delims=:" %%A in ('findstr "tag_name" "%RELEASE_JSON%"') do (
    set "REMOTE_VERSION=%%A"
    set "REMOTE_VERSION=!REMOTE_VERSION:~2,-2!"
)
if "!REMOTE_VERSION!"=="" (
    echo ERROR: Could not extract remote version. >> %LOG_FILE%
    del "%RELEASE_JSON%"
    pause
    exit /b 1
)

:: Check current version
set "LOCAL_VERSION=unknown"
if exist "%VERSION_FILE%" (
    set /p LOCAL_VERSION=<"%VERSION_FILE%"
)
echo Current version: !LOCAL_VERSION!

:: Compare versions
if "!LOCAL_VERSION!"=="!REMOTE_VERSION!" (
    echo Newest version is already installed
    del "%RELEASE_JSON%"
    exit /b 0
)

goto should_update

exit /b


:update
echo Downloading update...

:: Download zip
for /f "tokens=1 delims=," %%A in ('findstr /i "zipball_url" "%RELEASE_JSON%"') do (
    set "line=%%A"
    set "line=!line:*: =!"
    set "DOWNLOAD_URL=!line:"=!"
)

if "!DOWNLOAD_URL!"=="" (
    echo ERROR: Could not find download url. >> %LOG_FILE%
    del "%RELEASE_JSON%"
    pause
    exit /b 1
)

curl -L --max-time 30 -o "%RELEASE_ZIP%" "!DOWNLOAD_URL!"

if %ERRORLEVEL% NEQ 0 (
	echo test3 %ERRORLEVEL%
    echo ERROR: Could not download update file. >> %LOG_FILE%
    del "%RELEASE_JSON%"
    pause
    exit /b 1
)

if not exist "%RELEASE_ZIP%" (
    echo ERROR: Downloaded file does not exist. >> %LOG_FILE%
    pause
    exit /b 1
)

echo Extract new version...
"%~dp0..\..\..\tools\busybox\busybox.exe" unzip -o "%RELEASE_ZIP%" -d "%EXTRACT_DIR%"

for /d %%D in ("%EXTRACT_DIR%\*") do set "EXTRACTED_FOLDER=%%D"

"%~dp0..\..\busybox\busybox.exe" cp -r "%~dp0..\..\..\logs" "%EXTRACTED_FOLDER%\logs"

if exist "%~dp0..\..\..\config\settings.txt" (
    if not exist "%EXTRACTED_FOLDER%\config" (
        mkdir "%EXTRACTED_FOLDER%\config"
    )

    "%~dp0..\..\busybox\busybox.exe" cp -r "%~dp0..\..\..\config\settings.txt" "%EXTRACTED_FOLDER%\config\settings.txt"
)

echo Updating files...

:: Resolve path before updating files as structure might not exist like that afterwards
pushd "%~dp0..\..\.."
set "TARGET_FOLDER=%cd%"
popd

if exist "%EXTRACTED_FOLDER%\tools\busybox\busybox.exe" (
    set "BUSYBOX_PATH_NEW=%EXTRACTED_FOLDER%\tools\busybox\busybox.exe"
) else (
    :: TODO: This is a fallback for testing
    set "BUSYBOX_PATH_NEW=%EXTRACTED_FOLDER%\busybox\busybox.exe"
)

:: Increment version
echo !REMOTE_VERSION! > "%VERSION_FILE%"

"%BUSYBOX_PATH_NEW%" cp -a "%EXTRACTED_FOLDER%\." "%TARGET_FOLDER%"

echo Test: %TARGET_FOLDER%

if %ERRORLEVEL% NEQ 0 (
    echo ERROR: Something went wrong while copying files.
    exit /b 1
)

for /r "%TARGET_FOLDER%" %%F in (*) do (
    set "TARGET_ITEM=%%F"
    setlocal enabledelayedexpansion
    set "REL_PATH=!TARGET_ITEM:%TARGET_FOLDER%=!"
    if not exist "%EXTRACTED_FOLDER%!REL_PATH!" (
        echo Deleting obsolete item: %%F
        "%BUSYBOX_PATH_NEW%" rm -rf "%%F"
    )
    endlocal
)

pushd "%TARGET_FOLDER%"
echo Test: %TARGET_FOLDER%

:: Clean up
rd /s /q "%TARGET_FOLDER%\..\update_temp"

echo Patcher has been updated to !REMOTE_VERSION!.
echo Patcher will now restart...
popd

:: Restart patcher
:: start "" "%TARGET_FOLDER%\Create-Mod.bat"

exit


:should_update

:ask_update
echo Update available: !REMOTE_VERSION!
set /p shouldUpdate=Do you want to update? [y/n]

echo !shouldUpdate!

if "!shouldUpdate!"=="y" (
    goto update
) else (
	if "!shouldUpdate!"=="n" (
		:: TODO: start script "CreateMod.bat" in the root folder with parameter -skipUpdate true
		exit /b
	) else (
		goto ask_update
	)
)

exit /b