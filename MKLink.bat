@echo off
:: ===============================================
:: Script to Create a Symbolic Link with mklink
:: ===============================================

:: Request admin privileges
>nul 2>&1 "%SYSTEMROOT%\system32\cacls.exe" "%SYSTEMROOT%\system32\config\system"
if '%errorlevel%' NEQ '0' (
    echo Requesting administrative privileges...
    powershell -Command "Start-Process '%~f0' -Verb runAs"
    exit /b
)

:: Prompt the user for the new file name
echo ===============================================
echo Enter details to create the symbolic link
echo ===============================================
set /p newFileName=Enter the name for your link (e.g., DevProjects): 

:: Prompt the user for the folder/file path
set /p folderPath=Enter the target path (e.g., D:\Projects\Development): 

:: Remove quotes from the folderPath if present
set folderPath=%folderPath:"=%

:: Prompt for the link type (file/directory)
echo Select the type of symbolic link you want to create:
echo 1. Directory link (mklink /d) - most common for folders
echo 2. Hard link (mklink /h) - for files only
echo 3. Junction link (mklink /j) - for directories on the same drive
set /p linkType=Enter your choice (1, 2, or 3): 

:: Validate inputs
if "%newFileName%"=="" (
    echo Error: New file name cannot be empty.
    pause
    exit /b
)

if "%folderPath%"=="" (
    echo Error: Folder or file path cannot be empty.
    pause
    exit /b
)

:: Validate that the target folder/file exists
if not exist "%folderPath%" (
    echo Error: The specified path does not exist. Please check the path.
    pause
    exit /b
)

:: Validate the link type
if "%linkType%"=="1" (
    set linkOption=/d
) else if "%linkType%"=="2" (
    set linkOption=/h
) else if "%linkType%"=="3" (
    set linkOption=/j
) else (
    echo Error: Invalid link type selected.
    pause
    exit /b
)

:: Prompt for a custom location to create the symbolic link
set /p linkLocation=Enter the full location to create the symbolic link (e.g., C:\Users\ASUS\OneDrive\ or custom path): 

:: Validate the custom link location
if "%linkLocation%"=="" (
    echo Error: The link location cannot be empty.
    pause
    exit /b
)

:: Check if the link location already exists
if exist "%linkLocation%\%newFileName%" (
    echo Error: A file or folder already exists at this location. Choose a different name or location.
    pause
    exit /b
)

:: Create the symbolic link
echo Creating the symbolic link...
mklink %linkOption% "%linkLocation%\%newFileName%" "%folderPath%"
if '%errorlevel%' EQU '0' (
    echo Symbolic link created successfully!
) else (
    echo Failed to create the symbolic link. Please check your inputs and try again.
)

:: Pause to allow the user to see the result
pause
