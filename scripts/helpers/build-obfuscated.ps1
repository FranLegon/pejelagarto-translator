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
# Build the entire package (.) instead of main.go to properly handle build tags
garble -literals -tiny build -tags obfuscated -o $outputFile .

if ($LASTEXITCODE -eq 0) {
    Write-Host "Build successful: $outputFile"
} else {
    Write-Host "Build failed with exit code $LASTEXITCODE"
    exit $LASTEXITCODE
}
