# SailC Background Daemon Setup Script
# Configures 24/7 background MCP service for local AI agents without console window

param(
    [ValidateSet("install", "uninstall", "status", "start", "stop")]
    [string]$Action = "install",
    [int]$Port = 8765
)

$DaemonPath = Join-Path $PSScriptRoot "..\sailcd.exe"
if (-not (Test-Path $DaemonPath)) {
    $DaemonPath = (Get-Command sailcd -ErrorAction SilentlyContinue).Source
}

if (-not $DaemonPath -or -not (Test-Path $DaemonPath)) {
    Write-Error "sailcd.exe not found. Please build with: go build -ldflags `"-H=windowsgui -s -w`" -o sailcd.exe ./cmd/sailcd"
    exit 1
}

$ResolvedDaemon = (Resolve-Path $DaemonPath).Path

switch ($Action) {
    "install" {
        Write-Host "Registering SailC 24/7 Background MCP Hub..." -ForegroundColor Cyan
        
        # Register in HKCU Run registry (Runs automatically on user login, 0 admin required)
        & $ResolvedDaemon -install-startup -port $Port

        # Launch background daemon detached and hidden
        Start-Process -FilePath $ResolvedDaemon -ArgumentList "-port $Port" -WindowStyle Hidden

        Start-Sleep -Seconds 1

        Write-Host "[OK] SailC Background MCP Hub is registered in Windows Startup and running." -ForegroundColor Green
        Write-Host "[OK] Zero console windows will ever pop up. Local AI agents can now connect 24/7." -ForegroundColor Green
    }

    "uninstall" {
        Write-Host "Removing SailC Background Service..." -ForegroundColor Yellow
        & $ResolvedDaemon -uninstall-startup
        Get-Process -Name "sailcd" -ErrorAction SilentlyContinue | Where-Object { $_.Path -eq $ResolvedDaemon } | Stop-Process -Force -ErrorAction SilentlyContinue
        Write-Host "[OK] SailC background service removed." -ForegroundColor Green
    }

    "status" {
        $portActive = $false
        try {
            $tcp = New-Object System.Net.Sockets.TcpClient
            $tcp.Connect("127.0.0.1", $Port)
            $portActive = $tcp.Connected
            $tcp.Close()
        } catch {
            $portActive = $false
        }

        if ($portActive) {
            Write-Host "[STATUS] Background MCP Hub is ACTIVE on http://127.0.0.1:$Port/mcp" -ForegroundColor Green
        } else {
            Write-Host "[STATUS] Background MCP Hub is NOT running on port $Port" -ForegroundColor Red
        }
    }

    "start" {
        Start-Process -FilePath $ResolvedDaemon -ArgumentList "-port $Port" -WindowStyle Hidden
        Write-Host "[OK] Background daemon started." -ForegroundColor Green
    }

    "stop" {
        Get-Process -Name "sailcd" -ErrorAction SilentlyContinue | Where-Object { $_.Path -eq $ResolvedDaemon } | Stop-Process -Force -ErrorAction SilentlyContinue
        Write-Host "[OK] Background daemon stopped." -ForegroundColor Yellow
    }
}
