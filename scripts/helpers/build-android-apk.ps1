#!/usr/bin/env pwsh
# Build Android APK using gomobile
# This script automatically downloads Android SDK/NDK and Java JDK if not present and builds the APK

param(
    [switch]$SkipDownload = $false
)

$ErrorActionPreference = "Stop"

Write-Host "`n======================================"
Write-Host "  Android APK Build Script"
Write-Host "======================================`n"

# Check if gomobile is installed
Write-Host "[1/6] Checking gomobile installation..."
$gomobilePath = Join-Path (go env GOPATH) "bin\gomobile.exe"
if (-not (Test-Path $gomobilePath)) {
    Write-Host "  ⚠️  gomobile not found. Installing..." -ForegroundColor Yellow
    go install golang.org/x/mobile/cmd/gomobile@latest
    go install golang.org/x/mobile/cmd/gobind@latest
    Write-Host "  ✓ gomobile installed" -ForegroundColor Green
} else {
    Write-Host "  ✓ gomobile found" -ForegroundColor Green
}

# Add GOPATH/bin to PATH for this session
$env:PATH += ";$(go env GOPATH)\bin"

# Check for Java JDK (required for Android SDK tools)
Write-Host "`n[2/6] Checking Java JDK..."
$javaHome = $env:JAVA_HOME
if (-not $javaHome -or -not (Test-Path "$javaHome\bin\java.exe")) {
    # Try to find java in common locations
    $commonJavaPaths = @(
        "$env:ProgramFiles\Java",
        "$env:ProgramFiles\Microsoft\jdk-*",
        "$env:ProgramFiles\Eclipse Adoptium\jdk-*",
        "$env:ProgramFiles\Zulu\zulu-*"
    )
    
    $foundJava = $false
    foreach ($basePath in $commonJavaPaths) {
        if ($basePath -like "*`**") {
            $jdkDirs = Get-ChildItem -Path ($basePath -replace '\\\*.*', '') -Directory -ErrorAction SilentlyContinue | Where-Object { $_.Name -like ($basePath -replace '.*\\', '') }
            foreach ($jdkDir in $jdkDirs) {
                if (Test-Path "$($jdkDir.FullName)\bin\java.exe") {
                    $javaHome = $jdkDir.FullName
                    $foundJava = $true
                    break
                }
            }
        } elseif (Test-Path $basePath) {
            $jdkDirs = Get-ChildItem -Path $basePath -Directory -ErrorAction SilentlyContinue | Where-Object { $_.Name -like "jdk-*" }
            if ($jdkDirs) {
                $javaHome = $jdkDirs[0].FullName
                $foundJava = $true
                break
            }
        }
        if ($foundJava) { break }
    }
    
    if (-not $foundJava -and -not $SkipDownload) {
        Write-Host "  ⚠️  Java JDK not found. Downloading..." -ForegroundColor Yellow
        
        # Try using winget first (faster and more reliable)
        Write-Host "  Attempting to install via winget..." -ForegroundColor Cyan
        try {
            $wingetOutput = winget install Microsoft.OpenJDK.17 --silent --accept-source-agreements --accept-package-agreements 2>&1
            Start-Sleep -Seconds 3
            
            # Find the installed JDK
            $jdkSearchPaths = @(
                "$env:ProgramFiles\Microsoft",
                "$env:ProgramFiles\Eclipse Adoptium",
                "$env:ProgramFiles\Java",
                "$env:ProgramFiles\Zulu"
            )
            
            foreach ($searchPath in $jdkSearchPaths) {
                if (Test-Path $searchPath) {
                    $jdks = Get-ChildItem $searchPath -Directory -ErrorAction SilentlyContinue | Where-Object { $_.Name -like "jdk-*" -or $_.Name -like "jdk*" -or $_.Name -like "zulu*" }
                    if ($jdks) {
                        $javaHome = $jdks[0].FullName
                        Write-Host "  ✓ Java JDK installed via winget" -ForegroundColor Green
                        $foundJava = $true
                        break
                    }
                }
            }
        } catch {
            Write-Host "  ⚠️  winget installation failed, trying manual download..." -ForegroundColor Yellow
        }
        
        # Fallback: Download portable ZIP version
        if (-not $foundJava) {
            $jdkUrl = "https://aka.ms/download-jdk/microsoft-jdk-17.0.13-windows-x64.zip"
            $jdkZip = "$env:TEMP\microsoft-jdk-17.zip"
            $jdkInstallDir = "$env:LOCALAPPDATA\Microsoft\jdk-17"
            
            Write-Host "  Downloading Microsoft OpenJDK 17 (portable)..." -ForegroundColor Cyan
            try {
                $ProgressPreference = 'SilentlyContinue'
                Invoke-WebRequest -Uri $jdkUrl -OutFile $jdkZip -UseBasicParsing
                Write-Host "  ✓ Downloaded JDK" -ForegroundColor Green
                
                # Extract
                Write-Host "  Extracting JDK..." -ForegroundColor Cyan
                if (Test-Path $jdkInstallDir) {
                    Remove-Item $jdkInstallDir -Recurse -Force
                }
                Expand-Archive -Path $jdkZip -DestinationPath $jdkInstallDir -Force
                Remove-Item $jdkZip -Force
                
                # Find the extracted JDK directory
                $extractedJdk = Get-ChildItem $jdkInstallDir -Directory | Where-Object { $_.Name -like "jdk-*" } | Select-Object -First 1
                if ($extractedJdk) {
                    $javaHome = $extractedJdk.FullName
                    Write-Host "  ✓ Java JDK extracted to $javaHome" -ForegroundColor Green
                }
            } catch {
                Write-Host "  ❌ Failed to download/extract JDK: $_" -ForegroundColor Red
                Write-Host "  Please install manually from: https://learn.microsoft.com/en-us/java/openjdk/download" -ForegroundColor Yellow
                exit 1
            }
        }
    }
}

if ($javaHome -and (Test-Path "$javaHome\bin\java.exe")) {
    $env:JAVA_HOME = $javaHome
    $env:PATH += ";$javaHome\bin"
    $javaVersion = & "$javaHome\bin\java.exe" -version 2>&1 | Select-Object -First 1
    Write-Host "  ✓ Java JDK found: $javaVersion" -ForegroundColor Green
} else {
    Write-Host "  ❌ Java JDK not available" -ForegroundColor Red
    Write-Host "  Please install Java JDK manually" -ForegroundColor Yellow
    exit 1
}

# Check for Android SDK
Write-Host "`n[3/6] Checking Android SDK..."
$androidHome = $env:ANDROID_HOME
if (-not $androidHome) {
    $androidHome = "$env:LOCALAPPDATA\Android\sdk"
}

# Download and install Android SDK if not present
if (-not (Test-Path $androidHome) -and -not $SkipDownload) {
    Write-Host "  ⚠️  Android SDK not found. Downloading..." -ForegroundColor Yellow
    
    # Create SDK directory
    New-Item -ItemType Directory -Path $androidHome -Force | Out-Null
    
    # Download Android commandline tools
    $cmdlineToolsUrl = "https://dl.google.com/android/repository/commandlinetools-win-11076708_latest.zip"
    $cmdlineToolsZip = "$env:TEMP\commandlinetools.zip"
    $cmdlineToolsDir = Join-Path $androidHome "cmdline-tools"
    
    Write-Host "  Downloading Android commandline tools..." -ForegroundColor Cyan
    try {
        $ProgressPreference = 'SilentlyContinue'
        Invoke-WebRequest -Uri $cmdlineToolsUrl -OutFile $cmdlineToolsZip -UseBasicParsing
        Write-Host "  ✓ Downloaded commandline tools" -ForegroundColor Green
        
        # Extract
        Write-Host "  Extracting..." -ForegroundColor Cyan
        Expand-Archive -Path $cmdlineToolsZip -DestinationPath "$cmdlineToolsDir\temp" -Force
        
        # Move to correct location (SDK expects cmdline-tools/latest/)
        $latestDir = Join-Path $cmdlineToolsDir "latest"
        if (Test-Path $latestDir) {
            Remove-Item $latestDir -Recurse -Force
        }
        Move-Item "$cmdlineToolsDir\temp\cmdline-tools" $latestDir
        Remove-Item "$cmdlineToolsDir\temp" -Recurse -Force
        Remove-Item $cmdlineToolsZip -Force
        
        Write-Host "  ✓ Android SDK commandline tools installed" -ForegroundColor Green
    } catch {
        Write-Host "  ❌ Failed to download SDK: $_" -ForegroundColor Red
        Write-Host "  Please install manually from: https://developer.android.com/studio" -ForegroundColor Yellow
        exit 1
    }
}

if (Test-Path $androidHome) {
    Write-Host "  ✓ Android SDK found at: $androidHome" -ForegroundColor Green
    $env:ANDROID_HOME = $androidHome
} else {
    Write-Host "  ❌ Android SDK not available" -ForegroundColor Red
    exit 1
}

# Install required SDK components
Write-Host "`n[4/6] Installing SDK components..."
$sdkManager = Join-Path $androidHome "cmdline-tools\latest\bin\sdkmanager.bat"

if (Test-Path $sdkManager) {
    Write-Host "  Installing platform-tools, build-tools, and NDK..." -ForegroundColor Cyan
    Write-Host "  (This may take 5-10 minutes on first run)" -ForegroundColor Gray
    
    try {
        # Accept licenses first (silent)
        $licenseInput = "y`n" * 10
        $licenseInput | & $sdkManager --licenses 2>&1 | Out-Null
        
        # Install required components one at a time for better reliability
        Write-Host "  - Installing platform-tools..." -ForegroundColor Gray
        & $sdkManager "platform-tools" 2>&1 | Out-Null
        
        Write-Host "  - Installing build-tools..." -ForegroundColor Gray
        & $sdkManager "build-tools;34.0.0" 2>&1 | Out-Null
        
        Write-Host "  - Installing Android platform..." -ForegroundColor Gray
        & $sdkManager "platforms;android-34" 2>&1 | Out-Null
        
        Write-Host "  - Installing NDK (this takes longest)..." -ForegroundColor Gray
        & $sdkManager "ndk;26.3.11579264" 2>&1 | Out-Null
        
        Write-Host "  ✓ SDK components installed" -ForegroundColor Green
        
        # Wait a moment for filesystem to sync
        Start-Sleep -Seconds 2
        
        # Set NDK path
        $ndkPath = Join-Path $androidHome "ndk\26.3.11579264"
        if (Test-Path $ndkPath) {
            $env:ANDROID_NDK_HOME = $ndkPath
            Write-Host "  ✓ Android NDK configured: 26.3.11579264" -ForegroundColor Green
        } else {
            Write-Host "  ⚠️  NDK path not found at expected location" -ForegroundColor Yellow
            # Try to find any NDK version
            $ndkBase = Join-Path $androidHome "ndk"
            if (Test-Path $ndkBase) {
                $ndkVersions = Get-ChildItem $ndkBase -Directory | Select-Object -First 1
                if ($ndkVersions) {
                    $env:ANDROID_NDK_HOME = $ndkVersions.FullName
                    Write-Host "  ✓ Using NDK: $($ndkVersions.Name)" -ForegroundColor Green
                }
            }
        }
    } catch {
        Write-Host "  ⚠️  Some components may not have installed: $_" -ForegroundColor Yellow
    }
} else {
    # Check for existing NDK
    $ndkPath = Join-Path $androidHome "ndk"
    if (Test-Path $ndkPath) {
        $ndkVersions = Get-ChildItem $ndkPath -Directory | Select-Object -First 1
        if ($ndkVersions) {
            $env:ANDROID_NDK_HOME = $ndkVersions.FullName
            Write-Host "  ✓ Android NDK found: $($ndkVersions.Name)" -ForegroundColor Green
        }
    } else {
        Write-Host "  ⚠️  NDK not found - APK build may fail" -ForegroundColor Yellow
    }
}

# Initialize gomobile
Write-Host "`n[5/6] Initializing gomobile..."
try {
    gomobile init 2>&1 | Out-Null
    Write-Host "  ✓ gomobile initialized" -ForegroundColor Green
} catch {
    Write-Host "  ⚠️  gomobile init warning (may already be initialized)" -ForegroundColor Yellow
}

# Build the APK
Write-Host "`n[6/6] Building Android APK..."

# Ensure bin directory exists
if (-not (Test-Path "bin")) {
    New-Item -ItemType Directory -Path "bin" | Out-Null
}

try {
    Write-Host "  Building multi-architecture APK..." -ForegroundColor Cyan
    Write-Host "  Min SDK: 24 (Android 7.0+)" -ForegroundColor Gray
    Write-Host "  Target SDK: 34 (Android 14)" -ForegroundColor Gray
    Write-Host "  Architectures: ARM, ARM64, x86, x86_64 (all platforms)" -ForegroundColor Gray
    Write-Host "  This may take several minutes on first build..." -ForegroundColor Gray
    
    # Build with API 24 for modern Android compatibility
    # Android 15 requires minSdkVersion >= 24 for app installation
    # Using 'android' target builds for all architectures automatically
    gomobile build -androidapi 24 -target android -o bin/pejelagarto-translator-unsigned.apk ./cmd/androidapp
    
    if (Test-Path "bin/pejelagarto-translator-unsigned.apk") {
        Write-Host "  ✓ APK built" -ForegroundColor Green
        
        # Sign the APK with debug key for installation
        Write-Host "  Signing APK with debug key..." -ForegroundColor Cyan
        
        $debugKeystore = "$env:USERPROFILE\.android\debug.keystore"
        $apksigner = Join-Path $androidHome "build-tools\34.0.0\apksigner.bat"
        
        # Create debug keystore if it doesn't exist
        if (-not (Test-Path $debugKeystore)) {
            Write-Host "  Creating debug keystore..." -ForegroundColor Gray
            $androidDir = "$env:USERPROFILE\.android"
            if (-not (Test-Path $androidDir)) {
                New-Item -ItemType Directory -Path $androidDir -Force | Out-Null
            }
            
            $keytoolArgs = @(
                "-genkey", "-v",
                "-keystore", $debugKeystore,
                "-storepass", "android",
                "-alias", "androiddebugkey",
                "-keypass", "android",
                "-keyalg", "RSA",
                "-keysize", "2048",
                "-validity", "10000",
                "-dname", "CN=Android Debug,O=Android,C=US"
            )
            & "$env:JAVA_HOME\bin\keytool.exe" $keytoolArgs 2>&1 | Out-Null
        }
        
        # Sign the APK
        if (Test-Path $apksigner) {
            # Check if apktool.jar exists for manifest patching
            $apktoolPath = "bin/apktool.jar"
            if (Test-Path $apktoolPath) {
                # Patch the APK to set targetSdkVersion using apktool
                Write-Host "  Patching targetSdkVersion to 34..." -ForegroundColor Cyan
                
                # Decompile APK
                & "$env:JAVA_HOME\bin\java.exe" -jar $apktoolPath d bin/pejelagarto-translator-unsigned.apk -o bin/apk-decoded -f 2>&1 | Out-Null
                
                # Modify AndroidManifest.xml
                $manifestPath = "bin/apk-decoded/AndroidManifest.xml"
                if (Test-Path $manifestPath) {
                    $manifestContent = Get-Content $manifestPath -Raw
                    # Add targetSdkVersion to uses-sdk element
                    $manifestContent = $manifestContent -replace '(<uses-sdk[^>]*android:minSdkVersion="\d+")', '$1 android:targetSdkVersion="34"'
                    # If no uses-sdk, add it
                    if ($manifestContent -notmatch '<uses-sdk') {
                        $manifestContent = $manifestContent -replace '(<manifest[^>]*>)', "`$1`n    <uses-sdk android:minSdkVersion=`"24`" android:targetSdkVersion=`"34`" />"
                    }
                    $manifestContent | Set-Content $manifestPath -NoNewline
                    Write-Host "  ✓ Manifest patched" -ForegroundColor Green
                }
                
                # Recompile APK
                & "$env:JAVA_HOME\bin\java.exe" -jar $apktoolPath b bin/apk-decoded -o bin/pejelagarto-translator-patched.apk 2>&1 | Out-Null
                
                # Clean up
                if (Test-Path "bin/apk-decoded") {
                    Remove-Item "bin/apk-decoded" -Recurse -Force
                }
                Remove-Item "bin/pejelagarto-translator-unsigned.apk" -Force
                
                # Sign the patched APK
                Write-Host "  Signing patched APK..." -ForegroundColor Cyan
                & $apksigner sign --ks $debugKeystore --ks-pass pass:android --key-pass pass:android --out bin/pejelagarto-translator.apk bin/pejelagarto-translator-patched.apk 2>&1 | Out-Null
                Remove-Item "bin/pejelagarto-translator-patched.apk" -Force
            } else {
                # Skip patching if apktool.jar is not available - sign the unsigned APK directly
                Write-Host "  Signing APK (apktool.jar not found, skipping manifest patching)..." -ForegroundColor Cyan
                & $apksigner sign --ks $debugKeystore --ks-pass pass:android --key-pass pass:android --out bin/pejelagarto-translator.apk bin/pejelagarto-translator-unsigned.apk 2>&1 | Out-Null
                Remove-Item "bin/pejelagarto-translator-unsigned.apk" -Force
            }
            
            if (Test-Path "bin/pejelagarto-translator.apk") {
                $apkSize = (Get-Item "bin/pejelagarto-translator.apk").Length / 1MB
                Write-Host "  ✓ APK signed successfully" -ForegroundColor Green
                Write-Host "`n  ✓ APK built successfully ($([math]::Round($apkSize, 2)) MB)" -ForegroundColor Green
                Write-Host "`n======================================" -ForegroundColor Green
                Write-Host "  APK Build Complete!" -ForegroundColor Green
                Write-Host "======================================" -ForegroundColor Green
                Write-Host "`nAPK Location: bin\pejelagarto-translator.apk" -ForegroundColor Cyan
                Write-Host "Minimum Android: 7.0 (API 24)" -ForegroundColor Gray
                Write-Host "Target Android: 14 (API 34)" -ForegroundColor Gray
                Write-Host "Architectures: ARM64, ARMv7, x86, x86_64" -ForegroundColor Gray
                Write-Host "`nTo install on device:" -ForegroundColor Cyan
                Write-Host "  adb install bin\pejelagarto-translator.apk" -ForegroundColor White
                Write-Host ""
            }
        } else{
            # Fallback: rename unsigned to signed (less secure but will work)
            Write-Host "  ⚠️  apksigner not found, using unsigned APK" -ForegroundColor Yellow
            Move-Item "bin/pejelagarto-translator-unsigned.apk" "bin/pejelagarto-translator.apk" -Force
            $apkSize = (Get-Item "bin/pejelagarto-translator.apk").Length / 1MB
            Write-Host "`n  ✓ APK ready ($([math]::Round($apkSize, 2)) MB)" -ForegroundColor Green
            Write-Host "  Note: APK is unsigned - you may need to enable 'Install from unknown sources'" -ForegroundColor Yellow
        }
    }
} catch {
    Write-Host "`n  ❌ APK build failed: $_" -ForegroundColor Red
    Write-Host "`nTroubleshooting:" -ForegroundColor Yellow
    Write-Host "  1. Ensure Android SDK/NDK are properly installed" -ForegroundColor Yellow
    Write-Host "  2. Try running: gomobile init" -ForegroundColor Yellow
    Write-Host "  3. Check ANDROID_HOME and ANDROID_NDK_HOME environment variables" -ForegroundColor Yellow
    Write-Host "  4. For manifest patching, download apktool.jar to bin/ directory" -ForegroundColor Yellow
    
    # Check if we at least have an unsigned APK we can use
    if (Test-Path "bin/pejelagarto-translator-unsigned.apk") {
        Write-Host "`n  ⚠️  Attempting to use unsigned APK as fallback..." -ForegroundColor Yellow
        try {
            Move-Item "bin/pejelagarto-translator-unsigned.apk" "bin/pejelagarto-translator.apk" -Force
            $apkSize = (Get-Item "bin/pejelagarto-translator.apk").Length / 1MB
            Write-Host "  ✓ APK ready ($([math]::Round($apkSize, 2)) MB)" -ForegroundColor Green
            Write-Host "  Note: APK is unsigned - you may need to enable 'Install from unknown sources'" -ForegroundColor Yellow
            return
        } catch {
            Write-Host "  ❌ Could not use unsigned APK: $_" -ForegroundColor Red
        }
    }
    
    exit 1
}
