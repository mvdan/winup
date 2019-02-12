# Need a Win32 class to take ownership of the Registry key
$definition = @"
using System;
using System.Runtime.InteropServices;

namespace Win32Api {
	public class NtDll {
		[DllImport("ntdll.dll", EntryPoint="RtlAdjustPrivilege")]
		public static extern int RtlAdjustPrivilege(ulong Privilege, bool Enable, bool CurrentThread, ref bool Enabled);
	}
}
"@
Add-Type -TypeDefinition $definition -PassThru | Out-Null
[Win32Api.NtDll]::RtlAdjustPrivilege(9, $true, $false, [ref]$false) | Out-Null

# Set ownership to Administrators
$key = [Microsoft.Win32.Registry]::LocalMachine.OpenSubKey("SOFTWARE\Microsoft\Windows NT\CurrentVersion\NetworkList\DefaultMediaCost",[Microsoft.Win32.RegistryKeyPermissionCheck]::ReadWriteSubTree,[System.Security.AccessControl.RegistryRights]::takeownership)
$acl = $key.GetAccessControl()
$acl.SetOwner([System.Security.Principal.NTAccount]"Administrators")
$key.SetAccessControl($acl)

# Give Administrators full control to the key
$rule = New-Object System.Security.AccessControl.RegistryAccessRule ([System.Security.Principal.NTAccount]"Administrators","FullControl","Allow")
$acl.SetAccessRule($rule)
$key.SetAccessControl($acl)

$path = "HKLM:\SOFTWARE\Microsoft\Windows NT\CurrentVersion\NetworkList\DefaultMediaCost"
New-ItemProperty -Path $path -Name "Ethernet" -Value "2" -PropertyType DWORD -Force | Out-Null
Write-Host "Ethernet is now set to metered."
