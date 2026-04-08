Param(
  [Parameter(Mandatory = $true)]
  [string]$Name,

  [Parameter(Mandatory = $false)]
  [string]$Destination = (Join-Path (Get-Location) $Name),

  [Parameter(Mandatory = $false)]
  [string]$Module = ""
)

$ErrorActionPreference = "Stop"

$sdkRoot = Split-Path -Parent $PSScriptRoot
$templateRoot = Join-Path $sdkRoot "doc\template"

if (!(Test-Path $templateRoot)) {
  throw "No se encontró el template en: $templateRoot"
}

if (Test-Path $Destination) {
  throw "El destino ya existe: $Destination"
}

Write-Host "Copiando template a $Destination"
Copy-Item -Path $templateRoot -Destination $Destination -Recurse -Force

# Renombrar placeholders del módulo (opcional)
if ($Module -ne "") {
  Write-Host "Reemplazando 'your-module-name' por '$Module'"
  Get-ChildItem -Path $Destination -Recurse -File -Filter "*.go" | ForEach-Object {
    (Get-Content $_.FullName -Raw).Replace("your-module-name", $Module) | Set-Content -Path $_.FullName -NoNewline
  }
  $goMod = Join-Path $Destination "go.mod"
  if (Test-Path $goMod) {
    (Get-Content $goMod -Raw).Replace("module your-module-name", "module $Module") | Set-Content -Path $goMod -NoNewline
  }
}

# Crear driver.netsocs.json desde el ejemplo (sin credenciales)
$example = Join-Path $Destination "driver.netsocs.json.example"
$real = Join-Path $Destination "driver.netsocs.json"
if (Test-Path $example) {
  Copy-Item $example $real -Force
  Write-Host "Creado $real (recuerda completar credenciales)"
}

Write-Host "Listo. Siguientes pasos:"
Write-Host " - cd `"$Destination`""
Write-Host " - Edita driver.netsocs.json con tus credenciales"
Write-Host " - go mod tidy"
Write-Host " - go test ./..."
Write-Host " - go run ."

