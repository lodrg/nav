# nav 安装脚本（Windows PowerShell）
# 用法: irm https://<release-host>/install.ps1 | iex
# 或:   $env:NAV_URL = "https://<release-host>"; irm .\install.ps1 | iex
$ErrorActionPreference = "Stop"

$os = "windows"
$arch = if ($env:PROCESSOR_ARCHITECTURE -match "ARM") { "arm64" } else { "amd64" }

$dest = Join-Path $env:LOCALAPPDATA "bin"
New-Item -ItemType Directory -Force -Path $dest | Out-Null

$base = if ($env:NAV_URL) { $env:NAV_URL } else { "https://github.com/lodrg/nav/releases/latest/download" }
$url = "$base/nav-$os-$arch.exe"
$out = Join-Path $dest "nav.exe"

Write-Host "下载 $url"
Invoke-WebRequest -Uri $url -OutFile $out
Write-Host "已安装: $out"

$paths = [Environment]::GetEnvironmentVariable("Path", "User")
if ($paths -notlike "*$dest*") {
  [Environment]::SetEnvironmentVariable("Path", "$paths;$dest", "User")
  Write-Host "已把 $dest 加入用户 PATH（新开终端生效）"
}
& $out --version
