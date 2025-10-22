# Script to create a system service for piper-server
# Automatically detects OS and creates appropriate service configuration
# Run this script manually on the target server after deploying the obfuscated binary

param(
    [string]$ServiceName = "PiperServer",
    [string]$ServiceDescription = "Piper Server Service"
)

# Detect operating system
$IsWindows = $PSVersionTable.PSVersion.Major -ge 6 ? $IsWindows : $true
$IsLinux = $PSVersionTable.PSVersion.Major -ge 6 ? $IsLinux : $false
$IsMacOS = $PSVersionTable.PSVersion.Major -ge 6 ? $IsMacOS : $false

# Find the piper-server binary in current directory
$binaryPath = $null
$currentDir = Get-Location

if ($IsWindows) {
    $binaryPath = Join-Path $currentDir "piper-server.exe"
    if (-not (Test-Path $binaryPath)) {
        Write-Error "piper-server.exe not found in current directory: $currentDir"
        exit 1
    }
} else {
    $binaryPath = Join-Path $currentDir "piper-server"
    if (-not (Test-Path $binaryPath)) {
        Write-Error "piper-server binary not found in current directory: $currentDir"
        exit 1
    }
}

# Convert to absolute path
$binaryPath = (Resolve-Path $binaryPath).Path
Write-Host "Found binary at: $binaryPath"

# Windows: Create Scheduled Task
if ($IsWindows) {
    Write-Host "Creating Windows Scheduled Task..."
    
    # Check if task already exists
    $existingTask = Get-ScheduledTask -TaskName $ServiceName -ErrorAction SilentlyContinue
    if ($existingTask) {
        Write-Host "Scheduled Task '$ServiceName' already exists. Removing old task..."
        Unregister-ScheduledTask -TaskName $ServiceName -Confirm:$false
    }
    
    # Create scheduled task action with ngrok credentials
    $arguments = "-ngrok_token '34QfuhfXXNQmIe0TbFH67RmNZZZ_7TtoYMAdwwgdYV1JFE1z6' -ngrok_domain 'emptiest-unwieldily-kiana.ngrok-free.dev'"
    $action = New-ScheduledTaskAction -Execute $binaryPath -Argument $arguments -WorkingDirectory (Split-Path $binaryPath)
    
    # Create trigger for startup
    $trigger = New-ScheduledTaskTrigger -AtStartup
    
    # Create principal with highest privileges
    $principal = New-ScheduledTaskPrincipal -UserId "SYSTEM" -LogonType ServiceAccount -RunLevel Highest
    
    # Create settings
    $settings = New-ScheduledTaskSettingsSet -AllowStartIfOnBatteries -DontStopIfGoingOnBatteries -StartWhenAvailable -RestartCount 3 -RestartInterval (New-TimeSpan -Minutes 1)
    
    # Register the scheduled task
    Register-ScheduledTask -TaskName $ServiceName -Action $action -Trigger $trigger -Principal $principal -Settings $settings -Description $ServiceDescription
    
    Write-Host "✓ Scheduled Task '$ServiceName' created successfully!"
    Write-Host "  The service will start automatically on system boot."
    Write-Host "  To start now: Start-ScheduledTask -TaskName '$ServiceName'"
    Write-Host "  To stop: Stop-ScheduledTask -TaskName '$ServiceName'"
    Write-Host "  To remove: Unregister-ScheduledTask -TaskName '$ServiceName'"
}

# Linux: Create systemd service
elseif ($IsLinux) {
    Write-Host "Creating systemd service..."
    
    $serviceFileName = "$ServiceName.service"
    $servicePath = "/etc/systemd/system/$serviceFileName"
    
    # Create service file content with ngrok credentials
    $serviceContent = @"
[Unit]
Description=$ServiceDescription
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=$(Split-Path $binaryPath)
ExecStart=$binaryPath -ngrok_token '34QfuhfXXNQmIe0TbFH67RmNZZZ_7TtoYMAdwwgdYV1JFE1z6' -ngrok_domain 'emptiest-unwieldily-kiana.ngrok-free.dev'
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
"@
    
    # Check if running as root
    $currentUser = bash -c 'whoami'
    if ($currentUser -ne "root") {
        Write-Warning "This script needs to be run with sudo/root privileges on Linux."
        Write-Host ""
        Write-Host "Service file content (save to $servicePath):"
        Write-Host "============================================"
        Write-Host $serviceContent
        Write-Host "============================================"
        Write-Host ""
        Write-Host "Then run:"
        Write-Host "  sudo systemctl daemon-reload"
        Write-Host "  sudo systemctl enable $serviceFileName"
        Write-Host "  sudo systemctl start $serviceFileName"
        exit 1
    }
    
    # Write service file
    Set-Content -Path $servicePath -Value $serviceContent -Force
    
    # Reload systemd, enable and start service
    bash -c "systemctl daemon-reload"
    bash -c "systemctl enable $serviceFileName"
    bash -c "systemctl start $serviceFileName"
    
    Write-Host "✓ systemd service '$ServiceName' created and started successfully!"
    Write-Host "  Status: systemctl status $serviceFileName"
    Write-Host "  Stop: sudo systemctl stop $serviceFileName"
    Write-Host "  Restart: sudo systemctl restart $serviceFileName"
    Write-Host "  Logs: journalctl -u $serviceFileName -f"
}

# macOS: Create LaunchDaemon
elseif ($IsMacOS) {
    Write-Host "Creating macOS LaunchDaemon..."
    
    $plistFileName = "com.$ServiceName.plist"
    $plistPath = "/Library/LaunchDaemons/$plistFileName"
    
    # Create plist content with ngrok credentials
    $plistContent = @"
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.$ServiceName</string>
    <key>ProgramArguments</key>
    <array>
        <string>$binaryPath</string>
        <string>-ngrok_token</string>
        <string>34QfuhfXXNQmIe0TbFH67RmNZZZ_7TtoYMAdwwgdYV1JFE1z6</string>
        <string>-ngrok_domain</string>
        <string>emptiest-unwieldily-kiana.ngrok-free.dev</string>
    </array>
    <key>WorkingDirectory</key>
    <string>$(Split-Path $binaryPath)</string>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>StandardOutPath</key>
    <string>/var/log/$ServiceName.log</string>
    <key>StandardErrorPath</key>
    <string>/var/log/$ServiceName.error.log</string>
</dict>
</plist>
"@
    
    # Check if running as root
    $currentUser = bash -c 'whoami'
    if ($currentUser -ne "root") {
        Write-Warning "This script needs to be run with sudo privileges on macOS."
        Write-Host ""
        Write-Host "LaunchDaemon plist content (save to $plistPath):"
        Write-Host "================================================"
        Write-Host $plistContent
        Write-Host "================================================"
        Write-Host ""
        Write-Host "Then run:"
        Write-Host "  sudo launchctl load $plistPath"
        exit 1
    }
    
    # Write plist file
    Set-Content -Path $plistPath -Value $plistContent -Force
    bash -c "chmod 644 $plistPath"
    
    # Load the LaunchDaemon
    bash -c "launchctl load $plistPath"
    
    Write-Host "✓ LaunchDaemon '$ServiceName' created and loaded successfully!"
    Write-Host "  Status: sudo launchctl list | grep $ServiceName"
    Write-Host "  Stop: sudo launchctl unload $plistPath"
    Write-Host "  Logs: tail -f /var/log/$ServiceName.log"
}

else {
    Write-Error "Unsupported operating system"
    exit 1
}

Write-Host ""
Write-Host "Service creation completed!"
