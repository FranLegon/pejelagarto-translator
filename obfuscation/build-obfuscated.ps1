param(
    [string]$OS = "windows"
)

# Determine output filename based on OS
$outputFile = if ($OS -eq "windows") {
    "bin/piper-server.exe"
} else {
    "bin/piper-server"
}

Write-Host "Building obfuscated version for $OS..."
Write-Host "Output: $outputFile"

# Run garble with obfuscation flags
garble -literals -tiny build -tags obfuscated -o $outputFile main.go

if ($LASTEXITCODE -eq 0) {
    Write-Host "Build successful: $outputFile"
} else {
    Write-Host "Build failed with exit code $LASTEXITCODE"
    exit $LASTEXITCODE
}
