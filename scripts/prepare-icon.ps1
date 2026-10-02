$ErrorActionPreference = "Stop"

$projectRoot = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path
$source = Join-Path $projectRoot "ui\icon.svg"
$buildDir = Join-Path $projectRoot "build"
$png = Join-Path $buildDir "appicon.png"
$ico = Join-Path $buildDir "windows\icon.ico"

if (-not (Test-Path $source)) { throw "Missing application icon: $source" }

$magick = Get-Command magick -ErrorAction SilentlyContinue
if (-not $magick) { throw "ImageMagick (magick) is required to generate the application icon." }

New-Item -ItemType Directory -Force -Path $buildDir | Out-Null
New-Item -ItemType Directory -Force -Path (Split-Path $ico -Parent) | Out-Null

# ui/icon.svg is the single source asset for the application icon.
# Generate the PNG Wails embeds and the Windows ICO Wails uses.
& $magick.Source -background none -density 300 $source -resize 1024x1024 $png
if ($LASTEXITCODE -ne 0 -or -not (Test-Path $png)) {
    throw "Failed to generate $png from $source"
}

& $magick.Source $png -define icon:auto-resize=256,128,64,48,32,24,16 $ico
if ($LASTEXITCODE -ne 0 -or -not (Test-Path $ico)) {
    throw "Failed to generate $ico from $png"
}

# Fail early if ImageMagick produced an unexpectedly tiny file.
$size = (Get-Item $ico).Length
if ($size -lt 1024) {
    throw "Generated Windows icon is unexpectedly small: $size bytes"
}

Write-Host "Generated Windows application icon: $ico ($size bytes)"
