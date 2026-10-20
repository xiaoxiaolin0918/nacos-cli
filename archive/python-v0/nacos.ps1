# nacos-cli 入口（Windows PowerShell）
# 用法: .\nacos.ps1 --help
$ErrorActionPreference = "Stop"
$Root = Split-Path -Parent $MyInvocation.MyCommand.Path
Set-Location $Root
$env:PYTHONIOENCODING = "utf-8"
$env:PYTHONUTF8 = "1"
try { [Console]::OutputEncoding = [System.Text.Encoding]::UTF8 } catch {}
python -m nacos_cli @args
exit $LASTEXITCODE
