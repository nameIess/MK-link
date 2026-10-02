$ErrorActionPreference = "Stop"

$projectRoot = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path
$source = Join-Path $projectRoot "ui\icon.svg"
$buildDir = Join-Path $projectRoot "build"
$destination = Join-Path $buildDir "appicon.png"

if (-not (Test-Path $source)) {
    throw "Missing application icon: $source. Add ui/icon.svg before building."
}

$magick = Get-Command magick -ErrorAction SilentlyContinue
if (-not $magick) {
    throw "ImageMagick (magick) is required to convert ui/icon.svg into build/appicon.png."
}

New-Item -ItemType Directory -Force -Path $buildDir | Out-Null

& $magick.Source -background none -density 300 $source -resize 1024x1024 $destination
if ($LASTEXITCODE -ne 0) {
    throw "ImageMagick failed to generate $destination."
}

if (-not (Test-Path $destination)) {
    throw "Image conversion completed without producing $destination."
}

Write-Host "Prepared Wails application icon: $destination"
