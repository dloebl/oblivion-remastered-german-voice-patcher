@echo off
chcp 1252
call "%~dp0paths.bat"
call "%~dp0scripts\settings.bat"

setlocal enabledelayedexpansion

if not exist "%DIRECTORY_ORIGINAL%" (
    echo ERROR: Could not find Oblivion with the given path
    pause
    exit
)
if not exist "%DIRECTORY_OBRE%" (
    echo ERROR: Could not find Oblivion Remastered with the given path
    pause
    exit
)

set "VOICES_1_BSA_ORIGINAL=%DIRECTORY_ORIGINAL%\Oblivion - Voices1.bsa"
set "VOICES_2_BSA_ORIGINAL=%DIRECTORY_ORIGINAL%\Oblivion - Voices2.bsa"
set "SHIVERING_ISLES_BSA_ORIGINAL=%DIRECTORY_ORIGINAL%\DLCShiveringIsles - Voices.bsa"
set "KNIGHTS_BSA_ORIGINAL=%DIRECTORY_ORIGINAL%\Knights.bsa"

set "VOICES_1_BSA_OBRE=%DIRECTORY_OBRE%\Dev\ObvData\Data\Oblivion - Voices1.bsa"
set "VOICES_2_BSA_OBRE=%DIRECTORY_OBRE%\Dev\ObvData\Data\Oblivion - Voices2.bsa"
set "SHIVERING_ISLES_BSA_OBRE=%DIRECTORY_OBRE%\Dev\ObvData\Data\DLCShiveringIsles - Voices.bsa"
set "KNIGHTS_BSA_OBRE=%DIRECTORY_OBRE%\Dev\ObvData\Data\Knights.bsa"

:: Optional DLC
set "DLC_1_BSA_ORIGINAL=%DIRECTORY_ORIGINAL%\DLCHorseArmor.bsa"
set "DLC_2_BSA_ORIGINAL=%DIRECTORY_ORIGINAL%\DLCOrrery.bsa"
set "DLC_3_BSA_ORIGINAL=%DIRECTORY_ORIGINAL%\DLCThievesDen.bsa"
set "DLC_4_BSA_ORIGINAL=%DIRECTORY_ORIGINAL%\DLCVilelair.bsa"

:: Custom voice lines
set "CUSTOM_BSA=%~dp0custom\Oblivion - VoicesCustom.bsa"

set "DLC_1_BSA_OBRE=%DIRECTORY_OBRE%\Dev\ObvData\Data\DLCHorseArmor.bsa"
set "DLC_2_BSA_OBRE=%DIRECTORY_OBRE%\Dev\ObvData\Data\DLCOrrery.bsa"
set "DLC_3_BSA_OBRE=%DIRECTORY_OBRE%\Dev\ObvData\Data\DLCThievesDen.bsa"
set "DLC_4_BSA_OBRE=%DIRECTORY_OBRE%\Dev\ObvData\Data\DLCVilelair.bsa"

if not exist "%VOICES_1_BSA_ORIGINAL%" (
    echo ERROR: Could not find .bsa files of Oblivion with the given path. This probably means that you did not set the correct path in the 'paths.bat' file
    pause
    exit
)
if not exist "%VOICES_1_BSA_OBRE%" (
    echo ERROR: Could not find .bsa files of Oblivion Remastered with the given path. This probably means that you did not set the correct path in the 'paths.bat' file
    pause
    exit
)


if exist "%DIRECTORY_OBRE%\Paks\OblivionRemastered-Windows.pak" (
    :: Steam Version
    set "OBRE_PAK=%DIRECTORY_OBRE%\Paks\OblivionRemastered-Windows.pak"
) else if exist "%DIRECTORY_OBRE%\Paks\OblivionRemastered-WinGDK.pak" (
    :: Xbox Gamepass Version
    set "OBRE_PAK=%DIRECTORY_OBRE%\Paks\OblivionRemastered-WinGDK.pak"
) else (
    echo ERROR: Could not find .pak file for Oblivion Remastered
    pause
    exit
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


if not exist "%RESULT_FOLDER_DATA%\sound" (
    :: Extract the remaster .bsa files with VO
    .\tools\BSArch\bsa-multi.exe -o "%RESULT_FOLDER_DATA%" "%VOICES_1_BSA_OBRE%" "%VOICES_2_BSA_OBRE%" "%SHIVERING_ISLES_BSA_OBRE%" "%KNIGHTS_BSA_OBRE%" "%DLC_1_BSA_OBRE%" "%DLC_2_BSA_OBRE%" "%DLC_3_BSA_OBRE%" "%DLC_4_BSA_OBRE%"

    :: We only need mp3 files from sound/voice
    rd /s /q "%RESULT_FOLDER_DATA%\meshes"
    rd /s /q "%RESULT_FOLDER_DATA%\sound\fx"
    rd /s /q "%RESULT_FOLDER_DATA%\textures"
    del /S /Q "%RESULT_FOLDER_DATA%\sound\voice\*.lip"

    if not exist "%RESULT_FOLDER_DATA%\sound" (
        echo ERROR: Could not find extracted bsa files of Oblivion Remastered
        pause
        exit
    )
)

if not exist "%EXTRACT_FOLDER_BSA_ORIGINAL%\sound" (
    :: Extract the original MP3s from all original .bsa voice files
    .\tools\BSArch\bsa-multi.exe -o "%EXTRACT_FOLDER_BSA_ORIGINAL%" "%VOICES_1_BSA_ORIGINAL%" "%VOICES_2_BSA_ORIGINAL%" "%SHIVERING_ISLES_BSA_ORIGINAL%" "%KNIGHTS_BSA_ORIGINAL%" "%DLC_1_BSA_ORIGINAL%" "%DLC_2_BSA_ORIGINAL%" "%DLC_3_BSA_ORIGINAL%" "%DLC_4_BSA_ORIGINAL%"

    :: Copy intro and outro
    call :check_and_create_folder "%CONVERT_FOLDER_TO_CONVERT%"
    copy "%DIRECTORY_ORIGINAL%\Video\OblivionIntro.bik" "%CONVERT_FOLDER_TO_CONVERT%\scripted_intro_play.bik"
    copy "%DIRECTORY_ORIGINAL%\Video\OblivionOutro.bik" "%CONVERT_FOLDER_TO_CONVERT%\scripted_outro_play.bik"

    :: We only need mp3 files from sound/voice 
    rd /s /q "%EXTRACT_FOLDER_BSA_ORIGINAL%\meshes"
    rd /s /q "%EXTRACT_FOLDER_BSA_ORIGINAL%\sound\fx"
    rd /s /q "%EXTRACT_FOLDER_BSA_ORIGINAL%\textures"
    del /S /Q "%EXTRACT_FOLDER_BSA_ORIGINAL%\sound\voice\*.lip"

    if not exist "%EXTRACT_FOLDER_BSA_ORIGINAL%\sound" (
        echo ERROR: Could not find extracted .bsa files of Oblivion
        pause
        exit
    )
)

if not exist "%EXTRACT_FOLDER_PAK_REMASTER%" (
    :: Extract the BNKs from the OblivionRemastered-Windows.pak
    echo Extracting Pak file from the Oblivion Remastered...

    .\tools\repak\repak.exe unpack "%OBRE_PAK%" -o "%EXTRACT_FOLDER_PAK_REMASTER%"

    if not exist "%EXTRACT_FOLDER_PAK_REMASTER%" (
        echo ERROR: Could not find extracted .pak file data of Oblivion Remastered
        pause
        exit
    )
)

:: Check amount of wem files. Below 47000 would mean that most likely files are missing or the code did not run yet
set AMOUNT_WEM_BEFORE=0
if exist "%CONVERT_FOLDER_WEM%" (
    for /f %%A in ('dir /a-d /b "%CONVERT_FOLDER_WEM%" 2^>nul ^| find /v /c ""') do set AMOUNT_WEM_BEFORE=%%A
)

if !AMOUNT_WEM_BEFORE! lss 47000 (
    :: Check amount of mp3 files. Below 47000 would mean that most likely files are missing or the code did not run yet
    set AMOUNT_MP3_BEFORE=0
    if exist "%CONVERT_FOLDER_TO_CONVERT%" (
        for /f %%A in ('dir /a-d /b "%CONVERT_FOLDER_TO_CONVERT%" 2^>nul ^| find /v /c ""') do set AMOUNT_MP3_BEFORE=%%A
    )

    if !AMOUNT_MP3_BEFORE! lss 47000 (
        :: Copy all mp3 files to their respective folders
        .\tools\voxmeld\change-prefix-move-mp3s.exe

        set AMOUNT_MP3_AFTER=0
        if exist "%CONVERT_FOLDER_TO_CONVERT%" (
            for /f %%A in ('dir /a-d /b "%CONVERT_FOLDER_TO_CONVERT%" 2^>nul ^| find /v /c ""') do set AMOUNT_MP3_AFTER=%%A
        )

        if !AMOUNT_MP3_AFTER! lss 47000 (
            echo ERROR: Could not copy over .mp3 files correctly
            pause
            exit

        ) else (
            :: The bsa extract folder won't be needed anymore
            if %REMOVE_TEMP_FILES% == "true" (
                rd /s /q "%EXTRACT_FOLDER_BSA_ORIGINAL%\sound"
            )
        )
    )

    :: Convert all MP3s to WEMs with Vorbis codec (this is going to take quite a while)
    .\tools\sound2wem\sound2wem.exe "%CONVERT_FOLDER_TO_CONVERT%\*"

    :: Rename folder from Windows for to wem
    if exist "%CONVERT_FOLDER_WEM%\..\Windows" (
        ren "%CONVERT_FOLDER_WEM%\..\Windows" "wem"
    )
    
    set AMOUNT_WEM_AFTER=0
    if exist "%CONVERT_FOLDER_WEM%" (
        for /f %%A in ('dir /a-d /b "%CONVERT_FOLDER_WEM%" 2^>nul ^| find /v /c ""') do set AMOUNT_WEM_AFTER=%%A
    )

    if !AMOUNT_WEM_AFTER! lss 47000 (
        echo ERROR: Could not convert .mp3 files correctly
        pause
        exit
        
    ) else (
        :: The MP3s folder is no longer needed, so we can delete it to save space
        if %REMOVE_TEMP_FILES% == "true" (
            rd /s /q "%CONVERT_FOLDER_TO_CONVERT%"
        )
    )
)

:: Check amount of bnk files. Below 47000 would mean that most likely files are missing or the code did not run yet
set AMOUNT_BNK_BEFORE=0
if exist "%~dp0german-voices-oblivion-remastered-voxmeld_v%VERSION_NUMBER%_P\Content\WwiseAudio\Event\English(US)" (
    for /f %%A in ('dir /a-d /b "%~dp0german-voices-oblivion-remastered-voxmeld_v%VERSION_NUMBER%_P\Content\WwiseAudio\Event\English(US)" 2^>nul ^| find /v /c ""') do set AMOUNT_BNK_BEFORE=%%A
)

if !AMOUNT_BNK_BEFORE! lss 133000 (
    :: Patch the BNKs, update the WEMs file names and copy everything to the output folder in one go
    cmd /c .\tools\voxmeld\voxmeld.exe

    set AMOUNT_BNK_AFTER=0
    if exist "%CONVERT_FOLDER_BNK_EVENT%\English(US)" (
        for /f %%A in ('dir /a-d /b "%CONVERT_FOLDER_BNK_EVENT%\English(US)" 2^>nul ^| find /v /c ""') do set AMOUNT_BNK_AFTER=%%A
    )

    if !AMOUNT_BNK_AFTER! lss 133000 (
        echo ERROR: Could not create bnk files correctly
        pause
        exit
        
    )
)

if %EXECUTE_MP3_DIFF_SCRIPT% == "true" (
    :: TODO: Switch away from busybox to a native solution
    ::.\busybox\busybox.exe bash scripts\check-missing-wems.sh
)
echo Building the Mod PAK file...
:: Final step. Build the mod PAK file
cmd /c .\tools\repak\repak.exe pack -m "../../../OblivionRemastered" --version V11 "%CONVERT_FOLDER_BNK%\\" "%RESULT_FOLDER_PAK%\german-voices-oblivion-remastered-voxmeld_%VERSION_NUMBER%_P.pak"

set size=0
if exist "%RESULT_FOLDER_PAK%\german-voices-oblivion-remastered-voxmeld_v%VERSION_NUMBER%_P.pak" (
    for %%A in ("%RESULT_FOLDER_PAK%\german-voices-oblivion-remastered-voxmeld_v%VERSION_NUMBER%_P.pak") do set size=%%~zA
)
:: Check if file is bigger than 10 MB
if "%size%" GTR "10485760" (
    echo Die .pak Datei wurde erfolgreich erstellt.
    :: Delete rest of temporary files
    if %REMOVE_TEMP_FILES% == "true" (
        echo Temporärdateien werden entfernt...

        rd /s /q "%TMP_DIR%"
    )

    echo Die Mod wurde erfolgreich erstellt!
    echo Bitte kopiere den ganzen 'Content' Ordner aus dem 'Modfiles' Ordner in dein Spielverzeichnis!
    echo Du kannst die Konsole nun schließen.
) else (
    echo ERROR: The created .pak file is less than 10MB!
)

pause
exit

:check_and_create_folder
if not exist "%~1\" (
    echo INFO: Creating folder "%~1\"
    mkdir "%~1\"
)

exit /b