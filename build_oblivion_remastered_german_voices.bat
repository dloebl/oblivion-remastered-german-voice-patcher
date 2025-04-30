chcp 1252
call paths.bat

set VOICES_1_BSA_ORIGINAL=%DIRECTORY_ORIGINAL%\Oblivion - Voices1.bsa
set VOICES_2_BSA_ORIGINAL=%DIRECTORY_ORIGINAL%\Oblivion - Voices2.bsa
set SHIVERING_ISLES_BSA_ORIGINAL=%DIRECTORY_ORIGINAL%\DLCShiveringIsles - Voices.bsa
set KNIGHTS_BSA_ORIGINAL=%DIRECTORY_ORIGINAL%\Knights.bsa

set VOICES_1_BSA_OBRE=%DIRECTORY_OBRE%\Dev\ObvData\Data\Oblivion - Voices1.bsa
set VOICES_2_BSA_OBRE=%DIRECTORY_OBRE%\Dev\ObvData\Data\Oblivion - Voices2.bsa
set SHIVERING_ISLES_BSA_OBRE=%DIRECTORY_OBRE%\Dev\ObvData\Data\DLCShiveringIsles - Voices.bsa
set KNIGHTS_BSA_OBRE=%DIRECTORY_OBRE%\Dev\ObvData\Data\Knights.bsa

:: Optional DLC
set DLC_1_BSA_ORIGINAL=%DIRECTORY_ORIGINAL%\DLCHorseArmor.bsa
set DLC_2_BSA_ORIGINAL=%DIRECTORY_ORIGINAL%\DLCOrrery.bsa
set DLC_3_BSA_ORIGINAL=%DIRECTORY_ORIGINAL%\DLCThievesDen.bsa
set DLC_4_BSA_ORIGINAL=%DIRECTORY_ORIGINAL%\DLCVilelair.bsa

set DLC_1_BSA_OBRE=%DIRECTORY_OBRE%\Dev\ObvData\Data\DLCHorseArmor.bsa
set DLC_2_BSA_OBRE=%DIRECTORY_OBRE%\Dev\ObvData\Data\DLCOrrery.bsa
set DLC_3_BSA_OBRE=%DIRECTORY_OBRE%\Dev\ObvData\Data\DLCThievesDen.bsa
set DLC_4_BSA_OBRE=%DIRECTORY_OBRE%\Dev\ObvData\Data\DLCVilelair.bsa

if exist "%DIRECTORY_OBRE%\Paks\OblivionRemastered-Windows.pak" (
    :: Steam Version
    set OBRE_PAK=%DIRECTORY_OBRE%\Paks\OblivionRemastered-Windows.pak
) else (
    :: Xbox Gamepass Version
    set OBRE_PAK=%DIRECTORY_OBRE%\Paks\OblivionRemastered-WinGDK.pak
)

set RESULT_FOLDER_DATA=ModFiles\Content\Dev\ObvData\Data
set RESULT_FOLDER_PAK=ModFiles\Content\Paks\~mods

:: Create folders for temp files and final mod files
mkdir tmp\
mkdir "%RESULT_FOLDER_DATA%\"
mkdir "%RESULT_FOLDER_PAK%\"

:: Extract the remaster .bsa files with VO
.\BSArch\BSArch.exe unpack "%VOICES_1_BSA_OBRE%" "%RESULT_FOLDER_DATA%\" -mt
.\BSArch\BSArch.exe unpack "%VOICES_2_BSA_OBRE%" "%RESULT_FOLDER_DATA%\" -mt
.\BSArch\BSArch.exe unpack "%SHIVERING_ISLES_BSA_OBRE%" "%RESULT_FOLDER_DATA%\" -mt
.\BSArch\BSArch.exe unpack "%KNIGHTS_BSA_OBRE%" "%RESULT_FOLDER_DATA%\" -mt
:: Extract the original MP3s from all original .bsa voice files
set TMP_DIR=%CD%\tmp\
.\BSArch\BSArch.exe unpack "%VOICES_1_BSA_ORIGINAL%" tmp\ -mt
.\BSArch\BSArch.exe unpack "%VOICES_2_BSA_ORIGINAL%" tmp\ -mt
.\BSArch\BSArch.exe unpack "%SHIVERING_ISLES_BSA_ORIGINAL%" tmp\ -mt
.\BSArch\BSArch.exe unpack "%KNIGHTS_BSA_ORIGINAL%" tmp\ -mt

:: Check for optional dlc
if exist "%DLC_1_BSA_ORIGINAL%" (
    .\BSArch\BSArch.exe unpack "%DLC_1_BSA_OBRE%" "%RESULT_FOLDER_DATA%\" -mt
    .\BSArch\BSArch.exe unpack "%DLC_1_BSA_ORIGINAL%" tmp\ -mt
)
if exist "%DLC_2_BSA_ORIGINAL%" (
    .\BSArch\BSArch.exe unpack "%DLC_2_BSA_OBRE%" "%RESULT_FOLDER_DATA%\" -mt
    .\BSArch\BSArch.exe unpack "%DLC_2_BSA_ORIGINAL%" tmp\ -mt
)
if exist "%DLC_3_BSA_ORIGINAL%" (
    .\BSArch\BSArch.exe unpack "%DLC_3_BSA_OBRE%" "%RESULT_FOLDER_DATA%\" -mt
   .\BSArch\BSArch.exe unpack "%DLC_3_BSA_ORIGINAL%" tmp\ -mt
)
if exist "%DLC_4_BSA_ORIGINAL%" (
    .\BSArch\BSArch.exe unpack "%DLC_4_BSA_OBRE%" "%RESULT_FOLDER_DATA%\" -mt
    .\BSArch\BSArch.exe unpack "%DLC_4_BSA_ORIGINAL%" tmp\ -mt
)
:: Extract the BNKs from the OblivionRemastered-Windows.pak
set UNREAL_PAK_EXE=%UNREAL_BIN_DIR%\UnrealPak.exe
"%UNREAL_PAK_EXE%" -Extract "%OBRE_PAK%" "%CD%\tmp\pak"
:: Copy all MP3s to the MP3 to WEM input folder and bsa extract folders
.\busybox\busybox.exe bash scripts\change-prefix-move-mp3s.sh
:: Copy intro and outro
copy "%DIRECTORY_ORIGINAL%\Video\OblivionIntro.bik" "%TMP_DIR%MP3s\205096107.bik"
:: copy "%DIRECTORY_ORIGINAL%\Video\OblivionOutro.bik" "%TMP_DIR%MP3s\PlaceholderOutro.bik"
:: Convert all MP3s to WEMs with Vorbis codec (this is going to take quite a while)
cmd /c .\sound2wem\sound2wem.cmd "%TMP_DIR%MP3s\*"
:: Patch the BNKs, update the WEMs file names and copy everything to the output folder in one go
.\busybox\busybox.exe bash scripts\patch-bnks-copy-out.sh
:: Final step. Build the mod PAK file
cmd /c .\scripts\create_pak.bat "%CD%\german-voices-oblivion-remastered-voxmeld_v0.3.2_P\"
pause
exit
