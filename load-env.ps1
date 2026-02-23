Get-Content "$PSScriptRoot\.env" | ForEach-Object {
    if ($_ -match '^\s*([^#][^=]+)=(.*)$') {
        $name = $Matches[1].Trim()
        $value = $Matches[2].Trim().Trim('"')
        [Environment]::SetEnvironmentVariable($name, $value, 'Process')
    }
}

Write-Host "Atlas env loaded (ATLAS_IMAP_HOST=$env:ATLAS_IMAP_HOST)"
