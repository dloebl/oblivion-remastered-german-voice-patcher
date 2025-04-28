chcp 1252
:: ATTENTION:
:: PLEASE UPDATE THE BELOW SECTION WITH YOUR PATHS TO THE GAME FILES AND THE PATH TO THE UNREAL ENGINE 5 BINARIES
:: ALL FOUR PATHS HAVE TO BE UPDATED IN ORDER FOR THIS SCRIPT TO WORK
set VOICES_1_BSA="D:\SteamLibrary\steamapps\common\Oblivion\Data\Oblivion - Voices1.bsa"
set VOICES_2_BSA="D:\SteamLibrary\steamapps\common\Oblivion\Data\Oblivion - Voices2.bsa"
set SHIVERING_ISLES_BSA="D:\SteamLibrary\steamapps\common\Oblivion\Data\DLCShiveringIsles - Voices.bsa"
set KNIGHTS_BSA="D:\SteamLibrary\steamapps\common\Oblivion\Data\Knights.bsa"
set OBRE_PAK="D:\SteamLibrary\steamapps\common\Oblivion Remastered\OblivionRemastered\Content\Paks\OblivionRemastered-Windows.pak"
set UNREAL_BIN_DIR=D:\EpicGames\UE_5.5\Engine\Binaries\Win64

:: Extract the original MP3s from both BSA voice files
mkdir tmp\
set TMP_DIR=%CD%\tmp\
.\BSArch\BSArch.exe unpack %VOICES_1_BSA% tmp\ -mt
.\BSArch\BSArch.exe unpack %VOICES_2_BSA% tmp\ -mt
.\BSArch\BSArch.exe unpack %SHIVERING_ISLES_BSA% tmp\ -mt
.\BSArch\BSArch.exe unpack %KNIGHTS_BSA% tmp\ -mt
:: Extract the BNKs from the OblivionRemastered-Windows.pak
set UNREAL_PAK_EXE=%UNREAL_BIN_DIR%\UnrealPak.exe
%UNREAL_PAK_EXE% -Extract %OBRE_PAK% %CD%\tmp\pak\
:: Copy all MP3s to the MP3 to WEM input folder
ren "tmp\sound\voice\oblivion.esm\dunkler verf*" dark_seducer
ren "tmp\sound\voice\oblivion.esm\goldener heiliger" golden_saint
.\busybox\busybox.exe bash scripts\change-prefix-move-mp3s.sh
:: Convert all MP3s to WEMs with Vorbis codec (this is going to take quite a while)
set TMP_DIR=%CD%\tmp\
cmd /c .\sound2wem\sound2wem.cmd "%TMP_DIR%\MP3s\*"
:: Patch the BNKs, update the WEMs file names and copy everything to the output folder in one go
.\busybox\busybox.exe bash scripts\patch-bnks-copy-out.sh
:: Final step. Build the mod PAK file
cmd /c .\scripts\create_pak.bat "%CD%\german-voices-oblivion-remastered-voxmeld_v0.2.0_P\"
pause
exit
