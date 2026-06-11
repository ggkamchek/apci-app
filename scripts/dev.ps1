# Zapusk lokalnoj sredy APCI (Docker + Chat API + Security Center)
#
#   .\scripts\dev.ps1
#   .\scripts\dev.ps1 -Restart
#   .\scripts\dev.ps1 -Desktop

param(
    [switch]$Restart,
    [switch]$Desktop
)

$ErrorActionPreference = "Stop"
$Root = Split-Path -Parent $PSScriptRoot
Set-Location $Root

function Import-DotEnv {
    param([string]$Path = (Join-Path $Root ".env"))
    if (-not (Test-Path $Path)) {
        Write-Error ".env not found. Run: Copy-Item .env.example .env"
    }
    Get-Content $Path | ForEach-Object {
        if ($_ -match '^\s*([^#][^=]+)=(.*)$') {
            Set-Item -Path "env:$($matches[1].Trim())" -Value $matches[2].Trim()
        }
    }
}

function Stop-DevPorts {
    Write-Host "Stopping processes on ports 8080, 8081, 50051, 50052..."
    foreach ($port in 8080, 8081, 50051, 50052) {
        netstat -ano | findstr "LISTENING" | findstr ":$port " | ForEach-Object {
            $parts = $_ -split '\s+'
            $procId = $parts[-1]
            if ($procId -match '^\d+$') {
                cmd /c "taskkill /PID $procId /F >nul 2>&1"
            }
        }
    }
    Start-Sleep -Seconds 1
}

function Start-DevProcess {
    param(
        [string]$Title,
        [string]$WorkingDirectory,
        [string]$Command
    )
    $escapedRoot = $Root -replace "'", "''"
    $escapedWd = $WorkingDirectory -replace "'", "''"
    $escapedTitle = $Title -replace "'", "''"
    $script = @"
Set-Location '$escapedRoot'
Get-Content .env | ForEach-Object {
  if (`$_ -match '^\s*([^#][^=]+)=(.*)$') { Set-Item -Path "env:`$(`$matches[1].Trim())" -Value `$matches[2].Trim() }
}
Set-Location '$escapedWd'
`$Host.UI.RawUI.WindowTitle = '$escapedTitle'
Write-Host '=== $escapedTitle ===' -ForegroundColor Cyan
$Command
"@

    Start-Process powershell -ArgumentList "-NoExit", "-Command", $script | Out-Null
}

Write-Host ""
Write-Host "APCI dev start" -ForegroundColor Green
Write-Host "Root: $Root"
Write-Host ""

if (-not (Test-Path (Join-Path $Root ".env"))) {
    Write-Host "Creating .env from .env.example..."
    Copy-Item (Join-Path $Root ".env.example") (Join-Path $Root ".env")
}

if ($Restart) {
    Stop-DevPorts
}

Write-Host "Docker: postgres + postgres-security..."
$prevEap = $ErrorActionPreference
$ErrorActionPreference = "Continue"
$dockerCmd = Get-Command docker.exe -ErrorAction SilentlyContinue
if (-not $dockerCmd) {
    $dockerCmd = Get-Command docker -ErrorAction SilentlyContinue
}
if (-not $dockerCmd) {
    Write-Error "docker not found. Install Docker Desktop and ensure it is running."
}
& $dockerCmd.Source compose up -d postgres postgres-security
$dockerFailed = ($LASTEXITCODE -and $LASTEXITCODE -ne 0) -or (-not $?)
$ErrorActionPreference = $prevEap
if ($dockerFailed) {
    $code = if ($LASTEXITCODE) { $LASTEXITCODE } else { "unknown" }
    Write-Error "docker compose failed (exit code $code). Is Docker Desktop running?"
}

Import-DotEnv

Write-Host "Starting Chat API (new window)..."
Start-DevProcess -Title "APCI Chat API" -WorkingDirectory (Join-Path $Root "server") -Command "go run ./cmd/api"

Write-Host "Starting Security Center (new window)..."
Start-DevProcess -Title "APCI Security Center" -WorkingDirectory (Join-Path $Root "security-center") -Command "go run ./cmd/security"

if ($Desktop) {
    Write-Host "Starting Desktop (new window)..."
    Start-DevProcess -Title "APCI Desktop" -WorkingDirectory (Join-Path $Root "desktop") -Command "wails dev"
}

Write-Host ""
Write-Host "Done. Wait 5-10 sec for services to start." -ForegroundColor Yellow
Write-Host ""
Write-Host "  Chat API:    http://localhost:8080/health"
Write-Host "  Security:    http://localhost:8081/health"
Write-Host "  Admin UI:    http://localhost:8081/admin/"
Write-Host "  Admin token: dev-admin-token"
Write-Host ""
if (-not $Desktop) {
    Write-Host "  Desktop:     .\scripts\dev.ps1 -Desktop"
    Write-Host '  Or:          cd desktop  then  wails dev'
}
Write-Host ""
