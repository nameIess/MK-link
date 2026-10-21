$ErrorActionPreference = "Stop"

$projectRoot = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path
$buildDir = Join-Path $projectRoot "build"
$ico = Join-Path $buildDir "windows\icon.ico"
$png = Join-Path $buildDir "appicon.png"

if (-not (Test-Path $ico -PathType Leaf)) {
    throw "Missing Windows application icon: $ico"
}

New-Item -ItemType Directory -Force -Path $buildDir | Out-Null

# Wails also expects build/appicon.png. Derive it from the checked-in ICO
# using Windows/.NET APIs so normal builds do not require ImageMagick.
Add-Type -AssemblyName System.Drawing
$icon = $null
$bitmap = $null
try {
    $icon = New-Object System.Drawing.Icon($ico, 256, 256)
    $bitmap = $icon.ToBitmap()
    $bitmap.Save($png, [System.Drawing.Imaging.ImageFormat]::Png)
}
catch {
    throw "Could not create Wails appicon.png from $ico. Verify the ICO is valid. Details: $($_.Exception.Message)"
}
finally {
    if ($bitmap) { $bitmap.Dispose() }
    if ($icon) { $icon.Dispose() }
}

if (-not (Test-Path $png -PathType Leaf) -or (Get-Item $png).Length -lt 256) {
    throw "Failed to create a valid Wails application icon: $png"
}

Write-Host "Using existing Windows icon: $ico"
Write-Host "Generated Wails PNG icon: $png"
