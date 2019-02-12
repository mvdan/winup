# This script does extra work on top of the downloaded debloater.ps1.

# Try to get background tasks to use less CPU.
Set-AppBackgroundTaskResourcePolicy -Mode Conservative

# Disable background application access, excluding Cortana as it breaks the
# start menu.
Get-ChildItem -Path "HKCU:\Software\Microsoft\Windows\CurrentVersion\BackgroundAccessApplications" -Exclude "Microsoft.Windows.Cortana*" | ForEach {
	Set-ItemProperty -Path $_.PsPath -Name "Disabled" -Type DWord -Value 1
	Set-ItemProperty -Path $_.PsPath -Name "DisabledByUser" -Type DWord -Value 1
}

# Disable scheduled disk defragmentations.
Disable-ScheduledTask -TaskName "Microsoft\Windows\Defrag\ScheduledDefrag" | Out-Null

# Disable the indexing service.
Stop-Service "WSearch" -WarningAction SilentlyContinue
Set-Service "WSearch" -StartupType Disabled

# Disable the Diagnostics Tracking service.
Stop-Service "DiagTrack" -WarningAction SilentlyContinue
Set-Service "DiagTrack" -StartupType Disabled

# Uninstall Windows Defender, since we won't be browsing the internet or
# installing random software.
Uninstall-WindowsFeature -Name Windows-Defender

# Disable Windows updates, since most users won't need them by default.
$WindowsUpdatePath = "HKLM:SOFTWARE\Policies\Microsoft\Windows\WindowsUpdate\"
$AutoUpdatePath = "HKLM:SOFTWARE\Policies\Microsoft\Windows\WindowsUpdate\AU"
If(Test-Path -Path $WindowsUpdatePath) {
	Remove-Item -Path $WindowsUpdatePath -Recurse
}
New-Item -Path $WindowsUpdatePath
New-Item -Path $AutoUpdatePath
Set-ItemProperty -Path $AutoUpdatePath -Name NoAutoUpdate -Value 1

# Remove the Windows Store, since it also likes to run in the background.
Get-AppxPackage *windowsstore* | Remove-AppxPackage

# While at it, remove Edge too, since debloater doesn't do that.
Get-AppxPackage *edge* | Remove-AppxPackage

# TODO: disable background net usage, but how?
