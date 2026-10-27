<#
.SYNOPSIS
    Автоматическое развертывание виртуальной машины Multipass для автономного MCP-агента SailC.
.DESCRIPTION
    Создаёт легкую виртуальную машину Ubuntu, устанавливает Python/curl
    и запускает тестового ИИ-агента, синхронизирующегося с центральным хабом SailC.
#>

param (
    [string]$VmName = "sailc-agent-vm",
    [int]$Cpus = 1,
    [string]$Memory = "1G",
    [string]$Disk = "5G",
    [int]$HubPort = 8765
)

$ErrorActionPreference = "Stop"

Write-Host "=== [SailC] Развертывание виртуальной машины Multipass ===" -ForegroundColor Cyan

# 1. Проверка установки Multipass
if (-not (Get-Command multipass -ErrorAction SilentlyContinue)) {
    Write-Error "Multipass не найден. Установите его: winget install Canonical.Multipass"
    exit 1
}

# 2. Определение IP-адреса хоста Windows (для связи из VM)
$HostIP = (Get-NetIPAddress -AddressFamily IPv4 | Where-Object { 
    $_.InterfaceAlias -match "Wi-Fi|Ethernet|vEthernet" -and $_.IPAddress -notmatch "^127\.|^169\.254\." 
} | Select-Object -First 1).IPAddress

if (-not $HostIP) {
    $HostIP = "10.0.2.2"
}
Write-Host "[+] Хостовый IP-адрес для хаба: $HostIP`:$HubPort" -ForegroundColor Green

# 3. Проверка существующей VM
$existing = multipass list | Select-String "^$VmName\s"
if ($existing) {
    Write-Host "[*] Виртуальная машина '$VmName' уже существует. Запуск..." -ForegroundColor Yellow
    multipass start $VmName
} else {
    Write-Host "[*] Создание новой виртуальной машины '$VmName' ($Cpus CPU, $Memory RAM, $Disk Диск)..." -ForegroundColor Cyan
    multipass launch --name $VmName --cpus $Cpus --memory $Memory --disk $Disk noble
}

# 4. Получение IP адреса VM
$VmInfo = multipass info $VmName
Write-Host "[+] Информация о ВМ:`n$VmInfo" -ForegroundColor Gray

# 5. Копирование и запуск скрипта агента
$workerScript = Join-Path $PSScriptRoot "agent-worker.py"
if (Test-Path $workerScript) {
    Write-Host "[*] Передача агента в виртуальную машину..." -ForegroundColor Cyan
    multipass transfer $workerScript "$VmName`:agent-worker.py"
    multipass exec $VmName -- python3 -c "import urllib.request; print('Python готов к работе')"
    Write-Host "[+] Агент готов к запуску внутри VM! Команда для выполнения:" -ForegroundColor Green
    Write-Host "    multipass exec $VmName -- python3 agent-worker.py --hub http://$HostIP`:$HubPort" -ForegroundColor Yellow
}

Write-Host "`n=== [SailC] Развертывание успешно завершено! ===" -ForegroundColor Green
Write-Host "Для входа в консоль виртуальной машины: multipass shell $VmName"
Write-Host "Для остановки виртуальной машины:       multipass stop $VmName"
