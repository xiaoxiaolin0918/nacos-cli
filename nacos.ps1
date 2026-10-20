# Windows entry for nacos-cli (Go).
$ErrorActionPreference = "Stop"
$Root = Split-Path -Parent $MyInvocation.MyCommand.Path
$Exe = Join-Path $Root "nacos.exe"
if (-not (Test-Path $Exe)) {
  $env:PATH = "D:\tools\go\bin;" + $env:PATH
  $env:GOPROXY = "goproxy.cn,direct"
  Push-Location $Root
  go build -o nacos.exe .
  Pop-Location
}
& $Exe @args
exit $LASTEXITCODE
