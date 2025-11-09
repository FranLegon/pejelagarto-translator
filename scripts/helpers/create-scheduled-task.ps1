# Create Scheduled Task for Pejelagarto Server
# This keeps the server running even when computer is locked
# Run this script as Administrator

param(
    [string]$TaskName = "PejelagartoServer",
    [switch]$Remove
)

# Check if running as Administrator
$IsAdmin = ([Security.Principal.WindowsPrincipal] [Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
if (-not $IsAdmin) {
    Write-Host "ERROR: This script must be run as Administrator" -ForegroundColor Red
    exit 1
}

# Get script directory
$ScriptDir = if ($PSScriptRoot) { 
    $PSScriptRoot 
} else { 
    Split-Path -Parent $MyInvocation.MyCommand.Path 
}

# Navigate to root directory (two levels up from scripts/helpers/)
$RootDir = Split-Path -Parent (Split-Path -Parent $ScriptDir)
$ExePath = Join-Path $RootDir "bin\pejelagarto-server.exe"

if ($Remove) {
    Write-Host "Removing scheduled task '$TaskName'..." -ForegroundColor Yellow
    Unregister-ScheduledTask -TaskName $TaskName -Confirm:$false -ErrorAction SilentlyContinue
    Write-Host "Task removed successfully!" -ForegroundColor Green
    exit 0
}

# Verify executable exists
if (-not (Test-Path $ExePath)) {
    Write-Host "ERROR: Executable not found at: $ExePath" -ForegroundColor Red
    exit 1
}

# Remove existing task if it exists
Unregister-ScheduledTask -TaskName $TaskName -Confirm:$false -ErrorAction SilentlyContinue

# Create scheduled task action
$Action = New-ScheduledTaskAction -Execute $ExePath -WorkingDirectory $RootDir

# Create trigger (at logon, for any user)
$Trigger = New-ScheduledTaskTrigger -AtLogOn

# Create settings
$Settings = New-ScheduledTaskSettingsSet `
    -AllowStartIfOnBatteries `
    -DontStopIfGoingOnBatteries `
    -StartWhenAvailable `
    -DontStopOnIdleEnd `
    -ExecutionTimeLimit (New-TimeSpan -Days 0) `
    -RestartCount 3 `
    -RestartInterval (New-TimeSpan -Minutes 1)

# Create principal (run whether user is logged on or not, with highest privileges)
$Principal = New-ScheduledTaskPrincipal -UserId "SYSTEM" -LogonType ServiceAccount -RunLevel Highest

# Register the task
Register-ScheduledTask -TaskName $TaskName `
    -Action $Action `
    -Trigger $Trigger `
    -Settings $Settings `
    -Principal $Principal `
    -Description "Pejelagarto Translation Server - Runs in background even when locked" `
    -Force

Write-Host "`nSCHEDULED TASK CREATED SUCCESSFULLY!" -ForegroundColor Green
Write-Host "Task Name: $TaskName" -ForegroundColor Cyan
Write-Host "`nThe server will now:" -ForegroundColor Yellow
Write-Host "  ✓ Start automatically at logon" -ForegroundColor Green
Write-Host "  ✓ Keep running when you lock the computer" -ForegroundColor Green
Write-Host "  ✓ Restart automatically if it crashes (up to 3 times)" -ForegroundColor Green
Write-Host "  ✓ Run in the background (no visible window)" -ForegroundColor Green
Write-Host "  ✓ Continue running on battery power" -ForegroundColor Green

Write-Host "`nStarting the task now..." -ForegroundColor Yellow
Start-ScheduledTask -TaskName $TaskName
Start-Sleep -Seconds 2

$Task = Get-ScheduledTask -TaskName $TaskName
Write-Host "Task Status: $($Task.State)" -ForegroundColor Cyan

Write-Host "`nTask Management Commands:" -ForegroundColor Yellow
Write-Host "  Start:   Start-ScheduledTask -TaskName $TaskName" -ForegroundColor Cyan
Write-Host "  Stop:    Stop-ScheduledTask -TaskName $TaskName" -ForegroundColor Cyan
Write-Host "  Status:  Get-ScheduledTask -TaskName $TaskName" -ForegroundColor Cyan
Write-Host "  Remove:  .\scripts\helpers\create-scheduled-task.ps1 -Remove" -ForegroundColor Cyan
