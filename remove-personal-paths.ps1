# Remove personal information from git history

. .\Remove-StringFromGitHistory.ps1

$paths = @{
    "REDACTED_USER" = "REDACTED_USER"
    "REDACTED_COMPANY" = "REDACTED_COMPANY"
    "C:\Users\REDACTED_USER\OneDrive - REDACTED_COMPANY\Documentos\Personal\Go\pejelagarto-translator\" = "REDACTED_PATH/"
    "C:\Users\REDACTED_USER\OneDrive - REDACTED_COMPANY\Documentos\Personal\Go\pejelagarto-translator" = "REDACTED_PATH"
}

Write-Host "About to remove the following strings from git history:" -ForegroundColor Yellow
foreach ($key in $paths.Keys) {
    Write-Host "  '$key' → '$($paths[$key])'" -ForegroundColor Gray
}

Write-Host "`nThis will rewrite the entire git history." -ForegroundColor Yellow
$confirm = Read-Host "Continue? (y/n)"

if ($confirm -ne 'y') {
    Write-Host "Aborted" -ForegroundColor Red
    exit
}

Remove-StringsFromGitHistory -ReplacementsHashtable $paths -BackupBranch "backup-before-path-removal"
