# Determine architecture
$arch = if ([System.Runtime.InteropServices.RuntimeInformation]::ProcessArchitecture -eq [System.Runtime.InteropServices.Architecture]::Arm64) { "arm64" } else { "amd64" }

# Download gommit
$gommitUrl = "https://github.com/Moukrea/gommit/releases/latest/download/gommit-windows-$arch"
Invoke-WebRequest -Uri $gommitUrl -OutFile ".gommit\gommit.exe"

# Prepare commit-msg hook content for gommit
$gommitHookContent = @"
# <<<< Gommit managed block

# Set custom hooks here

# Gommit commit message validation
./.gommit/gommit.exe `$1
exit `$LASTEXITCODE
# >>>> Gommit managed block
"@

# Handle commit-msg hook
$destFile = ".git\hooks\commit-msg"

if (Test-Path $destFile) {
    Write-Host "Existing commit-msg hook found."
    $content = Get-Content $destFile -Raw
    if ($content -match "# <<<< Gommit managed block([\s\S]*?)# >>>> Gommit managed block") {
        $existingBlock = $Matches[0]
        if ($existingBlock -eq $gommitHookContent) {
            Write-Host "Gommit hook is up to date. No changes needed."
        } else {
            $content = $content -replace [regex]::Escape($existingBlock), $gommitHookContent
            Set-Content -Path $destFile -Value $content
            Write-Host "Updated existing Gommit managed block in commit-msg hook."
        }
    } else {
        $choice = Read-Host "Gommit hook not found. Choose action (append/skip)"
        switch ($choice.ToLower()) {
            "append" {
                Add-Content -Path $destFile -Value "`n$gommitHookContent"
                Write-Host "Appended Gommit managed block to existing commit-msg hook."
            }
            "skip" {
                Write-Host "Skipped modifying commit-msg hook."
            }
            default {
                Write-Host "Invalid choice. Skipping commit-msg hook modification."
            }
        }
    }
} else {
    Set-Content -Path $destFile -Value "#!/bin/sh`n$gommitHookContent"
    Write-Host "Created new commit-msg hook with Gommit managed block."
}

Write-Host "Gommit has been set up successfully."