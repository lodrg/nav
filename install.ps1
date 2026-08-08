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

# ---- ncd 函数自动配置（PowerShell Profile） ----
$profilePath = $PROFILE
New-Item -ItemType Directory -Force -Path (Split-Path $profilePath) | Out-Null
if (-not (Test-Path $profilePath)) { New-Item -ItemType File -Path $profilePath | Out-Null }

$markerStart = "# >>> nav ncd >>>"
$markerEnd = "# <<< nav ncd <<<"
if ((Get-Content $profilePath -Raw) -notlike "*$markerStart*") {
  $func = @'
# >>> nav ncd >>>
# ncd: 用 nav 导航并 cd（⏎ 选中目录→切过去；→ 深入；q 停在当前目录）
function ncd {
  $d = nav --print @args
  if ($LASTEXITCODE -eq 0 -and $d -and (Test-Path $d -PathType Container)) {
    Set-Location $d
  }
}
# <<< nav ncd <<<
'@
  Add-Content -Path $profilePath -Value "`n$func`n"
  Write-Host "已把 ncd 函数写入 $profilePath（新开终端生效）"
} else {
  Write-Host "ncd 已配置（$profilePath），跳过"
}

$paths = [Environment]::GetEnvironmentVariable("Path", "User")
if ($paths -notlike "*$dest*") {
  [Environment]::SetEnvironmentVariable("Path", "$paths;$dest", "User")
  Write-Host "已把 $dest 加入用户 PATH（新开终端生效）"
}
& $out --version
