call paths.bat
set UNREAL_PAK_EXE=%UNREAL_BIN_DIR%\UnrealPak.exe
@setlocal ENABLEDELAYEDEXPANSION
@if "%~1"=="" goto skip
@setlocal enableextensions
set FILE_LIST="%CD%\tmp\filelist.txt"
set OUT_PAK="%CD%\ModFiles\Content\Paks\~mods\german-voices-oblivion-remastered-voxmeld_v0.3.2_P.pak"
@pushd %~1
(for /R %%f in (*) do @set "filePath=%%f" & set "relativePath=!filePath:%~1=!" & @echo "%%f" "../../../OblivionRemastered/!relativePath!")>%FILE_LIST%
@pushd "%UNREAL_BIN_DIR%"
::-compresslevel=4 for Normal, -compresslevel=-4 for uncompressed hyperfast paking
"%UNREAL_PAK_EXE%" %OUT_PAK% -create=%FILE_LIST% -compress -compressionformats=Oodle -compressmethod=Kraken -compresslevel=4
@popd
:skip