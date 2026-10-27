<#
.SYNOPSIS
    Automated configuration script to connect local AI clients to the SailC Hub.
.DESCRIPTION
    Configures Claude Desktop, Antigravity, Cursor, and VS Code MCP settings
    so local AI agents automatically discover and interact with the SailC Central Hub.
#>

$ErrorActionPreference = "Continue"

Write-Host "=== [SailC] Local AI Clients Auto-Connector ===" -ForegroundColor Cyan

$SailcDir = (Resolve-Path "$PSScriptRoot\..").Path
$SailcExe = Join-Path $SailcDir "sailc.exe"
$Port = 8765

Write-Host "[*] SailC Directory: $SailcDir" -ForegroundColor Gray
Write-Host "[*] SailC Executable: $SailcExe" -ForegroundColor Gray

# -------------------------------------------------------------
# 1. Claude Desktop Configuration
# -------------------------------------------------------------
$ClaudeDir = Join-Path $env:APPDATA "Claude"
$ClaudeConfig = Join-Path $ClaudeDir "claude_desktop_config.json"

Write-Host "`n[1] Checking Claude Desktop..." -ForegroundColor Yellow
if (Test-Path $ClaudeDir) {
    $configObj = @{ mcpServers = @{} }
    if (Test-Path $ClaudeConfig) {
        try {
            $configObj = Get-Content $ClaudeConfig -Raw | ConvertFrom-Json
            if (-not $configObj.mcpServers) {
                $configObj | Add-Member -MemberType NoteProperty -Name "mcpServers" -Value @{}
            }
        } catch {}
    } else {
        New-Item -ItemType Directory -Path $ClaudeDir -Force | Out-Null
    }

    $configObj.mcpServers | Add-Member -MemberType NoteProperty -Name "sailc" -Value @{
        command = $SailcExe
        args = @("-mcp-stdio")
    } -Force

    $jsonStr = $configObj | ConvertTo-Json -Depth 5
    [System.IO.File]::WriteAllText($ClaudeConfig, $jsonStr, [System.Text.Encoding]::UTF8)
    Write-Host "[+] Configured Claude Desktop successfully: $ClaudeConfig" -ForegroundColor Green
} else {
    Write-Host "[-] Claude Desktop folder not found ($ClaudeDir). Skipping." -ForegroundColor Gray
}

# -------------------------------------------------------------
# 2. Antigravity MCP Integration
# -------------------------------------------------------------
$AntigravityMcpDir = Join-Path $env:USERPROFILE ".gemini\antigravity\mcp\sailc"
Write-Host "`n[2] Checking Antigravity..." -ForegroundColor Yellow

New-Item -ItemType Directory -Path $AntigravityMcpDir -Force | Out-Null

$tools = @("hub_search", "hub_ask", "hub_answer", "hub_update_tags", "agent_report_status", "agent_list")
foreach ($t in $tools) {
    $toolSchema = @{
        name = $t
        description = "SailC Hub Tool: $t"
        parameters = @{ type = "object"; properties = @{} }
    }
    $schemaPath = Join-Path $AntigravityMcpDir "$t.json"
    $toolSchema | ConvertTo-Json -Depth 4 | Set-Content $schemaPath
}

$instructionsContent = @"
# SailC Central AI Hub Tools
Connects to the local SailC Hub at http://127.0.0.1:$Port/mcp.
- Always use PascalCase tags without spaces (e.g. ProjectSailC, GoBackend).
- Use hub_search before asking new questions.
- Use hub_ask to publish inquiries when context is missing.
- Use hub_update_tags to review and refine chat tags after receiving answers.
"@
[System.IO.File]::WriteAllText((Join-Path $AntigravityMcpDir "instructions.md"), $instructionsContent, [System.Text.Encoding]::UTF8)
Write-Host "[+] Configured Antigravity MCP provider: $AntigravityMcpDir" -ForegroundColor Green

# -------------------------------------------------------------
# 3. Cursor AI Configuration Template
# -------------------------------------------------------------
Write-Host "`n[3] Checking Cursor..." -ForegroundColor Yellow
$cursorRulesSnippet = Join-Path $SailcDir "SAILC_RULES_SNIPPET.md"
Write-Host "[+] In Cursor: Add MCP server 'sailc' via Settings -> Features -> MCP:" -ForegroundColor Gray
Write-Host "    Type: command" -ForegroundColor Gray
Write-Host "    Command: $SailcExe -mcp-stdio" -ForegroundColor Cyan
Write-Host "    Or URL (SSE/HTTP): http://127.0.0.1:$Port/mcp" -ForegroundColor Cyan

Write-Host "`n=== [SailC] Setup Complete! ===" -ForegroundColor Green
Write-Host "Next Step: Place the cognitive rule snippet from SAILC_RULES_SNIPPET.md" -ForegroundColor Yellow
Write-Host "into your project's .cursorrules, CLAUDE.md, or system prompt." -ForegroundColor Yellow
