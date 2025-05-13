@echo off
setlocal enabledelayedexpansion

set "CONFIG_FILE=%~dp0..\..\..\config\settings.txt"

:manager
cls
if exist "%CONFIG_FILE%" (
    for /f "usebackq tokens=1,* delims==" %%A in ("%CONFIG_FILE%") do (
        if not "%%A"=="" if not "%%A:~0,1"==";" (
            set "key=%%A"
            set "value=%%B"

            if "!value:~-1!"==" " (
                set "value=!value:~0,-1!"
            )

            set "!key!=!value!"
        )
    )

    if not exist "!DIRECTORY_ORIGINAL!\Oblivion - Voices1.bsa" (
        call :check_path "!DIRECTORY_ORIGINAL!" "Oblivion - Voices1.bsa" "Original Oblivion" "eingabe_original"
    )

    if not exist "!DIRECTORY_OBRE!\Dev\ObvData\Data\Oblivion - Voices1.bsa" (
        call :check_path "!DIRECTORY_OBRE!" "Dev\ObvData\Data\Oblivion - Voices1.bsa" "Oblivion Remastered" "eingabe_obre"
    )

    if defined DIRECTORY_BACKUP (
        if not exist "!DIRECTORY_BACKUP!\" (
            call :check_path "!DIRECTORY_BACKUP!" "" "backup directory" "eingabe_backup"
        )
    ) else (
        call :check_path "!DIRECTORY_BACKUP!" "" "backup directory" "eingabe_backup"
    )

    if not defined REMOVE_TEMP_FILES (
        call :prompt_flag "Remove temporary files after successfully creating the mod" "REMOVE_TEMP_FILES"
    )
) else (
    goto eingabe_obre
)

exit /b

:check_path
set "PATH_VALUE=%~1"
set "TEST_FILE=%~2"
set "DESC=%~3"
set "RETRY_FUNC=%~4"

if not "!PATH_VALUE!"=="" (
	if "%TEST_FILE%" == "" (
		if not exist "!PATH_VALUE!" (
			echo ERROR: Could not find %DESC% with the given path! Please enter a new path.

			findstr /v /r "^%VAR_NAME%=.*" "%CONFIG_FILE%" > "%CONFIG_FILE%.tmp"
			move /y "%CONFIG_FILE%.tmp" "%CONFIG_FILE%" >nul

			pause
			goto %RETRY_FUNC%
		)
	) else (
		if not exist "!PATH_VALUE!\%TEST_FILE%" (
			echo ERROR: Could not find %DESC% with the given path! Please enter a new path.

			findstr /v /r "^%VAR_NAME%=.*" "%CONFIG_FILE%" > "%CONFIG_FILE%.tmp"
			move /y "%CONFIG_FILE%.tmp" "%CONFIG_FILE%" >nul

			pause
			goto %RETRY_FUNC%
		)
	)
) else (
    goto %RETRY_FUNC%
)

goto manager

:eingabe_original
cls
call :prompt_path "Please enter the path to Original Oblivion" "Oblivion - Voices1.bsa" "DIRECTORY_ORIGINAL" "Example path: '...\Steam\SteamApps\common\Oblivion\Data'"
goto manager

:eingabe_obre
cls
call :prompt_path "Please enter the path to Oblivion Remastered" "Dev\ObvData\Data\Oblivion - Voices1.bsa" "DIRECTORY_OBRE" "Example path: '...\Steam\SteamApps\common\Oblivion Remastered\OblivionRemastered\Content'"
goto manager

:eingabe_backup
cls
call :prompt_path "Please choose a folder for us to backup files in" "" "DIRECTORY_BACKUP" "This directory is used to save the original version of files we have to modify"
goto manager


:prompt_path
set "DESC=%~1"
set "REQUIRED_FILE=%~2"
set "VAR_NAME=%~3"
set "EXAMPLE=%~4"

echo %DESC%:

if not "%EXAMPLE%" == "" (
    echo Example path: '...\%EXAMPLE%'
)

:prompt_path_loop
set /p "selectedPath="
set selectedPath=!selectedPath:"=!

if exist "!selectedPath!\%REQUIRED_FILE%" (
	if exist "%CONFIG_FILE%" (
		findstr /v /r "^%VAR_NAME%=.*" "%CONFIG_FILE%" > "%CONFIG_FILE%.tmp"
		move /y "%CONFIG_FILE%.tmp" "%CONFIG_FILE%" >nul
	)

    echo %VAR_NAME%=!selectedPath!>> "%CONFIG_FILE%"
    goto manager
) else (
    echo ERROR: Wrong folder selected, please try again.
    goto prompt_path_loop
)

exit /b

:prompt_flag
set "DESC=%~1"
set "VAR_NAME=%~2"

echo %DESC%:

:prompt_flag_loop
set /p "answer=Do you want to enable this settings? [y/n]"
set answer=!answer:"=!

if "!answer!"=="y" (
    if exist "%CONFIG_FILE%" (
		findstr /v /r "^%VAR_NAME%=.*" "%CONFIG_FILE%" > "%CONFIG_FILE%.tmp"
		move /y "%CONFIG_FILE%.tmp" "%CONFIG_FILE%" >nul
	)

    echo %VAR_NAME%=true>> "%CONFIG_FILE%"
    goto manager
) else (
	if "!answer!"=="n" (
        if exist "%CONFIG_FILE%" (
            findstr /v /r "^%VAR_NAME%=.*" "%CONFIG_FILE%" > "%CONFIG_FILE%.tmp"
            move /y "%CONFIG_FILE%.tmp" "%CONFIG_FILE%" >nul
        )

        echo %VAR_NAME%=false>> "%CONFIG_FILE%"
        goto manager
	) else (
		goto prompt_flag_loop
	)
)

exit /b
