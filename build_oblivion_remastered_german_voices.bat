chcp 1252
:: ATTENTION:
:: PLEASE UPDATE THE BELOW SECTION WITH YOUR PATHS TO THE GAME FILES AND THE PATH TO THE UNREAL ENGINE 5 BINARIES
:: ALL THREE PATHS HAVE TO BE UPDATED IN ORDER FOR THIS SCRIPT TO WORK
set DIRECTORY_ORIGINAL=F:\Steam\SteamApps\common\Oblivion\Data
set DIRECTORY_OBRE=F:\Steam\SteamApps\common\Oblivion Remastered\OblivionRemastered\Content
set UNREAL_BIN_DIR=F:\UE_5.5\Engine\Binaries\Win64
set RESULT_FOLDER=german-voices-oblivion-remastered-voxmeld_v0.2.2_P

set VOICES_1_BSA_ORIGINAL=%DIRECTORY_ORIGINAL%\Oblivion - Voices1.bsa
set VOICES_2_BSA_ORIGINAL=%DIRECTORY_ORIGINAL%\Oblivion - Voices2.bsa
set SHIVERING_ISLES_BSA_ORIGINAL=%DIRECTORY_ORIGINAL%\DLCShiveringIsles - Voices.bsa
set KNIGHTS_BSA_ORIGINAL=%DIRECTORY_ORIGINAL%\Knights.bsa

set VOICES_1_BSA_OBRE=%DIRECTORY_OBRE%\Dev\ObvData\Data\Oblivion - Voices1.bsa
set VOICES_2_BSA_OBRE=%DIRECTORY_OBRE%\Dev\ObvData\Data\Oblivion - Voices2.bsa
set SHIVERING_ISLES_BSA_OBRE=%DIRECTORY_OBRE%\Dev\ObvData\Data\DLCShiveringIsles - Voices.bsa
set KNIGHTS_BSA_OBRE=%DIRECTORY_OBRE%\Dev\ObvData\Data\Knights.bsa
set OBRE_PAK=%DIRECTORY_OBRE%\Paks\OblivionRemastered-Windows.pak

mkdir tmp\
mkdir %RESULT_FOLDER%\

:: Extract the remaster .bsa files with VO
.\BSArch\BSArch.exe unpack "%VOICES_1_BSA_OBRE%" "%RESULT_FOLDER%\" -mt
.\BSArch\BSArch.exe unpack "%VOICES_2_BSA_OBRE%" "%RESULT_FOLDER%\" -mt
.\BSArch\BSArch.exe unpack "%SHIVERING_ISLES_BSA_OBRE%" "%RESULT_FOLDER%\" -mt
.\BSArch\BSArch.exe unpack "%KNIGHTS_BSA_OBRE%" "%RESULT_FOLDER%\" -mt
:: Extract the original MP3s from all original .bsa voice files
set TMP_DIR=%CD%\tmp\
.\BSArch\BSArch.exe unpack "%VOICES_1_BSA_ORIGINAL%" tmp\ -mt
.\BSArch\BSArch.exe unpack "%VOICES_2_BSA_ORIGINAL%" tmp\ -mt
.\BSArch\BSArch.exe unpack "%SHIVERING_ISLES_BSA_ORIGINAL%" tmp\ -mt
.\BSArch\BSArch.exe unpack "%KNIGHTS_BSA_ORIGINAL%" tmp\ -mt
:: Extract the BNKs from the OblivionRemastered-Windows.pak
set UNREAL_PAK_EXE=%UNREAL_BIN_DIR%\UnrealPak.exe
"%UNREAL_PAK_EXE%" -Extract "%OBRE_PAK%" "%CD%\tmp\pak\"
:: Copy all MP3s to the MP3 to WEM input folder and bsa extract folders
.\busybox\busybox.exe bash scripts\change-prefix-move-mp3s.sh
:: Convert all MP3s to WEMs with Vorbis codec (this is going to take quite a while)
set TMP_DIR=%CD%\tmp\
cmd /c .\sound2wem\sound2wem.cmd "%TMP_DIR%\MP3s\*"

mkdir sound2wem\audiotemp\
:: Duplicate wav files if variant exists
.\busybox\busybox.exe bash scripts\duplicate-wavs-altraces-variants.sh
:: Patch the BNKs, update the WEMs file names and copy everything to the output folder in one go
.\busybox\busybox.exe bash scripts\patch-bnks-copy-out.sh
:: Rename .bsa files of remaster so folders are used
rename "%VOICES_1_BSA_OBRE%" "Oblivion - Voices1_bak.bsa"
rename "%VOICES_2_BSA_OBRE%" "Oblivion - Voices2_bak.bsa"
rename "%SHIVERING_ISLES_BSA_OBRE%" "DLCShiveringIsles - Voices_bak.bsa"
rename "%KNIGHTS_BSA_OBRE%" "Knights_bak.bsa"
:: Final step. Build the mod PAK file
cmd /c .\scripts\create_pak.bat "%CD%\%RESULT_FOLDER%\"
pause
exit
