@echo off
chcp 65001
setlocal enabledelayedexpansion

cd /d "%~dp0"

echo Execution directory: "%~dp0"

:: Check for update of patcher
::call "%~dp0tools\scripts\batch\check-update.bat"

call "%~dp0tools\scripts\batch\get-settings.bat"
cls


set "VERSION_FILE=%~dp0config\version.txt"

if not exist "%VERSION_FILE%" (
    call :throw_error "ERROR: Version file not found at %VERSION_FILE%"
)
set /p VERSION=<"%VERSION_FILE%"

echo =======================================================================================
echo ===                                                                                 ===
echo ===                Oblivion Remastered German Voice Patcher !VERSION!                  ===
echo ===                                                                                 ===
echo =======================================================================================
echo === Nexusmods: https://www.nexusmods.com/oblivionremastered/mods/1092               ===
echo === Github:    https://github.com/dloebl/oblivion-remastered-german-voice-patcher   ===
echo === Discord:   https://discord.gg/CTsmFfj5                                          ===
echo =======================================================================================

timeout /t 2 >nul
echo STEP: Initialising patcher...


:: =================================================================
:: Check required files
:: =================================================================

set "CONFIG_FILE=%~dp0config\settings.txt"
set "AMOUNTS_FILE=%~dp0custom\german\amounts.txt"
set "LAST_SUCCESSFUL_STEP_FILE=%~dp0tmp\lastStep.txt"

:: Load settings file
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
) else (
    call :throw_error "ERROR: No settings file found"
)

:: Load amounts file
if exist "%AMOUNTS_FILE%" (
    for /f "usebackq tokens=1,* delims==" %%A in ("%AMOUNTS_FILE%") do (
        if not "%%A"=="" if not "%%A:~0,1"==";" (
            set "key=%%A"
            set "value=%%B"

            if "!value:~-1!"==" " (
                set "value=!value:~0,-1!"
            )

            set "!key!=!value!"
        )
    )
) else (
    call :throw_error "ERROR: No amounts file found"
)


:: =================================================================
:: Set Paths
:: =================================================================

set "VOICES_1=Oblivion - Voices1"
set "VOICES_2=Oblivion - Voices2"
set "SHIVERING_ISLES=DLCShiveringIsles - Voices"
set "KNIGHTS=Knights"

set "DLC_1=DLCHorseArmor"
set "DLC_2=DLCOrrery"
set "DLC_3=DLCThievesDen"
set "DLC_4=DLCVilelair"

set "VOICES_1_BSA_ORIGINAL=!DIRECTORY_ORIGINAL!\%VOICES_1%.bsa"
set "VOICES_2_BSA_ORIGINAL=!DIRECTORY_ORIGINAL!\%VOICES_2%.bsa"
set "SHIVERING_ISLES_BSA_ORIGINAL=!DIRECTORY_ORIGINAL!\%SHIVERING_ISLES%.bsa"
set "KNIGHTS_BSA_ORIGINAL=!DIRECTORY_ORIGINAL!\%KNIGHTS%.bsa"

for %%A in ("%VOICES_1_BSA_ORIGINAL%") do set "size_base=%%~zA"
if !size_base! NEQ !EXPECTED_SIZE_BASE! (
    if "!IGNORE_MISMATCH!" == "true" (
        echo INFO: Original Oblivion has incorrect size. Amount: !size_base! bytes
        echo INFO: Will continue since 'ignore mismatch' setting is enabled
    ) else (
        echo ERROR: Original Oblivion has incorrect size. Amount: !size_base! bytes
        call :throw_error "This probably means that your original oblivion files are in english"
    )
)

echo INFO: Checking for optional DLC in original Oblivion
:: Optional DLC
if exist "!DIRECTORY_ORIGINAL!\%DLC_1%.bsa" (
    for %%A in ("!DIRECTORY_ORIGINAL!\%DLC_1%.bsa") do set "size_dlc=%%~zA"
    if !size_dlc! EQU !EXPECTED_SIZE_DLC! (
        echo INFO: DLC for original Oblivion were found!

        set "DLC_1_BSA_ORIGINAL=!DIRECTORY_ORIGINAL!\%DLC_1%.bsa"
        set "DLC_2_BSA_ORIGINAL=!DIRECTORY_ORIGINAL!\%DLC_2%.bsa"
        set "DLC_3_BSA_ORIGINAL=!DIRECTORY_ORIGINAL!\%DLC_3%.bsa"
        set "DLC_4_BSA_ORIGINAL=!DIRECTORY_ORIGINAL!\%DLC_4%.bsa"

        :: Change amounts to use for later checks to DLC count
        set EXPECTED_AMOUNT_AUDIOS=!EXPECTED_AMOUNT_AUDIOS_WITH_DLC!
        set EXPECTED_AMOUNT_BNKS=!EXPECTED_AMOUNT_BNKS_WITH_DLC!
    ) else (

        if !size_dlc! EQU 4018605 (
            echo INFO: Optional DLC were found but are in english
            echo INFO: This usually means that both the GOTY and GOTY Deluxe editions are installed at the same time.
            echo INFO: This is no problem, optional DLC will be ignored and patcher will proceed.
        ) else (
            echo INFO: Optional DLC has incorrect size. Amount: !size_dlc! bytes. Will be handled as missing optional DLC.
        )
    )
) else (
    echo INFO: DLC for original Oblivion could not be found!
)

if not exist "%VOICES_1_BSA_ORIGINAL%" (
    call :throw_error "ERROR: Could not find .bsa files for original Oblivion"
)

set "VOICES_1_BSA_OBRE=!DIRECTORY_BACKUP!\%VOICES_1%.bsa"
set "VOICES_2_BSA_OBRE=!DIRECTORY_BACKUP!\%VOICES_2%.bsa"
set "SHIVERING_ISLES_BSA_OBRE=!DIRECTORY_BACKUP!\%SHIVERING_ISLES%.bsa"
set "KNIGHTS_BSA_OBRE=!DIRECTORY_BACKUP!\%KNIGHTS%.bsa"
:: Optional DLC
set "DLC_1_BSA_OBRE=!DIRECTORY_BACKUP!\%DLC_1%.bsa"
set "DLC_2_BSA_OBRE=!DIRECTORY_BACKUP!\%DLC_2%.bsa"
set "DLC_3_BSA_OBRE=!DIRECTORY_BACKUP!\%DLC_3%.bsa"
set "DLC_4_BSA_OBRE=!DIRECTORY_BACKUP!\%DLC_4%.bsa"

if not exist "%VOICES_1_BSA_OBRE%" (
    :: Use normal game files instead
    if exist "!DIRECTORY_OBRE!\Dev\ObvData\Data\%VOICES_1%.bsa" (
        set "VOICES_1_BSA_OBRE=!DIRECTORY_OBRE!\Dev\ObvData\Data\%VOICES_1%.bsa"
        set "VOICES_2_BSA_OBRE=!DIRECTORY_OBRE!\Dev\ObvData\Data\%VOICES_2%.bsa"
        set "SHIVERING_ISLES_BSA_OBRE=!DIRECTORY_OBRE!\Dev\ObvData\Data\%SHIVERING_ISLES%.bsa"
        set "KNIGHTS_BSA_OBRE=!DIRECTORY_OBRE!\Dev\ObvData\Data\%KNIGHTS%.bsa"
        :: Optional DLC
        set "DLC_1_BSA_OBRE=!DIRECTORY_OBRE!\Dev\ObvData\Data\%DLC_1%.bsa"
        set "DLC_2_BSA_OBRE=!DIRECTORY_OBRE!\Dev\ObvData\Data\%DLC_2%.bsa"
        set "DLC_3_BSA_OBRE=!DIRECTORY_OBRE!\Dev\ObvData\Data\%DLC_3%.bsa"
        set "DLC_4_BSA_OBRE=!DIRECTORY_OBRE!\Dev\ObvData\Data\%DLC_4%.bsa"
    ) else (
        call :throw_error "ERROR: Could not find .bsa files for Oblivion Remastered"
    )
) else (
    echo INFO: Found backup files!
)

if exist "!DIRECTORY_OBRE!\Paks\OblivionRemastered-Windows.pak" (
    :: Steam Version
    set "OBRE_PAK=!DIRECTORY_OBRE!\Paks\OblivionRemastered-Windows.pak"
) else if exist "!DIRECTORY_OBRE!\Paks\OblivionRemastered-WinGDK.pak" (
    :: Xbox Gamepass Version
    set "OBRE_PAK=!DIRECTORY_OBRE!\Paks\OblivionRemastered-WinGDK.pak"
) else (
    call :throw_error "ERROR: Could not find .pak file for Oblivion Remastered"
)

set "TMP_DIR=%~dp0tmp"

:: Create folders for temp files and final mod files
set "RESULT_FOLDER=Mod\Content"
set "RESULT_FOLDER_DATA=%RESULT_FOLDER%\Dev\ObvData\Data"
set "RESULT_FOLDER_PAK=%RESULT_FOLDER%\Paks\~mods"

call :check_and_create_folder "%TMP_DIR%"
call :check_and_create_folder "%RESULT_FOLDER_DATA%"
call :check_and_create_folder "%RESULT_FOLDER_PAK%"

:: Define paths of extract and convert folders
set "EXTRACT_FOLDER_BSA_ORIGINAL=%TMP_DIR%\bsa_original"
set "EXTRACT_FOLDER_BSA_REMASTER=%TMP_DIR%\bsa_remaster"
set "EXTRACT_FOLDER_PAK_REMASTER=%TMP_DIR%\pak"

set "CONVERT_FOLDER_TO_CONVERT=%TMP_DIR%\toConvert"
set "CONVERT_FOLDER_WAV=%TMP_DIR%\wav"
set "CONVERT_FOLDER_WEM=%TMP_DIR%\wem"

set "CONVERT_FOLDER_BNK=%TMP_DIR%\bnk"
set "CONVERT_FOLDER_BNK_EVENT=%CONVERT_FOLDER_BNK%\Content\WwiseAudio\Event"
set "CONVERT_FOLDER_BNK_MEDIA=%CONVERT_FOLDER_BNK%\Content\WwiseAudio\Media"

:: Check for last step file
set SUCCESSFUL_STEP=0
if exist "%LAST_SUCCESSFUL_STEP_FILE%" (
    for /f "usebackq tokens=1,* delims==" %%A in ("%LAST_SUCCESSFUL_STEP_FILE%") do (
        set SUCCESSFUL_STEP=%%A
    )
) else (
    echo 0>"%LAST_SUCCESSFUL_STEP_FILE%"
)


:: =================================================================
:: Extract original .bsa files
:: =================================================================

if !SUCCESSFUL_STEP! == 0 (
    echo STEP: Extracting .bsa files from original Oblivion...
    
    if defined DLC_1_BSA_ORIGINAL (
        cmd /c .\tools\BSArch\bsa-multi.exe -o "%EXTRACT_FOLDER_BSA_ORIGINAL%" "%VOICES_1_BSA_ORIGINAL%" "%VOICES_2_BSA_ORIGINAL%" "%SHIVERING_ISLES_BSA_ORIGINAL%" "%KNIGHTS_BSA_ORIGINAL%" "%DLC_1_BSA_ORIGINAL%" "%DLC_2_BSA_ORIGINAL%" "%DLC_3_BSA_ORIGINAL%" "%DLC_4_BSA_ORIGINAL%"
    ) else (
        cmd /c .\tools\BSArch\bsa-multi.exe -o "%EXTRACT_FOLDER_BSA_ORIGINAL%" "%VOICES_1_BSA_ORIGINAL%" "%VOICES_2_BSA_ORIGINAL%" "%SHIVERING_ISLES_BSA_ORIGINAL%" "%KNIGHTS_BSA_ORIGINAL%"
    )

    echo INFO: Unused files are being removed...

    :: We only need mp3 files from sound/voice 
    rd /s /q "%EXTRACT_FOLDER_BSA_ORIGINAL%\meshes" >nul
    rd /s /q "%EXTRACT_FOLDER_BSA_ORIGINAL%\textures" >nul
    rd /s /q "%EXTRACT_FOLDER_BSA_ORIGINAL%\sound\fx" >nul
    del /S /Q "%EXTRACT_FOLDER_BSA_ORIGINAL%\sound\voice\*.lip" >nul

    if not exist "%EXTRACT_FOLDER_BSA_ORIGINAL%\sound" (
        call :throw_error "ERROR: Could not extract .bsa files of original Oblivion"
    )

    call :update_last_step 1
)


:: =================================================================
:: Extract remaster .bsa files
:: =================================================================

if !SUCCESSFUL_STEP! == 1 (
    echo STEP: Extracting .bsa files from Oblivion Remastered...

    cmd /c .\tools\BSArch\bsa-multi.exe -o1 "%EXTRACT_FOLDER_BSA_REMASTER%\%VOICES_1%.bsa" -o2 "%EXTRACT_FOLDER_BSA_REMASTER%\%VOICES_2%.bsa" -o3 "%EXTRACT_FOLDER_BSA_REMASTER%\%SHIVERING_ISLES%.bsa" -o4 "%EXTRACT_FOLDER_BSA_REMASTER%\%KNIGHTS%.bsa" -o5 "%EXTRACT_FOLDER_BSA_REMASTER%\%DLC_1%.bsa" -o6 "%EXTRACT_FOLDER_BSA_REMASTER%\%DLC_2%.bsa" -o7 "%EXTRACT_FOLDER_BSA_REMASTER%\%DLC_3%.bsa" -o8 "%EXTRACT_FOLDER_BSA_REMASTER%\%DLC_4%.bsa" "%VOICES_1_BSA_OBRE%" "%VOICES_2_BSA_OBRE%" "%SHIVERING_ISLES_BSA_OBRE%" "%KNIGHTS_BSA_OBRE%" "%DLC_1_BSA_OBRE%" "%DLC_2_BSA_OBRE%" "%DLC_3_BSA_OBRE%" "%DLC_4_BSA_OBRE%"

    if not exist "%EXTRACT_FOLDER_BSA_REMASTER%\%VOICES_1%.bsa" (
        call :throw_error "ERROR: Could not extract .bsa files of Oblivion Remastered"
    )

    call :update_last_step 2
)


:: =================================================================
:: Extract remaster pak file
:: =================================================================

if !SUCCESSFUL_STEP! == 2 (
    echo STEP: Extracting .pak file from Oblivion Remastered...

    cmd /c .\tools\repak\repak.exe unpack "%OBRE_PAK%" -o "%EXTRACT_FOLDER_PAK_REMASTER%"

    if not exist "%EXTRACT_FOLDER_PAK_REMASTER%" (
        call :throw_error "ERROR: Could not extract .pak file of Oblivion Remastered"
    )

    call :update_last_step 3
)


:: =================================================================
:: Copy mp3s to folder with files to convert
:: =================================================================
if !SUCCESSFUL_STEP! == 3 (
    echo STEP: Renaming localized folders...

    cmd /c .\tools\voxmeld\rename-localized-folders.exe

    call :update_last_step 4
)


:: =================================================================
:: Apply replace fix
:: =================================================================

if !SUCCESSFUL_STEP! == 4 (
    echo STEP: Applying replace fix...

    cmd /c .\tools\voxmeld\apply-replace-fix.exe

    call :update_last_step 5
)


:: =================================================================
:: Apply voice fix
:: =================================================================

if !SUCCESSFUL_STEP! == 5 (
    echo STEP: Applying voice fix...

    cmd /c .\tools\voxmeld\apply-voice-fix.exe

    call :update_last_step 6
)


:: =================================================================
:: Copy files to convert in the 'toConvert' folder
:: =================================================================

if !SUCCESSFUL_STEP! == 6 (
    echo STEP: Preparing files to convert...

    :: Copy intro and outro
    call :check_and_create_folder "%CONVERT_FOLDER_TO_CONVERT%"
    copy "!DIRECTORY_ORIGINAL!\Video\OblivionIntro.bik" "%CONVERT_FOLDER_TO_CONVERT%\scripted_intro_play.bik"
    copy "!DIRECTORY_ORIGINAL!\Video\OblivionOutro.bik" "%CONVERT_FOLDER_TO_CONVERT%\scripted_outro_play.bik"

    cmd /c .\tools\voxmeld\change-prefix-move-mp3s.exe

    if exist "%CONVERT_FOLDER_TO_CONVERT%" (
        set AMOUNT_TO_CONVERT_AFTER=0
        for /f %%A in ('dir /a-d /b "%CONVERT_FOLDER_TO_CONVERT%" 2^>nul ^| find /v /c ""') do set AMOUNT_TO_CONVERT_AFTER=%%A

        if !AMOUNT_TO_CONVERT_AFTER! NEQ !EXPECTED_AMOUNT_AUDIOS! (
            if "!IGNORE_MISMATCH!" == "true" (
                echo INFO: Incorrect amount of files to convert found.
                echo INFO: Will continue since 'ignore mismatch' setting is enabled
            ) else (
                call :throw_error "ERROR: Amount of files to convert does not match the expected amount. Expected: !EXPECTED_AMOUNT_AUDIOS!, Got: !AMOUNT_TO_CONVERT_AFTER!"
            )
        )

        call :update_last_step 7
    ) else (
        call :throw_error "ERROR: Could not find folder with files to convert"
    )
)


:: =================================================================
:: Pack .bsa files with voice fix included
:: =================================================================

if !SUCCESSFUL_STEP! == 7 (
    echo STEP: Creating .bsa files...

    for /d %%F in ("%EXTRACT_FOLDER_BSA_REMASTER%\*") do (
        cmd /c .\tools\BSArch\BSArch.exe pack "%%F" "..\..\..\%RESULT_FOLDER_DATA%\%%~nF.bsa" -tes4 -share -mt 
    )

    set AMOUNT_BSA_AFTER=0
    if exist "%RESULT_FOLDER_DATA%" (
        for %%F in ("%RESULT_FOLDER_DATA%\*") do (
            if not exist "%%F\" (
                set /a AMOUNT_BSA_AFTER+=1
            )
        )

        if !AMOUNT_BSA_AFTER! == 0 (
            call :throw_error "ERROR: Could not create .bsa files"
        )

        :: The bsa extract folders won't be needed anymore
        if "!REMOVE_TEMP_FILES!" == "true" (
            echo INFO: Temporary files are being removed...

            rd /s /q "%EXTRACT_FOLDER_BSA_ORIGINAL%" >nul
            rd /s /q "%EXTRACT_FOLDER_BSA_REMASTER%" >nul
        )

        call :update_last_step 8
    ) else (
        call :throw_error "ERROR: Could not find folder with final .bsa files"
    )
)


:: =================================================================
:: Convert mp3 and video files to wav and wem
:: =================================================================

if !SUCCESSFUL_STEP! == 8 (
    echo STEP: Converting files to .wem... 

    :: Convert all MP3s to WEMs with Vorbis codec
    cmd /c .\tools\sound2wem\sound2wem-go.exe "%CONVERT_FOLDER_TO_CONVERT%\*"

    :: Rename folder for to wem
    if exist "%CONVERT_FOLDER_WEM%\..\Windows" (
        ren "%CONVERT_FOLDER_WEM%\..\Windows" "wem"
    )

    if exist "%CONVERT_FOLDER_WEM%" (
        set AMOUNT_WEM_AFTER=0
        for /f %%A in ('dir /a-d /b "%CONVERT_FOLDER_WEM%" 2^>nul ^| find /v /c ""') do set AMOUNT_WEM_AFTER=%%A

        if !AMOUNT_WEM_AFTER! EQU 0 (
            set AMOUNT_WAV_AFTER=0
            if exist "%CONVERT_FOLDER_WAV%" (
                for /f %%A in ('dir /a-d /b "%CONVERT_FOLDER_WAV%" 2^>nul ^| find /v /c ""') do set AMOUNT_WAV_AFTER=%%A
				
				if !AMOUNT_WAV_AFTER! EQU 0 (
					call :throw_error "ERROR: Could not create .wav files"
				)

				if !AMOUNT_WAV_AFTER! NEQ !EXPECTED_AMOUNT_AUDIOS! (
                    if "!IGNORE_MISMATCH!" == "true" (
                        echo INFO: Incorrect amount of .wav files found.
                        echo INFO: Will continue since 'ignore mismatch' setting is enabled
                    ) else (
                        echo ERROR: !AMOUNT_WAV_AFTER! .wav files does not match the expected amount
					    call :throw_error "This probably means that there was an error while converting a file"
                    )
				)

				call :throw_error "ERROR: An unknown error occured while converting .wav files"
            ) else (
				call :throw_error "ERROR: Could not find wav folder"
			) 
        )

        if !AMOUNT_WEM_AFTER! NEQ !EXPECTED_AMOUNT_AUDIOS! (
            if "!IGNORE_MISMATCH!" == "true" (
                echo INFO: Incorrect amount of .wav files found.
                echo INFO: Will continue since 'ignore mismatch' setting is enabled
            ) else (
                echo ERROR: !AMOUNT_WEM_AFTER! .wem files does not match the expected amount
                call :throw_error "This probably means that there was an error while converting a file"
            )
        )

        echo INFO: Successfully created !AMOUNT_WEM_AFTER! .wem files.

        :: The 'toConvert' folder is no longer needed, so we can delete it to save space
        if "!REMOVE_TEMP_FILES!" == "true" (
            echo INFO: Temporary files are being removed...

            rd /s /q "%CONVERT_FOLDER_TO_CONVERT%" >nul
            rd /s /q "%CONVERT_FOLDER_WAV%" >nul
        )

        call :update_last_step 9
    ) else (
        call :throw_error "ERROR: Could not find folder with .wem files"
    )
)


:: =================================================================
:: Generate .bnk files
:: =================================================================

if !SUCCESSFUL_STEP! == 9 (
    echo STEP: Creating .bnk files...

    :: Create .bnk files
    cmd /c .\tools\voxmeld\voxmeld.exe

	if exist "%CONVERT_FOLDER_BNK_EVENT%\English(US)" (
		set AMOUNT_BNK_AFTER=0
		for /f %%A in ('dir /a-d /b "%CONVERT_FOLDER_BNK_EVENT%\English(US)" 2^>nul ^| find /v /c ""') do set AMOUNT_BNK_AFTER=%%A

		if !AMOUNT_BNK_AFTER! EQU 0 (
			call :throw_error "ERROR: Could not create .bnk files"
		)

		:: Add 2 for video bnks
		set /a AMOUNT_BNK_AFTER=!AMOUNT_BNK_AFTER! + 2

		if !AMOUNT_BNK_AFTER! NEQ !EXPECTED_AMOUNT_BNKS! (
            if "!IGNORE_MISMATCH!" == "true" (
                echo INFO: Incorrect amount of .bnk files found.
                echo INFO: Will continue since 'ignore mismatch' setting is enabled
            ) else (
                call :throw_error "ERROR: !AMOUNT_BNK_AFTER! .bnk files does not match the expected amount"
            )
		)

		echo INFO: Successfully created !AMOUNT_BNK_AFTER! .bnk files.

		call :update_last_step 10
	) else (
        call :throw_error "ERROR: Could not find folder with .bnk files"
    )
)


:: =================================================================
:: Build new .pak .and bsa files
:: =================================================================

if !SUCCESSFUL_STEP! == 10 (
    :: Final step. Build the mod PAK file
    echo STEP: Creating .pak file...
    cmd /c .\tools\repak\repak.exe pack -m "../../../OblivionRemastered" --version V11 "%CONVERT_FOLDER_BNK%\\" "%RESULT_FOLDER_PAK%\german-voices-oblivion-remastered-voxmeld_!VERSION!_P.pak"

    if exist "%RESULT_FOLDER_PAK%\german-voices-oblivion-remastered-voxmeld_!VERSION!_P.pak" (
        set size=0
        for %%A in ("%RESULT_FOLDER_PAK%\german-voices-oblivion-remastered-voxmeld_!VERSION!_P.pak") do set size=%%~zA

        :: Check if file is bigger than 10 MB. Broke files usually only are a few KB in size
        if !size! GEQ 10485760 (
            call :update_last_step 0
            echo INFO: The .pak creation was successful!
            :: Delete rest of temporary files
            if "!REMOVE_TEMP_FILES!" == "true" (
                echo INFO: Temporary files are being removed...

                rd /s /q "%TMP_DIR%" >nul
            )
            
            echo INFO: Mod creation was successful!
            call "%~dp0tools\scripts\batch\install-mod.bat"
            call :throw_error "You can close the window now"
        ) else (
            call :throw_error "ERROR: The created .pak file is less than 10MB!"
        )
    ) else (
        call :throw_error "ERROR: The .pak could not be created"
    )
)

call :throw_error "ERROR: An unknown error occured. Last working step: !SUCCESSFUL_STEP!"

:update_last_step
set SUCCESSFUL_STEP=%1
> "%LAST_SUCCESSFUL_STEP_FILE%" echo %1

exit /b

:check_and_create_folder
if not exist "%~1\" (
    echo INFO: Creating folder "%~1\"
    mkdir "%~1\"
)

exit /b

:throw_error
    echo %1
    pause
    exit
exit