$ErrorActionPreference = "Stop"

$projectRoot = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path
$source = Join-Path $projectRoot "ui\icon.svg"
$buildDir = Join-Path $projectRoot "build"
$png = Join-Path $buildDir "appicon.png"
$ico = Join-Path $buildDir "windows\icon.ico"

if (-not (Test-Path $source)) {
    throw "Missing application icon: $source"
}

$magick = Get-Command magick -ErrorAction SilentlyContinue
if (-not $magick) {
    throw "ImageMagick (magick) is required to generate the application icon."
}

New-Item -ItemType Directory -Force -Path $buildDir | Out-Null
New-Item -ItemType Directory -Force -Path (Split-Path $ico -Parent) | Out-Null

# Generate the canonical Wails source image from the checked-in SVG.
& $magick.Source -background none -density 300 $source -resize 1024x1024 $png
if ($LASTEXITCODE -ne 0 -or -not (Test-Path $png)) {
    throw "Failed to generate $png from $source"
}

# Generate the Windows ICO explicitly and keep it in the repository build tree.
# Wails will use this existing icon.ico instead of generating its default icon.
& $magick.Source $png -define icon:auto-resize=256,128,64,48,32,16 $ico
if ($LASTEXITCODE -ne 0 -or -not (Test-Path $ico)) {
    throw "Failed to generate $ico from $png"
}

Write-Host "Prepared application icon:"
Write-Host "  SVG: $source"
Write-Host "  PNG: $png"
Write-Host "  ICO: $ico"
