@echo off
chcp 65001
setlocal enabledelayedexpansion

cd /d "%~dp0"

call "%~dp0paths.bat"
call "%~dp0tools\scripts\batch\get-settings.bat"
cls


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

if not exist "%VOICES_1_BSA_ORIGINAL%" (
    call :throw_error "ERROR: Could not find .bsa files for original Oblivion"
)

for %%A in ("%VOICES_1_BSA_ORIGINAL%") do set "size_base=%%~zA"
if !size_base! NEQ !EXPECTED_SIZE_BASE! (
    echo ERROR: Original Oblivion has incorrect size. Amount: !size_base! bytes
    call :throw_error "This probably means that your original oblivion files are in english"
)

echo INFO: Checking for optional DLC in original Oblivion

if exist "!DIRECTORY_ORIGINAL!\%DLC_1%.bsa" (
    for %%A in ("!DIRECTORY_ORIGINAL!\%DLC_1%.bsa") do set "size_dlc=%%~zA"
    
    :: Check if DLC has expected size for german translation
    if !size_dlc! EQU !EXPECTED_SIZE_DLC! (
        echo INFO: DLC for original Oblivion were found!

        set "DLC_1_BSA_ORIGINAL=!DIRECTORY_ORIGINAL!\%DLC_1%.bsa"
        set "DLC_2_BSA_ORIGINAL=!DIRECTORY_ORIGINAL!\%DLC_2%.bsa"
        set "DLC_3_BSA_ORIGINAL=!DIRECTORY_ORIGINAL!\%DLC_3%.bsa"
        set "DLC_4_BSA_ORIGINAL=!DIRECTORY_ORIGINAL!\%DLC_4%.bsa"

        :: Change amounts to use DLC count for later checks
        set EXPECTED_AMOUNT_AUDIOS=!EXPECTED_AMOUNT_AUDIOS_WITH_DLC!
        set EXPECTED_AMOUNT_BNKS=!EXPECTED_AMOUNT_BNKS_WITH_DLC!
    ) else (
        set SIZE_ENGLISH_DLC=4018605

        if !size_dlc! EQU %SIZE_ENGLISH_DLC% (
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

set "VOICES_1_BSA_OBRE=!DIRECTORY_OBRE!\Dev\ObvData\Data\%VOICES_1%.bsa"
set "VOICES_2_BSA_OBRE=!DIRECTORY_OBRE!\Dev\ObvData\Data\%VOICES_2%.bsa"
set "SHIVERING_ISLES_BSA_OBRE=!DIRECTORY_OBRE!\Dev\ObvData\Data\%SHIVERING_ISLES%.bsa"
set "KNIGHTS_BSA_OBRE=!DIRECTORY_OBRE!\Dev\ObvData\Data\%KNIGHTS%.bsa"
:: Optional DLC
set "DLC_1_BSA_OBRE=!DIRECTORY_OBRE!\Dev\ObvData\Data\%DLC_1%.bsa"
set "DLC_2_BSA_OBRE=!DIRECTORY_OBRE!\Dev\ObvData\Data\%DLC_2%.bsa"
set "DLC_3_BSA_OBRE=!DIRECTORY_OBRE!\Dev\ObvData\Data\%DLC_3%.bsa"
set "DLC_4_BSA_OBRE=!DIRECTORY_OBRE!\Dev\ObvData\Data\%DLC_4%.bsa"

if not exist "%VOICES_1_BSA_OBRE%" (
    call :throw_error "ERROR: Could not find .bsa files for Oblivion Remastered"
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


if !SUCCESSFUL_STEP! == 0 (
    echo STEP: Extracting .bsa files from Oblivion Remastered...

    :: Extract the remaster .bsa files with VO
    cmd /c .\tools\BSArch\bsa-multi.exe -o "%RESULT_FOLDER_DATA%" "%VOICES_1_BSA_OBRE%" "%VOICES_2_BSA_OBRE%" "%SHIVERING_ISLES_BSA_OBRE%" "%KNIGHTS_BSA_OBRE%" "%DLC_1_BSA_OBRE%" "%DLC_2_BSA_OBRE%" "%DLC_3_BSA_OBRE%" "%DLC_4_BSA_OBRE%"

    :: We only need mp3 files from sound/voice
    rd /s /q "%RESULT_FOLDER_DATA%\meshes" >nul
    rd /s /q "%RESULT_FOLDER_DATA%\sound\fx" >nul
    rd /s /q "%RESULT_FOLDER_DATA%\textures" >nul
    del /S /Q "%RESULT_FOLDER_DATA%\sound\voice\*.lip" >nul

    if not exist "%RESULT_FOLDER_DATA%\sound" (
        call :throw_error "ERROR: Could not find extracted bsa files of Oblivion Remastered"
    )

    call :update_last_step 1
)

if !SUCCESSFUL_STEP! == 1 (
    :: Extract the original MP3s from all original .bsa voice files
    .\tools\BSArch\bsa-multi.exe -o "%EXTRACT_FOLDER_BSA_ORIGINAL%" "%VOICES_1_BSA_ORIGINAL%" "%VOICES_2_BSA_ORIGINAL%" "%SHIVERING_ISLES_BSA_ORIGINAL%" "%KNIGHTS_BSA_ORIGINAL%" "%DLC_1_BSA_ORIGINAL%" "%DLC_2_BSA_ORIGINAL%" "%DLC_3_BSA_ORIGINAL%" "%DLC_4_BSA_ORIGINAL%"

    :: Copy intro and outro
    call :check_and_create_folder "%CONVERT_FOLDER_TO_CONVERT%"
    copy "!DIRECTORY_ORIGINAL!\Video\OblivionIntro.bik" "%CONVERT_FOLDER_TO_CONVERT%\scripted_intro_play.bik"
    copy "!DIRECTORY_ORIGINAL!\Video\OblivionOutro.bik" "%CONVERT_FOLDER_TO_CONVERT%\scripted_outro_play.bik"

    :: We only need mp3 files from sound/voice 
    rd /s /q "%EXTRACT_FOLDER_BSA_ORIGINAL%\meshes" >nul
    rd /s /q "%EXTRACT_FOLDER_BSA_ORIGINAL%\sound\fx" >nul
    rd /s /q "%EXTRACT_FOLDER_BSA_ORIGINAL%\textures" >nul
    del /S /Q "%EXTRACT_FOLDER_BSA_ORIGINAL%\sound\voice\*.lip" >nul

    if not exist "%EXTRACT_FOLDER_BSA_ORIGINAL%\sound" (
        call :throw_error "ERROR: Could not find extracted .bsa files of Oblivion"
    )
    
    call :update_last_step 2
)

if !SUCCESSFUL_STEP! == 2 (
    :: Extract the BNKs from the OblivionRemastered-Windows.pak
    echo Extracting Pak file from the Oblivion Remastered...

    cmd /c .\tools\repak\repak.exe unpack "%OBRE_PAK%" -o "%EXTRACT_FOLDER_PAK_REMASTER%"

    if not exist "%EXTRACT_FOLDER_PAK_REMASTER%" (
        call :throw_error "ERROR: Could not extract .pak file of Oblivion Remastered"
    )
    
    call :update_last_step 3
)

if !SUCCESSFUL_STEP! == 3 (
    call :update_last_step 4
)

if !SUCCESSFUL_STEP! == 4 (
    call :update_last_step 5
)

if !SUCCESSFUL_STEP! == 5 (
    call :update_last_step 6
)

if !SUCCESSFUL_STEP! == 6 (
    :: Check amount of mp3 files. Below 47000 would mean that most likely files are missing or the code did not run yet

    :: Copy all mp3 files to their respective folders
    cmd /c .\tools\voxmeld\change-prefix-move-mp3s.exe

    if exist "%CONVERT_FOLDER_TO_CONVERT%" (
        set AMOUNT_TO_CONVERT_AFTER=0
        for /f %%A in ('dir /a-d /b "%CONVERT_FOLDER_TO_CONVERT%" 2^>nul ^| find /v /c ""') do set AMOUNT_TO_CONVERT_AFTER=%%A

        if !AMOUNT_TO_CONVERT_AFTER! NEQ !EXPECTED_AMOUNT_AUDIOS! (
            call :throw_error "ERROR: Amount of files to convert does not match the expected amount. Expected: !EXPECTED_AMOUNT_AUDIOS!, Got: !AMOUNT_TO_CONVERT_AFTER!"
        )

        call :update_last_step 7
    ) else (
        call :throw_error "ERROR: Could not find folder with files to convert"
    )
)

if !SUCCESSFUL_STEP! == 7 (
    call :update_last_step 8
)

if !SUCCESSFUL_STEP! == 8 (
    :: Convert all MP3s to WEMs with Vorbis codec (this is going to take quite a while)
    cmd /c .\tools\sound2wem\sound2wem.exe "%CONVERT_FOLDER_TO_CONVERT%\*"

    :: Rename folder from Windows for to wem
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
                    echo ERROR: !AMOUNT_WAV_AFTER! .wav files does not match the expected amount
                    call :throw_error "This probably means that there was an error while converting a file"
				)

				call :throw_error "ERROR: An unknown error occured while converting .wav files"
            ) else (
				call :throw_error "ERROR: Could not find wav folder"
			) 
        )

        if !AMOUNT_WEM_AFTER! NEQ !EXPECTED_AMOUNT_AUDIOS! (
            echo ERROR: !AMOUNT_WEM_AFTER! .wem files does not match the expected amount
            call :throw_error "This probably means that there was an error while converting a file"
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

if !SUCCESSFUL_STEP! == 9 (
    :: Patch the BNKs, update the WEMs file names and copy everything to the output folder in one go
    cmd /c .\tools\voxmeld\voxmeld.exe

    if exist "%CONVERT_FOLDER_BNK_EVENT%\English(US)" (
        set AMOUNT_BNK_AFTER=0
        for /f %%A in ('dir /a-d /b "%CONVERT_FOLDER_BNK_EVENT%\English(US)" 2^>nul ^| find /v /c ""') do set AMOUNT_BNK_AFTER=%%A
        
		if !AMOUNT_BNK_AFTER! EQU 0 (
			call :throw_error "ERROR: Could not create .bnk files"
		)

		:: Add 2 for video bnks in different folder
		set /a AMOUNT_BNK_AFTER=!AMOUNT_BNK_AFTER! + 2

        if !AMOUNT_BNK_AFTER! NEQ !EXPECTED_AMOUNT_BNKS! (
            call :throw_error "ERROR: !AMOUNT_BNK_AFTER! .bnk files does not match the expected amount"
        )

		echo INFO: Successfully created !AMOUNT_BNK_AFTER! .bnk files.
    
        call :update_last_step 10
    ) else (
        call :throw_error "ERROR: Could not find folder with .bnk files"
    )
)

if !SUCCESSFUL_STEP! == 10 (
    echo Building the Mod PAK file...

    :: Final step. Build the mod PAK file
    cmd /c .\tools\repak\repak.exe pack -m "../../../OblivionRemastered" --version V11 "%CONVERT_FOLDER_BNK%\\" "%RESULT_FOLDER_PAK%\german-voices-oblivion-remastered-voxmeld_%VERSION_NUMBER%_P.pak"

    if exist "%RESULT_FOLDER_PAK%\german-voices-oblivion-remastered-voxmeld_v%VERSION_NUMBER%_P.pak" (
        set size=0
        for %%A in ("%RESULT_FOLDER_PAK%\german-voices-oblivion-remastered-voxmeld_v%VERSION_NUMBER%_P.pak") do set size=%%~zA
        :: Check if file is bigger than 10 MB
        if !size! GEQ 10485760 (
            call :update_last_step 0
            echo Die .pak Datei wurde erfolgreich erstellt.
            :: Delete rest of temporary files
            if "!REMOVE_TEMP_FILES!" == "true" (
                echo Temporärdateien werden entfernt...

                rd /s /q "%TMP_DIR%" >nul
            )

            echo Die Mod wurde erfolgreich erstellt!
            echo Bitte kopiere den ganzen 'Content' Ordner aus dem 'Modfiles' Ordner in dein Spielverzeichnis!
            
            call :throw_error "Du kannst die Konsole nun schließen."
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