goto should_install

:install

if exist "%RESULT_FOLDER%\" (
    echo INFO: Instaling Mod...

    if not exist "!DIRECTORY_BACKUP!\%VOICES_1%.bsa" (
        echo INFO: Creating backup for original .bsa files...

        move "!DIRECTORY_OBRE!\Dev\ObvData\Data\%VOICES_1%.bsa" "!DIRECTORY_BACKUP!\%VOICES_1%.bsa" >nul
        move "!DIRECTORY_OBRE!\Dev\ObvData\Data\%VOICES_2%.bsa" "!DIRECTORY_BACKUP!\%VOICES_2%.bsa" >nul
        move "!DIRECTORY_OBRE!\Dev\ObvData\Data\%SHIVERING_ISLES%.bsa" "!DIRECTORY_BACKUP!\%SHIVERING_ISLES%.bsa" >nul
        move "!DIRECTORY_OBRE!\Dev\ObvData\Data\%KNIGHTS%.bsa" "!DIRECTORY_BACKUP!\%KNIGHTS%.bsa" >nul
        move "!DIRECTORY_OBRE!\Dev\ObvData\Data\%DLC_1%.bsa" "!DIRECTORY_BACKUP!\%DLC_1%.bsa" >nul
        move "!DIRECTORY_OBRE!\Dev\ObvData\Data\%DLC_2%.bsa" "!DIRECTORY_BACKUP!\%DLC_2%.bsa" >nul
        move "!DIRECTORY_OBRE!\Dev\ObvData\Data\%DLC_3%.bsa" "!DIRECTORY_BACKUP!\%DLC_3%.bsa" >nul
        move "!DIRECTORY_OBRE!\Dev\ObvData\Data\%DLC_4%.bsa" "!DIRECTORY_BACKUP!\%DLC_4%.bsa" >nul

        echo Success. Backup has been saved in !DIRECTORY_BACKUP!
    ) else (
        if exist "!DIRECTORY_OBRE!\Dev\ObvData\Data\%VOICES_1%.bsa" (
            echo INFO: Deleting current .bsa files...

            del "!DIRECTORY_OBRE!\Dev\ObvData\Data\%VOICES_1%.bsa" >nul
            del "!DIRECTORY_OBRE!\Dev\ObvData\Data\%VOICES_2%.bsa" >nul
            del "!DIRECTORY_OBRE!\Dev\ObvData\Data\%SHIVERING_ISLES%.bsa" >nul
            del "!DIRECTORY_OBRE!\Dev\ObvData\Data\%KNIGHTS%.bsa" >nul
            del "!DIRECTORY_OBRE!\Dev\ObvData\Data\%DLC_1%.bsa" >nul
            del "!DIRECTORY_OBRE!\Dev\ObvData\Data\%DLC_2%.bsa" >nul
            del "!DIRECTORY_OBRE!\Dev\ObvData\Data\%DLC_3%.bsa" >nul
            del "!DIRECTORY_OBRE!\Dev\ObvData\Data\%DLC_4%.bsa" >nul
        ) else (
            echo ERROR: Could not find .bsa files
            pause
            exit
        )
    )

    :: Remove outdated Voice Fix if present
    if exist "!DIRECTORY_OBRE!\Dev\ObvData\Data\sound" (
		echo INFO: Found sound folder. This is probably an old version of the voice fix part of this mod
		call :should_delete

		if %errorlevel% == 0 (
			echo Removing old version of voice fix...
			rd /s /q "!DIRECTORY_OBRE!\Dev\ObvData\Data\sound"
		)
    )

    :: Remove previous versions of the .pak file
    for %%f in ("!DIRECTORY_OBRE!\Paks\~mods\*german-voices-oblivion-remastered-voxmeld_*") do (
        echo INFO: Removing old version of the mod...
        del "%%f"
    )

    :: Move files to game directory
    robocopy "%RESULT_FOLDER%" "!DIRECTORY_OBRE!" * /MOVE /E /NFL /NDL /NJH /NJS /NP

    if exist "!DIRECTORY_OBRE!\Dev\ObvData\Data\%VOICES_1%.bsa" (
        echo Successfully installed mod!
    ) else (
        echo ERROR: Installation failed!
        pause
        exit
    )
) else (
    echo ERROR: Could not find mod files
    pause
    exit
)

exit /b

:should_install

:ask_install
set /p shouldInstall=Do you want us to install the mod for you? [y/n]

echo !shouldInstall!

if "!shouldInstall!"=="y" (
    goto install
) else (
	if "!shouldInstall!"=="n" (
        echo Please copy the whole 'Content' folder from the 'Mod' folder into your game directory!
		exit /b
	) else (
		goto ask_install
	)
)

exit /b

:should_delete

:ask_delete
set /p shouldDelete=Do you want to delete it? [y/n]

echo !shouldDelete!

if "!shouldDelete!"=="y" (
    exit /b 0
) else (
	if "!shouldDelete!"=="n" (
		exit /b 1
	) else (
		goto ask_delete
	)
)

exit /b