# Create Windows Service for Pejelagarto Server
# Run this script as Administrator

param(
    [string]$ServiceName = "PejelagartoServer",
    [string]$DisplayName = "Pejelagarto Translation Server",
    [string]$Description = "Pejelagarto bidirectional translator server with TTS support"
)

# Get the current directory and project root
$ScriptDir = Split-Path -Parent $PSCommandPath
if (-not $ScriptDir) {
    $ScriptDir = $PWD.Path
}
$RootDir = Split-Path -Parent (Split-Path -Parent $ScriptDir)

# Check if running as Administrator
$IsAdmin = ([Security.Principal.WindowsPrincipal] [Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
if (-not $IsAdmin) {
    Write-Host "ERROR: This script must be run as Administrator" -ForegroundColor Red
    Write-Host "Right-click PowerShell and select 'Run as Administrator', then run this script again" -ForegroundColor Yellow
    exit 1
}

# Define the PowerShell command to run
$TaskCommand = "Set-Location '$RootDir' ; Get-Process -Name 'pejelagarto-server' -ErrorAction SilentlyContinue | Stop-Process -Force; .\scripts\helpers\build-prod-unobfuscated.ps1 ; Start-Process -FilePath '.\bin\pejelagarto-server.exe' -WorkingDirectory `$PWD -WindowStyle Hidden"

# Check if scheduled task already exists
$ExistingTask = Get-ScheduledTask -TaskName $ServiceName -ErrorAction SilentlyContinue
if ($ExistingTask) {
    Write-Host "Scheduled Task '$ServiceName' already exists. Stopping and removing..." -ForegroundColor Yellow
    Stop-ScheduledTask -TaskName $ServiceName -ErrorAction SilentlyContinue
    Unregister-ScheduledTask -TaskName $ServiceName -Confirm:$false -ErrorAction SilentlyContinue
    Start-Sleep -Seconds 2
}

# Create the scheduled task action
$Action = New-ScheduledTaskAction -Execute "powershell.exe" -Argument "-WindowStyle Hidden -ExecutionPolicy Bypass -Command `"$TaskCommand`""

# Create the scheduled task trigger (at startup)
$Trigger = New-ScheduledTaskTrigger -AtStartup

# Create the scheduled task principal (run as SYSTEM with highest privileges)
$Principal = New-ScheduledTaskPrincipal -UserId "SYSTEM" -LogonType ServiceAccount -RunLevel Highest

# Create the scheduled task settings
$Settings = New-ScheduledTaskSettingsSet -AllowStartIfOnBatteries -DontStopIfGoingOnBatteries -StartWhenAvailable -RestartCount 3 -RestartInterval (New-TimeSpan -Minutes 1)

# Register the scheduled task
try {
    Register-ScheduledTask -TaskName $ServiceName -Description $Description -Action $Action -Trigger $Trigger -Principal $Principal -Settings $Settings -Force
    Write-Host "Scheduled Task created successfully!" -ForegroundColor Green
} catch {
    Write-Host "ERROR: Failed to create scheduled task: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}

# Create logs directory
$LogsDir = Join-Path $RootDir "logs"
if (-not (Test-Path $LogsDir)) {
    New-Item -ItemType Directory -Path $LogsDir -Force | Out-Null
    Write-Host "Created logs directory: $LogsDir" -ForegroundColor Green
}

# Start the scheduled task
Write-Host "Starting scheduled task..." -ForegroundColor Green
Start-ScheduledTask -TaskName $ServiceName

# Check task status
Start-Sleep -Seconds 3
$Task = Get-ScheduledTask -TaskName $ServiceName
$TaskInfo = Get-ScheduledTaskInfo -TaskName $ServiceName

Write-Host "`nSCHEDULED TASK CREATED AND STARTED SUCCESSFULLY!" -ForegroundColor Green
Write-Host "Task Name: $ServiceName" -ForegroundColor Cyan
Write-Host "Status: $($Task.State)" -ForegroundColor Cyan
Write-Host "Last Run: $($TaskInfo.LastRunTime)" -ForegroundColor Cyan
Write-Host "Next Run: At system startup" -ForegroundColor Cyan

Write-Host "`nThe task will:" -ForegroundColor Yellow
Write-Host "  ✓ Run automatically when Windows boots" -ForegroundColor Green
Write-Host "  ✓ Stop any existing pejelagarto-server process" -ForegroundColor Green
Write-Host "  ✓ Build the latest unobfuscated production version" -ForegroundColor Green
Write-Host "  ✓ Start the server in hidden window mode" -ForegroundColor Green
Write-Host "  ✓ Restart automatically if it fails (up to 3 times)" -ForegroundColor Green
Write-Host "  ✓ Run even on battery power" -ForegroundColor Green

Write-Host "`nTask Management Commands:" -ForegroundColor Yellow
Write-Host "  Start:   Start-ScheduledTask -TaskName $ServiceName" -ForegroundColor Cyan
Write-Host "  Stop:    Stop-ScheduledTask -TaskName $ServiceName" -ForegroundColor Cyan
Write-Host "  Status:  Get-ScheduledTask -TaskName $ServiceName" -ForegroundColor Cyan
Write-Host "  Info:    Get-ScheduledTaskInfo -TaskName $ServiceName" -ForegroundColor Cyan
Write-Host "  Remove:  Unregister-ScheduledTask -TaskName $ServiceName -Confirm:`$false" -ForegroundColor Cyan

Write-Host "`nCommand being executed:" -ForegroundColor Yellow
Write-Host "  $TaskCommand" -ForegroundColor Gray
