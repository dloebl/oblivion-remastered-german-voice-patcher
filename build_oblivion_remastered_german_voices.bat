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
if not exist "%UNREAL_BIN_DIR%" (
    echo ERROR: Could not find Unreal Engine with the given path
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
if not exist "%UNREAL_BIN_DIR%\UnrealPak.exe" (
    echo ERROR: Could not find Unreal Engine with the given path. This probably means that you did not set to correct path in the 'paths.bat' file
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

set "RESULT_FOLDER_DATA=ModFiles\Content\Dev\ObvData\Data"
set "RESULT_FOLDER_PAK=ModFiles\Content\Paks\~mods"
set "TMP_DIR=%~dp0tmp"

:: Create folders for temp files and final mod files
mkdir tmp\
mkdir "%RESULT_FOLDER_DATA%\"
mkdir "%RESULT_FOLDER_PAK%\"


if not exist "%RESULT_FOLDER_DATA%\sound" (
    :: Extract the remaster .bsa files with VO
    .\BSArch\BSArch.exe unpack "%VOICES_1_BSA_OBRE%" "%RESULT_FOLDER_DATA%\" -mt
    .\BSArch\BSArch.exe unpack "%VOICES_2_BSA_OBRE%" "%RESULT_FOLDER_DATA%\" -mt
    .\BSArch\BSArch.exe unpack "%SHIVERING_ISLES_BSA_OBRE%" "%RESULT_FOLDER_DATA%\" -mt
    .\BSArch\BSArch.exe unpack "%KNIGHTS_BSA_OBRE%" "%RESULT_FOLDER_DATA%\" -mt

    .\BSArch\BSArch.exe unpack "%DLC_1_BSA_OBRE%" "%RESULT_FOLDER_DATA%\" -mt
    .\BSArch\BSArch.exe unpack "%DLC_2_BSA_OBRE%" "%RESULT_FOLDER_DATA%\" -mt
    .\BSArch\BSArch.exe unpack "%DLC_3_BSA_OBRE%" "%RESULT_FOLDER_DATA%\" -mt
    .\BSArch\BSArch.exe unpack "%DLC_4_BSA_OBRE%" "%RESULT_FOLDER_DATA%\" -mt

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

if not exist "%TMP_DIR%\sound" (
    :: Extract the original MP3s from all original .bsa voice files
    .\BSArch\BSArch.exe unpack "%VOICES_1_BSA_ORIGINAL%" tmp\ -mt
    .\BSArch\BSArch.exe unpack "%VOICES_2_BSA_ORIGINAL%" tmp\ -mt
    .\BSArch\BSArch.exe unpack "%SHIVERING_ISLES_BSA_ORIGINAL%" tmp\ -mt
    .\BSArch\BSArch.exe unpack "%KNIGHTS_BSA_ORIGINAL%" tmp\ -mt
    
    :: Optional: DLCs
    .\BSArch\BSArch.exe unpack "%DLC_1_BSA_ORIGINAL%" "%TMP_DIR%" -mt
    .\BSArch\BSArch.exe unpack "%DLC_2_BSA_ORIGINAL%" "%TMP_DIR%" -mt
    .\BSArch\BSArch.exe unpack "%DLC_3_BSA_ORIGINAL%" "%TMP_DIR%" -mt
    .\BSArch\BSArch.exe unpack "%DLC_4_BSA_ORIGINAL%" "%TMP_DIR%" -mt

    :: Custom voice lines
    :: .\BSArch\BSArch.exe unpack "%CUSTOM_BSA%" tmp\ -mt

    :: Copy intro and outro
    mkdir "%TMP_DIR%\MP3s\"
    copy "%DIRECTORY_ORIGINAL%\Video\OblivionIntro.bik" "%TMP_DIR%\MP3s\scripted_intro_play.bik"
    copy "%DIRECTORY_ORIGINAL%\Video\OblivionOutro.bik" "%TMP_DIR%\MP3s\scripted_outro_play.bik"

    :: We only need mp3 files from sound/voice 
    rd /s /q "%TMP_DIR%\meshes"
    rd /s /q "%TMP_DIR%\sound\fx"
    rd /s /q "%TMP_DIR%\textures"
    del /S /Q "%TMP_DIR%\sound\voice\*.lip"

    if not exist "%TMP_DIR%\sound" (
        echo ERROR: Could not find extracted .bsa files of Oblivion
        pause
        exit
    )
)

if not exist "%TMP_DIR%\pak" (
    :: Extract the BNKs from the OblivionRemastered-Windows.pak
    "%UNREAL_BIN_DIR%\UnrealPak.exe" -Extract "%OBRE_PAK%" "%TMP_DIR%\pak"

    if not exist "%TMP_DIR%\pak" (
        echo ERROR: Could not find extracted .pak file data of Oblivion Remastered
        pause
        exit
    )
)

:: Check amount of wem files. Below 47000 would mean that most likely files are missing or the code did not run yet
set AMOUNT_WEM_BEFORE=0
if exist "%~dp0sound2wem\Windows" (
    for /f %%A in ('dir /a-d /b "%~dp0sound2wem\Windows" 2^>nul ^| find /v /c ""') do set AMOUNT_WEM_BEFORE=%%A
)

if !AMOUNT_WEM_BEFORE! lss 47000 (
    :: Check amount of mp3 files. Below 47000 would mean that most likely files are missing or the code did not run yet
    set AMOUNT_MP3_BEFORE=0
    if exist "%TMP_DIR%\MP3s" (
        for /f %%A in ('dir /a-d /b "%TMP_DIR%\MP3s" 2^>nul ^| find /v /c ""') do set AMOUNT_MP3_BEFORE=%%A
    )

    if !AMOUNT_MP3_BEFORE! lss 47000 (
        :: Copy all mp3 files to their respective folders
        .\voxmeld\change-prefix-move-mp3s.exe

        set AMOUNT_MP3_AFTER=0
        if exist "%TMP_DIR%\MP3s" (
            for /f %%A in ('dir /a-d /b "%TMP_DIR%\MP3s" 2^>nul ^| find /v /c ""') do set AMOUNT_MP3_AFTER=%%A
        )

        if !AMOUNT_MP3_AFTER! lss 47000 (
            echo ERROR: Could not copy over .mp3 files correctly
            pause
            exit
            
        ) else (
            :: The bsa extract folder won't be needed anymore
            if %REMOVE_TEMP_FILES% == "true" (
                rd /s /q "%TMP_DIR%\sound"
            )
        )
    )

    :: Convert all MP3s to WEMs with Vorbis codec (this is going to take quite a while)
    .\sound2wem\sound2wem.exe "%TMP_DIR%\MP3s\*"

    set AMOUNT_WEM_AFTER=0
    if exist "%~dp0sound2wem\Windows" (
        for /f %%A in ('dir /a-d /b "%~dp0sound2wem\Windows" 2^>nul ^| find /v /c ""') do set AMOUNT_WEM_AFTER=%%A
    )

    if !AMOUNT_WEM_AFTER! lss 47000 (
        echo ERROR: Could not convert .mp3 files correctly
        pause
        exit
        
    ) else (
        :: The MP3s folder is no longer needed, so we can delete it to save space
        if %REMOVE_TEMP_FILES% == "true" (
            rd /s /q "%TMP_DIR%\MP3s"
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
    .\busybox\busybox.exe bash scripts\patch-bnks-copy-out.sh

    set AMOUNT_BNK_AFTER=0
    if exist "%~dp0german-voices-oblivion-remastered-voxmeld_v%VERSION_NUMBER%_P\Content\WwiseAudio\Event\English(US)" (
        for /f %%A in ('dir /a-d /b "%~dp0german-voices-oblivion-remastered-voxmeld_v%VERSION_NUMBER%_P\Content\WwiseAudio\Event\English(US)" 2^>nul ^| find /v /c ""') do set AMOUNT_BNK_AFTER=%%A
    )

    if !AMOUNT_BNK_AFTER! lss 133000 (
        echo ERROR: Could not create bnk files correctly
        pause
        exit
        
    )
)

if %EXECUTE_MP3_DIFF_SCRIPT% == "true" (
    .\busybox\busybox.exe bash scripts\check-missing-wems.sh
)

:: Final step. Build the mod PAK file
cmd /c .\scripts\create_pak.bat "%CD%\german-voices-oblivion-remastered-voxmeld_v%VERSION_NUMBER%_P\"

set size=0
if exist "%RESULT_FOLDER_PAK%\german-voices-oblivion-remastered-voxmeld_v%VERSION_NUMBER%_P.pak" (
    for %%A in ("%RESULT_FOLDER_PAK%\german-voices-oblivion-remastered-voxmeld_v%VERSION_NUMBER%_P.pak") do set size=%%~zA
)

:: Check if file is bigger than 10 MB
if "%size%" GTR "10485760" (
    echo Die .pak Datei wurde erfolgreich erstellt. Bitte kopiere den ganzen 'Content' Ordner aus dem 'Modfiles' Ordner in dein Spielverzeichnis!
    :: Delete rest of temporary files
    if %REMOVE_TEMP_FILES% == "true" (
        rd /s /q "%TMP_DIR%"
        rd /s /q "%~dp0\sound2wem\audiotemp"
        rd /s /q "%~dp0\sound2wem\Windows"
        rd /s /q "%~dp0\german-voices-oblivion-remastered-voxmeld_v%VERSION_NUMBER%_P"
    )
) else (
    echo ERROR: The created .pak file is less than 10MB!
)

pause
exit
