set count=0

if exist "%~1" (
    for /f %%A in ('dir /a-d /b "%~1" ^| find /v /c ""') do set count=%%A
)

set %~2=%count%