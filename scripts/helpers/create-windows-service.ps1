# Create Windows Service for Pejelagarto Server
# Run this script as Administrator

param(
    [string]$ServiceName = "PejelagartoServer",
    [string]$DisplayName = "Pejelagarto Translation Server",
    [string]$Description = "Pejelagarto bidirectional translator server with TTS support"
)

# Get the current directory and executable path
$ScriptDir = Split-Path -Parent $PSCommandLineArgs[0]
if (-not $ScriptDir) {
    $ScriptDir = $PWD.Path
}
$RootDir = Split-Path -Parent (Split-Path -Parent $ScriptDir)
$ExePath = Join-Path $RootDir "bin\pejelagarto-server.exe"

# Verify executable exists
if (-not (Test-Path $ExePath)) {
    Write-Host "ERROR: Executable not found at: $ExePath" -ForegroundColor Red
    Write-Host "Please build the server first using: .\scripts\helpers\build-prod-unobfuscated.ps1" -ForegroundColor Yellow
    exit 1
}

# Check if running as Administrator
$IsAdmin = ([Security.Principal.WindowsPrincipal] [Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
if (-not $IsAdmin) {
    Write-Host "ERROR: This script must be run as Administrator" -ForegroundColor Red
    Write-Host "Right-click PowerShell and select 'Run as Administrator', then run this script again" -ForegroundColor Yellow
    exit 1
}

# Check if service already exists
$ExistingService = Get-Service -Name $ServiceName -ErrorAction SilentlyContinue
if ($ExistingService) {
    Write-Host "Service '$ServiceName' already exists. Stopping and removing..." -ForegroundColor Yellow
    Stop-Service -Name $ServiceName -Force -ErrorAction SilentlyContinue
    Start-Sleep -Seconds 2
    
    # Remove using sc.exe (more reliable than Remove-Service for some versions)
    sc.exe delete $ServiceName
    Start-Sleep -Seconds 2
}

# Create the service using NSSM (Non-Sucking Service Manager) if available
$NssmPath = Get-Command nssm.exe -ErrorAction SilentlyContinue
if ($NssmPath) {
    Write-Host "Using NSSM to create service..." -ForegroundColor Green
    
    nssm install $ServiceName "$ExePath"
    nssm set $ServiceName AppDirectory "$RootDir"
    nssm set $ServiceName DisplayName "$DisplayName"
    nssm set $ServiceName Description "$Description"
    nssm set $ServiceName Start SERVICE_AUTO_START
    nssm set $ServiceName AppStdout "$RootDir\logs\service-output.log"
    nssm set $ServiceName AppStderr "$RootDir\logs\service-error.log"
    nssm set $ServiceName AppRotateFiles 1
    nssm set $ServiceName AppRotateBytes 1048576
    
    Write-Host "Service created successfully using NSSM!" -ForegroundColor Green
} else {
    # Fallback: Use sc.exe (Windows built-in)
    Write-Host "NSSM not found, using sc.exe..." -ForegroundColor Yellow
    Write-Host "Note: For better service management, consider installing NSSM from https://nssm.cc/" -ForegroundColor Cyan
    
    sc.exe create $ServiceName binPath= "`"$ExePath`"" start= auto DisplayName= "$DisplayName"
    sc.exe description $ServiceName "$Description"
    sc.exe failure $ServiceName reset= 86400 actions= restart/60000/restart/60000/restart/60000
    
    Write-Host "Service created successfully using sc.exe!" -ForegroundColor Green
}

# Create logs directory
$LogsDir = Join-Path $RootDir "logs"
if (-not (Test-Path $LogsDir)) {
    New-Item -ItemType Directory -Path $LogsDir -Force | Out-Null
    Write-Host "Created logs directory: $LogsDir" -ForegroundColor Green
}

# Start the service
Write-Host "Starting service..." -ForegroundColor Green
Start-Service -Name $ServiceName

# Check service status
Start-Sleep -Seconds 3
$Service = Get-Service -Name $ServiceName
if ($Service.Status -eq 'Running') {
    Write-Host "`nSERVICE CREATED AND STARTED SUCCESSFULLY!" -ForegroundColor Green
    Write-Host "Service Name: $ServiceName" -ForegroundColor Cyan
    Write-Host "Status: $($Service.Status)" -ForegroundColor Cyan
    Write-Host "Startup Type: Automatic" -ForegroundColor Cyan
    Write-Host "`nThe server will now:" -ForegroundColor Yellow
    Write-Host "  ✓ Start automatically when Windows boots" -ForegroundColor Green
    Write-Host "  ✓ Keep running when you lock the computer" -ForegroundColor Green
    Write-Host "  ✓ Restart automatically if it crashes" -ForegroundColor Green
    Write-Host "  ✓ Run in the background (no visible window)" -ForegroundColor Green
    
    Write-Host "`nService Management Commands:" -ForegroundColor Yellow
    Write-Host "  Stop:    Stop-Service -Name $ServiceName" -ForegroundColor Cyan
    Write-Host "  Start:   Start-Service -Name $ServiceName" -ForegroundColor Cyan
    Write-Host "  Restart: Restart-Service -Name $ServiceName" -ForegroundColor Cyan
    Write-Host "  Status:  Get-Service -Name $ServiceName" -ForegroundColor Cyan
    Write-Host "  Remove:  sc.exe delete $ServiceName" -ForegroundColor Cyan
} else {
    Write-Host "`nWARNING: Service created but not running!" -ForegroundColor Yellow
    Write-Host "Status: $($Service.Status)" -ForegroundColor Red
    Write-Host "Check Event Viewer for error details" -ForegroundColor Yellow
}
