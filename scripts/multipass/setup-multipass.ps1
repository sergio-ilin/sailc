<#
.SYNOPSIS
    Automated Multipass VM deployment for SailC autonomous MCP agent.
.DESCRIPTION
    Creates a lightweight Ubuntu VM, verifies Python environment,
    and launches an autonomous AI agent synchronizing with the SailC Central Hub.
#>

param (
    [string]$VmName = "sailc-agent-vm",
    [int]$Cpus = 1,
    [string]$Memory = "1G",
    [string]$Disk = "5G",
    [int]$HubPort = 8765
)

$ErrorActionPreference = "Stop"

Write-Host "=== [SailC] Multipass Virtual Machine Provisioning ===" -ForegroundColor Cyan

# 1. Verify Multipass CLI
if (-not (Get-Command multipass -ErrorAction SilentlyContinue)) {
    Write-Error "Multipass not found. Install it via: winget install Canonical.Multipass"
    exit 1
}

# 2. Determine host IP for VM routing
$HostIP = (Get-NetIPAddress -AddressFamily IPv4 | Where-Object { 
    $_.InterfaceAlias -match "Wi-Fi|Ethernet|vEthernet" -and $_.IPAddress -notmatch "^127\.|^169\.254\." 
} | Select-Object -First 1).IPAddress

if (-not $HostIP) {
    $HostIP = "10.0.2.2"
}
Write-Host "[+] Host IP for Central Hub: $HostIP`:$HubPort" -ForegroundColor Green

# 3. Check VM state
$existing = multipass list | Select-String "^$VmName\s"
if ($existing) {
    Write-Host "[*] Virtual machine '$VmName' already exists. Starting..." -ForegroundColor Yellow
    multipass start $VmName
} else {
    Write-Host "[*] Launching new virtual machine '$VmName' ($Cpus CPU, $Memory RAM, $Disk Disk)..." -ForegroundColor Cyan
    multipass launch --name $VmName --cpus $Cpus --memory $Memory --disk $Disk noble
}

# 4. Display VM status
$VmInfo = multipass info $VmName
Write-Host "[+] VM Info:`n$VmInfo" -ForegroundColor Gray

# 5. Transfer worker script
$workerScript = Join-Path $PSScriptRoot "agent-worker.py"
if (Test-Path $workerScript) {
    Write-Host "[*] Transferring agent worker into VM..." -ForegroundColor Cyan
    multipass transfer $workerScript "$VmName`:agent-worker.py"
    multipass exec $VmName -- python3 -c "import urllib.request; print('Python runtime verified')"
    Write-Host "[+] Agent worker ready! Run inside the VM with:" -ForegroundColor Green
    Write-Host "    multipass exec $VmName -- python3 agent-worker.py --hub http://$HostIP`:$HubPort" -ForegroundColor Yellow
}

Write-Host "`n=== [SailC] Deployment completed successfully! ===" -ForegroundColor Green
Write-Host "Shell into VM: multipass shell $VmName"
Write-Host "Stop VM:       multipass stop $VmName"
