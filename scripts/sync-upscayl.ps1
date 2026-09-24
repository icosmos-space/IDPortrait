$ErrorActionPreference = "Stop"
$root = Split-Path (Split-Path $PSScriptRoot -Parent) -Parent
# Repo layout: IDPortrait/scripts or IDPortrait/IDPortrait/...
# Prefer locating from this script path.
$repo = Resolve-Path (Join-Path $PSScriptRoot "..")
if (Test-Path (Join-Path $repo "IDPortrait\core\runtime")) {
  $idRoot = Join-Path $repo "IDPortrait"
} elseif (Test-Path (Join-Path $repo "core\runtime")) {
  $idRoot = $repo
} else {
  throw "Cannot find IDPortrait core/runtime from $PSScriptRoot"
}

$src = "D:\gospace\src\github.com\icosmos-space\Upscaler\upscaler\assets"
if (-not (Test-Path $src)) {
  throw "Upscaler assets not found: $src"
}

$dst = Join-Path $idRoot "core\runtime\upscayl"
New-Item -ItemType Directory -Force -Path "$dst\bin", "$dst\models" | Out-Null
Copy-Item "$src\bin\upscayl-bin.exe" "$dst\bin\" -Force
Copy-Item "$src\bin\vcomp140.dll" "$dst\bin\" -Force
Copy-Item "$src\models\upscayl-standard-4x.bin" "$dst\models\" -Force
Copy-Item "$src\models\upscayl-standard-4x.param" "$dst\models\" -Force
Write-Host "Synced Upscayl assets to $dst"
Get-ChildItem -Recurse $dst | Select-Object FullName, Length
