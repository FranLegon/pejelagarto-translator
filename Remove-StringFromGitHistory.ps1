<#
.SYNOPSIS
    Functions to remove sensitive strings and large files from Git history.

.DESCRIPTION
    This script provides PowerShell functions to safely rewrite Git history to:
    - Replace sensitive strings across all commits
    - Remove files that exceed GitHub's size limits (100MB hard limit, 50MB warning)
    
.NOTES
    WARNING: These operations rewrite Git history and require force-pushing.
    Always create a backup branch before running these functions.
    
    Author: Generated based on git-filter-branch experience
    Date: November 8, 2025
#>

function Remove-StringFromGitHistory {
    <#
    .SYNOPSIS
        Replace a string across entire Git history.
    
    .DESCRIPTION
        Uses git filter-branch to replace all instances of a string in the repository history.
        Also removes files larger than GitHub's limits (100MB).
    
    .PARAMETER ReplaceString
        The string to search for and replace (e.g., a leaked API token).
    
    .PARAMETER WithString
        The replacement string (e.g., "REDACTED_TOKEN").
    
    .PARAMETER BackupBranch
        Name of the backup branch to create before rewriting. Default: "backup-before-rewrite-$(Get-Date -Format 'yyyy-MM-dd-HHmmss')"
    
    .PARAMETER MaxFileSizeMB
        Maximum file size in MB. Files larger than this will be removed from history. Default: 99 (just under GitHub's 100MB limit)
    
    .PARAMETER DeleteBackupAfterSuccess
        If specified, the backup branch will be deleted after successful completion. Default: $false (keeps backup for safety)
    
    .EXAMPLE
        Remove-StringFromGitHistory -ReplaceString "secret_api_key_12345" -WithString "REDACTED_API_KEY"
    
    .EXAMPLE
        Remove-StringFromGitHistory -ReplaceString "password123" -WithString "REDACTED" -BackupBranch "backup-before-password-removal"
    
    .EXAMPLE
        Remove-StringFromGitHistory -ReplaceString "token123" -WithString "REDACTED" -DeleteBackupAfterSuccess
    #>
    [CmdletBinding(SupportsShouldProcess=$true, ConfirmImpact='High')]
    param(
        [Parameter(Mandatory=$true)]
        [string]$ReplaceString,
        
        [Parameter(Mandatory=$true)]
        [string]$WithString,
        
        [Parameter(Mandatory=$false)]
        [string]$BackupBranch = "backup-before-rewrite-$(Get-Date -Format 'yyyy-MM-dd-HHmmss')",
        
        [Parameter(Mandatory=$false)]
        [int]$MaxFileSizeMB = 99,
        
        [Parameter(Mandatory=$false)]
        [switch]$DeleteBackupAfterSuccess
    )
    
    # Verify we're in a git repository
    if (-not (Test-Path .git)) {
        Write-Error "Not in a Git repository. Please run this from the repository root."
        return
    }
    
    # Check for uncommitted changes
    $status = git status --porcelain
    if ($status) {
        Write-Error "You have uncommitted changes. Please commit or stash them before rewriting history."
        Write-Host "Run: git status" -ForegroundColor Yellow
        return
    }
    
    # Confirm with user
    Write-Host "`n⚠️  WARNING: This will rewrite Git history!" -ForegroundColor Red
    Write-Host "This operation will:" -ForegroundColor Yellow
    Write-Host "  1. Replace all instances of '$ReplaceString' with '$WithString'" -ForegroundColor Yellow
    Write-Host "  2. Remove all files larger than ${MaxFileSizeMB}MB" -ForegroundColor Yellow
    Write-Host "  3. Create backup branch: $BackupBranch" -ForegroundColor Yellow
    Write-Host "  4. Require force-push to update remote" -ForegroundColor Yellow
    
    if (-not $PSCmdlet.ShouldProcess("Git repository", "Rewrite history")) {
        Write-Host "Operation cancelled." -ForegroundColor Cyan
        return
    }
    
    # Create backup branch
    Write-Host "`n📦 Creating backup branch: $BackupBranch..." -ForegroundColor Cyan
    git branch $BackupBranch
    if ($LASTEXITCODE -ne 0) {
        Write-Error "Failed to create backup branch. Aborting."
        return
    }
    Write-Host "✅ Backup branch created successfully" -ForegroundColor Green
    
    # Set environment variable to suppress filter-branch warning
    $env:FILTER_BRANCH_SQUELCH_WARNING = '1'
    
    # Create temporary PowerShell script for filter-branch
    $tempScript = Join-Path $env:TEMP "git-filter-$(Get-Date -Format 'yyyyMMddHHmmss').ps1"
    $maxBytes = $MaxFileSizeMB * 1024 * 1024
    
    # Escape special regex characters in strings
    $escapedReplace = [regex]::Escape($ReplaceString)
    
    $scriptContent = @"
# Remove large files
Get-ChildItem -Recurse -File -ErrorAction SilentlyContinue | Where-Object { `$_.Length -gt $maxBytes } | Remove-Item -Force -ErrorAction SilentlyContinue

# Replace string in all text files
`$extensions = @('*.ps1', '*.sh', '*.go', '*.txt', '*.md', '*.json', '*.yaml', '*.yml', '*.js', '*.ts', '*.py', '*.java', '*.cs', '*.cpp', '*.c', '*.h', '*.bat', '*.cmd', '*.config', '*.xml', '*.ini')
Get-ChildItem -Recurse -File -Include `$extensions -ErrorAction SilentlyContinue | ForEach-Object {
    try {
        `$content = Get-Content `$_.FullName -Raw -ErrorAction Stop
        if (`$content -match '$escapedReplace') {
            `$newContent = `$content -replace '$escapedReplace', '$WithString'
            Set-Content -Path `$_.FullName -Value `$newContent -NoNewline -ErrorAction Stop
        }
    } catch {
        # Silently continue on errors (binary files, permission issues, etc.)
    }
}
"@
    
    # Don't create temp script - use find/sed directly which is MUCH faster
    Remove-Item $tempScript -ErrorAction SilentlyContinue
    
    # Run git filter-branch using fast find/sed approach
    Write-Host "`n🔄 Rewriting Git history (this may take several minutes)..." -ForegroundColor Cyan
    Write-Host "Processing all commits and branches..." -ForegroundColor Gray
    
    # Use find + sed which is orders of magnitude faster than spawning PowerShell for each commit
    $findCmd = "find . -type f \( -name '*.ps1' -o -name '*.sh' -o -name '*.go' -o -name '*.txt' -o -name '*.md' -o -name '*.json' -o -name '*.yaml' -o -name '*.yml' -o -name '*.js' -o -name '*.ts' -o -name '*.py' -o -name '*.java' -o -name '*.cs' -o -name '*.cpp' -o -name '*.c' -o -name '*.h' -o -name '*.bat' -o -name '*.cmd' -o -name '*.config' -o -name '*.xml' -o -name '*.ini' \) -exec sed -i 's/$escapedReplace/$WithString/g' {} + 2>/dev/null || true; find . -type f -size +${maxBytes}c -delete 2>/dev/null || true"
    $filterCmd = "git filter-branch --force --tree-filter `"$findCmd`" --tag-name-filter cat --prune-empty -- --all"
    
    try {
        Invoke-Expression $filterCmd
        
        if ($LASTEXITCODE -eq 0) {
            Write-Host "✅ History rewrite completed successfully" -ForegroundColor Green
        } else {
            Write-Error "Git filter-branch failed with exit code $LASTEXITCODE"
            return
        }
    }
    catch {
        Write-Error "Error during history rewrite: $_"
        return
    }
    
    # Verify string replacement
    Write-Host "`n🔍 Verifying string removal..." -ForegroundColor Cyan
    $found = git log --all --source --full-history -S $ReplaceString --oneline
    if ($found) {
        Write-Warning "String '$ReplaceString' still found in history!"
        Write-Host $found -ForegroundColor Yellow
    } else {
        Write-Host "✅ String '$ReplaceString' successfully removed from all commits" -ForegroundColor Green
    }
    
    # Clean up filter-branch references
    Write-Host "`n🧹 Cleaning up..." -ForegroundColor Cyan
    git for-each-ref --format="delete %(refname)" refs/original/ | git update-ref --stdin
    git reflog expire --expire=now --all
    
    # Run garbage collection (with error suppression for file locks on Windows)
    Write-Host "Running garbage collection..." -ForegroundColor Gray
    git gc --prune=now --aggressive 2>&1 | Out-Null
    
    Write-Host "✅ Cleanup completed" -ForegroundColor Green
    
    # Optionally delete backup branch
    if ($DeleteBackupAfterSuccess) {
        Write-Host "`n🗑️  Deleting backup branch: $BackupBranch..." -ForegroundColor Cyan
        git branch -D $BackupBranch
        if ($LASTEXITCODE -eq 0) {
            Write-Host "✅ Backup branch deleted" -ForegroundColor Green
        } else {
            Write-Warning "Failed to delete backup branch"
        }
    }
    
    # Show summary
    Write-Host "`n" + "="*60 -ForegroundColor Cyan
    Write-Host "📊 SUMMARY" -ForegroundColor Cyan
    Write-Host "="*60 -ForegroundColor Cyan
    Write-Host "✅ String replaced: '$ReplaceString' → '$WithString'" -ForegroundColor Green
    Write-Host "✅ Large files (>${MaxFileSizeMB}MB) removed from history" -ForegroundColor Green
    if ($DeleteBackupAfterSuccess) {
        Write-Host "✅ Backup branch deleted: $BackupBranch" -ForegroundColor Green
    } else {
        Write-Host "✅ Backup created: $BackupBranch" -ForegroundColor Green
    }
    Write-Host ""
    Write-Host "⚠️  NEXT STEPS:" -ForegroundColor Yellow
    Write-Host "1. Review changes: git log --oneline -20" -ForegroundColor White
    Write-Host "2. Force push to remote: git push --force --all" -ForegroundColor White
    Write-Host "3. Force push tags: git push --force --tags" -ForegroundColor White
    Write-Host "4. Notify collaborators to re-clone the repository" -ForegroundColor White
    Write-Host "5. If removing secrets, REVOKE the old credentials immediately!" -ForegroundColor Red
    Write-Host ""
    if (-not $DeleteBackupAfterSuccess) {
        Write-Host "To undo: git reset --hard $BackupBranch" -ForegroundColor Gray
        Write-Host "To delete backup later: git branch -D $BackupBranch" -ForegroundColor Gray
    }
    Write-Host "="*60 -ForegroundColor Cyan
}

function Remove-StringsFromGitHistory {
    <#
    .SYNOPSIS
        Replace multiple strings across entire Git history.
    
    .DESCRIPTION
        Uses git filter-branch to replace multiple string pairs in the repository history.
        Also removes files larger than GitHub's limits (100MB).
    
    .PARAMETER ReplacementsHashtable
        A hashtable where keys are strings to replace and values are replacement strings.
        Example: @{ "secret1" = "REDACTED1"; "api_key_xyz" = "REDACTED_API_KEY" }
    
    .PARAMETER BackupBranch
        Name of the backup branch to create before rewriting. Default: "backup-before-rewrite-$(Get-Date -Format 'yyyy-MM-dd-HHmmss')"
    
    .PARAMETER MaxFileSizeMB
        Maximum file size in MB. Files larger than this will be removed from history. Default: 99 (just under GitHub's 100MB limit)
    
    .PARAMETER DeleteBackupAfterSuccess
        If specified, the backup branch will be deleted after successful completion. Default: $false (keeps backup for safety)
    
    .EXAMPLE
        $replacements = @{
            "old_api_key_12345" = "REDACTED_API_KEY"
            "password123" = "REDACTED_PASSWORD"
            "secret_token_xyz" = "REDACTED_TOKEN"
        }
        Remove-StringsFromGitHistory -ReplacementsHashtable $replacements
    
    .EXAMPLE
        Remove-StringsFromGitHistory -ReplacementsHashtable @{ "leak1" = "REDACTED"; "leak2" = "REDACTED" } -BackupBranch "backup-security-fix"
    
    .EXAMPLE
        Remove-StringsFromGitHistory -ReplacementsHashtable $secrets -DeleteBackupAfterSuccess
    #>
    [CmdletBinding(SupportsShouldProcess=$true, ConfirmImpact='High')]
    param(
        [Parameter(Mandatory=$true)]
        [hashtable]$ReplacementsHashtable,
        
        [Parameter(Mandatory=$false)]
        [string]$BackupBranch = "backup-before-rewrite-$(Get-Date -Format 'yyyy-MM-dd-HHmmss')",
        
        [Parameter(Mandatory=$false)]
        [int]$MaxFileSizeMB = 99,
        
        [Parameter(Mandatory=$false)]
        [switch]$DeleteBackupAfterSuccess
    )
    
    # Verify we're in a git repository
    if (-not (Test-Path .git)) {
        Write-Error "Not in a Git repository. Please run this from the repository root."
        return
    }
    
    # Validate input
    if ($ReplacementsHashtable.Count -eq 0) {
        Write-Error "ReplacementsHashtable is empty. Please provide at least one string to replace."
        return
    }
    
    # Check for uncommitted changes
    $status = git status --porcelain
    if ($status) {
        Write-Error "You have uncommitted changes. Please commit or stash them before rewriting history."
        Write-Host "Run: git status" -ForegroundColor Yellow
        return
    }
    
    # Confirm with user
    Write-Host "`n⚠️  WARNING: This will rewrite Git history!" -ForegroundColor Red
    Write-Host "This operation will:" -ForegroundColor Yellow
    Write-Host "  1. Replace $($ReplacementsHashtable.Count) different strings" -ForegroundColor Yellow
    Write-Host "  2. Remove all files larger than ${MaxFileSizeMB}MB" -ForegroundColor Yellow
    Write-Host "  3. Create backup branch: $BackupBranch" -ForegroundColor Yellow
    Write-Host "  4. Require force-push to update remote" -ForegroundColor Yellow
    Write-Host "`nReplacements:" -ForegroundColor Cyan
    foreach ($key in $ReplacementsHashtable.Keys) {
        Write-Host "  '$key' → '$($ReplacementsHashtable[$key])'" -ForegroundColor Gray
    }
    
    if (-not $PSCmdlet.ShouldProcess("Git repository", "Rewrite history with $($ReplacementsHashtable.Count) replacements")) {
        Write-Host "Operation cancelled." -ForegroundColor Cyan
        return
    }
    
    # Create backup branch
    Write-Host "`n📦 Creating backup branch: $BackupBranch..." -ForegroundColor Cyan
    git branch $BackupBranch
    if ($LASTEXITCODE -ne 0) {
        Write-Error "Failed to create backup branch. Aborting."
        return
    }
    Write-Host "✅ Backup branch created successfully" -ForegroundColor Green
    
    # Set environment variable to suppress filter-branch warning
    $env:FILTER_BRANCH_SQUELCH_WARNING = '1'
    
    # Create temporary PowerShell script for filter-branch
    $tempScript = Join-Path $env:TEMP "git-filter-$(Get-Date -Format 'yyyyMMddHHmmss').ps1"
    $maxBytes = $MaxFileSizeMB * 1024 * 1024
    
    # Build replacement commands for all strings
    $replacementCommands = ""
    foreach ($key in $ReplacementsHashtable.Keys) {
        $escapedKey = [regex]::Escape($key)
        $value = $ReplacementsHashtable[$key]
        $replacementCommands += "        if (`$content -match '$escapedKey') {`n"
        $replacementCommands += "            `$content = `$content -replace '$escapedKey', '$value'`n"
        $replacementCommands += "            `$modified = `$true`n"
        $replacementCommands += "        }`n"
    }
    
    $scriptContent = @"
# Remove large files
Get-ChildItem -Recurse -File -ErrorAction SilentlyContinue | Where-Object { `$_.Length -gt $maxBytes } | Remove-Item -Force -ErrorAction SilentlyContinue

# Replace multiple strings in all text files
`$extensions = @('*.ps1', '*.sh', '*.go', '*.txt', '*.md', '*.json', '*.yaml', '*.yml', '*.js', '*.ts', '*.py', '*.java', '*.cs', '*.cpp', '*.c', '*.h', '*.bat', '*.cmd', '*.config', '*.xml', '*.ini')
Get-ChildItem -Recurse -File -Include `$extensions -ErrorAction SilentlyContinue | ForEach-Object {
    try {
        `$content = Get-Content `$_.FullName -Raw -ErrorAction Stop
        `$modified = `$false
        
$replacementCommands
        
        if (`$modified) {
            Set-Content -Path `$_.FullName -Value `$content -NoNewline -ErrorAction Stop
        }
    } catch {
        # Silently continue on errors (binary files, permission issues, etc.)
    }
}
"@
    
    # Don't create temp script - use find/sed directly which is MUCH faster
    Remove-Item $tempScript -ErrorAction SilentlyContinue
    
    # Run git filter-branch using fast find/sed approach
    Write-Host "`n🔄 Rewriting Git history (this may take several minutes)..." -ForegroundColor Cyan
    Write-Host "Processing all commits and branches..." -ForegroundColor Gray
    
    # Build a single sed command for all replacements
    $sedReplacements = ($sedCommands -join "; ")
    $findCmd = "find . -type f \( -name '*.ps1' -o -name '*.sh' -o -name '*.go' -o -name '*.txt' -o -name '*.md' -o -name '*.json' -o -name '*.yaml' -o -name '*.yml' -o -name '*.js' -o -name '*.ts' -o -name '*.py' -o -name '*.java' -o -name '*.cs' -o -name '*.cpp' -o -name '*.c' -o -name '*.h' -o -name '*.bat' -o -name '*.cmd' -o -name '*.config' -o -name '*.xml' -o -name '*.ini' \) -exec sed -i '$sedReplacements' {} + 2>/dev/null || true; find . -type f -size +${maxBytes}c -delete 2>/dev/null || true"
    $filterCmd = "git filter-branch --force --tree-filter `"$findCmd`" --tag-name-filter cat --prune-empty -- --all"
    
    try {
        Invoke-Expression $filterCmd
        
        if ($LASTEXITCODE -eq 0) {
            Write-Host "✅ History rewrite completed successfully" -ForegroundColor Green
        } else {
            Write-Error "Git filter-branch failed with exit code $LASTEXITCODE"
            return
        }
    }
    catch {
        Write-Error "Error during history rewrite: $_"
        return
    }
    
    # Verify string replacements
    Write-Host "`n🔍 Verifying string removals..." -ForegroundColor Cyan
    $allSuccess = $true
    foreach ($key in $ReplacementsHashtable.Keys) {
        $found = git log --all --source --full-history -S $key --oneline
        if ($found) {
            Write-Warning "String '$key' still found in history!"
            Write-Host $found -ForegroundColor Yellow
            $allSuccess = $false
        } else {
            Write-Host "✅ '$key' successfully removed" -ForegroundColor Green
        }
    }
    
    if ($allSuccess) {
        Write-Host "`n✅ All strings successfully removed from history" -ForegroundColor Green
    }
    
    # Clean up filter-branch references
    Write-Host "`n🧹 Cleaning up..." -ForegroundColor Cyan
    git for-each-ref --format="delete %(refname)" refs/original/ | git update-ref --stdin
    git reflog expire --expire=now --all
    
    # Run garbage collection (with error suppression for file locks on Windows)
    Write-Host "Running garbage collection..." -ForegroundColor Gray
    git gc --prune=now --aggressive 2>&1 | Out-Null
    
    Write-Host "✅ Cleanup completed" -ForegroundColor Green
    
    # Optionally delete backup branch
    if ($DeleteBackupAfterSuccess) {
        Write-Host "`n🗑️  Deleting backup branch: $BackupBranch..." -ForegroundColor Cyan
        git branch -D $BackupBranch
        if ($LASTEXITCODE -eq 0) {
            Write-Host "✅ Backup branch deleted" -ForegroundColor Green
        } else {
            Write-Warning "Failed to delete backup branch"
        }
    }
    
    # Show summary
    Write-Host "`n" + "="*60 -ForegroundColor Cyan
    Write-Host "📊 SUMMARY" -ForegroundColor Cyan
    Write-Host "="*60 -ForegroundColor Cyan
    Write-Host "✅ $($ReplacementsHashtable.Count) strings replaced successfully" -ForegroundColor Green
    Write-Host "✅ Large files (>${MaxFileSizeMB}MB) removed from history" -ForegroundColor Green
    if ($DeleteBackupAfterSuccess) {
        Write-Host "✅ Backup branch deleted: $BackupBranch" -ForegroundColor Green
    } else {
        Write-Host "✅ Backup created: $BackupBranch" -ForegroundColor Green
    }
    Write-Host ""
    Write-Host "Replacements performed:" -ForegroundColor Cyan
    foreach ($key in $ReplacementsHashtable.Keys) {
        Write-Host "  '$key' → '$($ReplacementsHashtable[$key])'" -ForegroundColor Gray
    }
    Write-Host ""
    Write-Host "⚠️  NEXT STEPS:" -ForegroundColor Yellow
    Write-Host "1. Review changes: git log --online -20" -ForegroundColor White
    Write-Host "2. Force push to remote: git push --force --all" -ForegroundColor White
    Write-Host "3. Force push tags: git push --force --tags" -ForegroundColor White
    Write-Host "4. Notify collaborators to re-clone the repository" -ForegroundColor White
    Write-Host "5. If removing secrets, REVOKE all old credentials immediately!" -ForegroundColor Red
    Write-Host ""
    if (-not $DeleteBackupAfterSuccess) {
        Write-Host "To undo: git reset --hard $BackupBranch" -ForegroundColor Gray
        Write-Host "To delete backup later: git branch -D $BackupBranch" -ForegroundColor Gray
    }
    Write-Host "="*60 -ForegroundColor Cyan
}

# Export functions (only works when loaded as module, not when dot-sourced)
if ((Get-Command -Name Export-ModuleMember -ErrorAction SilentlyContinue) -and $MyInvocation.MyCommand.ModuleName) {
    Export-ModuleMember -Function Remove-StringFromGitHistory, Remove-StringsFromGitHistory
}
