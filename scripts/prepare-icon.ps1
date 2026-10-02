$ErrorActionPreference = "Stop"

$projectRoot = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path
$source = Join-Path $projectRoot "ui\\icon.svg"
$buildDir = Join-Path $projectRoot "build"
$png = Join-Path $buildDir "appicon.png"
$ico = Join-Path $buildDir "windows\\icon.ico"

if (-not (Test-Path $source)) { throw "Missing application icon: $source" }

$magick = Get-Command magick -ErrorAction SilentlyContinue
if (-not $magick) { throw "ImageMagick (magick) is required to generate the application icon." }

New-Item -ItemType Directory -Force -Path $buildDir | Out-Null
New-Item -ItemType Directory -Force -Path (Split-Path $ico -Parent) | Out-Null

# ui/icon.svg is the only source asset. Generate the Wails PNG from it.
& $magick.Source -background none -density 300 $source -resize 1024x1024 $png
if ($LASTEXITCODE -ne 0 -or -not (Test-Path $png)) { throw "Failed to generate $png from $source" }

# The Windows ICO is checked in at build/windows/icon.ico so Wails cannot fall back to its default icon.
if (-not (Test-Path $ico)) {
    throw "Missing checked-in Windows icon: $ico"
}
Write-Host "Using checked-in Windows application icon: $ico"
