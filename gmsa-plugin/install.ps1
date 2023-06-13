# This is a slightly modified version of the publicly available AKS ccg plugin install script which includes many
# more comments which are helpful to those uninitiated with Windows internals / Concepts. 

# Note: 
# You need to run this script from the same directory as the RanchergMSACredentialProvider binary. 
# You need to have the Containers feature enabled for any of this to work.
# You need to run this script as an administrator.

## GLOBAL CONSTANTS
## DO NOT CHANGE GUIDs

$appId = "24DC734A-E2D4-4F12-A387-F6209CBAF4FC" 
$CLSID = "E4781092-F116-4B79-B55E-28EB6A224E26" 
$dllName  = "RanchergMSACredentialProvider.dll"
$dllDirectoryPath = "C:\Program Files\RanchergMSACredentialProvider\"
$dllFileLocationEscaped = "C:\\Program Files\\RanchergMSACredentialProvider\\RanchergMSACredentialProvider.dll"

<#
    This file registers the CCG plugin DLL onto the windows system. It does so by doing the following:

    1. Place the dll file in an appropriate location on the system. This location is referenced by the registry.
    2. Define a utility function which allows us to enable or disable privileges for a user.
    3. Create and Register an AppId for the COM component in the registry.
    4. Define launch and access permissions for the AppId using the Security Descriptor Definition Lanauge (sddl).
    5. Register a CLSID which uses the dll file and the AppId to provide a GUID for the COM class.
    6. Configure thread access rules for the COM object using the AppId.
    7. Configure access rules for the current user so we can register the COM component in this script.
    8. Actually register the COM component in the private registry.
    9. Clean up access rules and privileges.
#>

# 1. Place the dll file in a more appropriate location on the system
mkdir $dllDirectoryPath
copy $dllName $dllDirectoryPath


# 2. This function Enables or disables a specified privilege for the current user.
function enable-privilege {
 param(
  ## The privilege to adjust. This set is taken from
  ## http://msdn.microsoft.com/en-us/library/bb530716(VS.85).aspx
  [ValidateSet(
   "SeAssignPrimaryTokenPrivilege", "SeAuditPrivilege", "SeBackupPrivilege",
   "SeChangeNotifyPrivilege", "SeCreateGlobalPrivilege", "SeCreatePagefilePrivilege",
   "SeCreatePermanentPrivilege", "SeCreateSymbolicLinkPrivilege", "SeCreateTokenPrivilege",
   "SeDebugPrivilege", "SeEnableDelegationPrivilege", "SeImpersonatePrivilege", "SeIncreaseBasePriorityPrivilege",
   "SeIncreaseQuotaPrivilege", "SeIncreaseWorkingSetPrivilege", "SeLoadDriverPrivilege",
   "SeLockMemoryPrivilege", "SeMachineAccountPrivilege", "SeManageVolumePrivilege",
   "SeProfileSingleProcessPrivilege", "SeRelabelPrivilege", "SeRemoteShutdownPrivilege",
   "SeRestorePrivilege", "SeSecurityPrivilege", "SeShutdownPrivilege", "SeSyncAgentPrivilege",
   "SeSystemEnvironmentPrivilege", "SeSystemProfilePrivilege", "SeSystemtimePrivilege",
   "SeTakeOwnershipPrivilege", "SeTcbPrivilege", "SeTimeZonePrivilege", "SeTrustedCredManAccessPrivilege",
   "SeUndockPrivilege", "SeUnsolicitedInputPrivilege")]
  $Privilege,
  ## The process on which to adjust the privilege. Defaults to the current process.
  $ProcessId = $pid,
  ## Switch to disable the privilege, rather than enable it.
  [Switch] $Disable
 )

 ## Taken from P/Invoke.NET with minor adjustments.
 # DO NOT MODIFY - Source: https://github.com/microsoft/Azure-Key-Vault-Plugin-gMSA/blob/main/src/CCGAKVPlugin/InstallPlugin.ps1
 $definition = @'
 using System;
 using System.Runtime.InteropServices;

 public class AdjPriv
 {
  [DllImport("advapi32.dll", ExactSpelling = true, SetLastError = true)]
  internal static extern bool AdjustTokenPrivileges(IntPtr htok, bool disall,
   ref TokPriv1Luid newst, int len, IntPtr prev, IntPtr relen);
  
  [DllImport("advapi32.dll", ExactSpelling = true, SetLastError = true)]
  internal static extern bool OpenProcessToken(IntPtr h, int acc, ref IntPtr phtok);
  [DllImport("advapi32.dll", SetLastError = true)]
  internal static extern bool LookupPrivilegeValue(string host, string name, ref long pluid);
  [StructLayout(LayoutKind.Sequential, Pack = 1)]
  internal struct TokPriv1Luid
  {
   public int Count;
   public long Luid;
   public int Attr;
  }
  
  internal const int SE_PRIVILEGE_ENABLED = 0x00000002;
  internal const int SE_PRIVILEGE_DISABLED = 0x00000000;
  internal const int TOKEN_QUERY = 0x00000008;
  internal const int TOKEN_ADJUST_PRIVILEGES = 0x00000020;
  public static bool EnablePrivilege(long processHandle, string privilege, bool disable)
  {
   bool retVal;
   TokPriv1Luid tp;
   IntPtr hproc = new IntPtr(processHandle);
   IntPtr htok = IntPtr.Zero;
   retVal = OpenProcessToken(hproc, TOKEN_ADJUST_PRIVILEGES | TOKEN_QUERY, ref htok);
   tp.Count = 1;
   tp.Luid = 0;
   if(disable)
   {
    tp.Attr = SE_PRIVILEGE_DISABLED;
   }
   else
   {
    tp.Attr = SE_PRIVILEGE_ENABLED;
   }
   retVal = LookupPrivilegeValue(null, privilege, ref tp.Luid);
   retVal = AdjustTokenPrivileges(htok, false, ref tp, 0, IntPtr.Zero, IntPtr.Zero);
   return retVal;
  }
 }
'@

 $processHandle = (Get-Process -id $ProcessId).Handle
 $type = Add-Type $definition -PassThru
 $type[0]::EnablePrivilege($processHandle, $Privilege, $Disable)
}

# 3. Register the class via an AppID in the Registry under the same GUID as specified in the source code.
New-Item -path "HKLM:\Software\CLASSES\Appid\{$appId}"

# 4. sddl stands for 'Security Descriptor Definition Language'. This configures Windows Access Control Entries for the COM object so it can be called. 
# don't touch this. 
$sddlString = "01,00,04,80,44,00,00,00,54,00,00,00,00,00,00,00,14,00,00,00,02,00,30,00,02,00,00,00,00,00,14,00,0B,00,00,00,01,01,00,00,00,00,00,05,12,00,00,00,00,00,14,00,0B,00,00,00,01,01,00,00,00,00,00,05,06,00,00,00,01,02,00,00,00,00,00,05,20,00,00,00,20,02,00,00,01,02,00,00,00,00,00,05,20,00,00,00,20,02,00,00"
$hexSddl = $sddlString.Split(',') | ForEach-Object { "0x$_"}

# Here we assign some launch permissions to the AppId. The AppId used here must match the one defined in the source code.
# AppIds are used both for handling launch permissions as well as defining the context in which the dll exists (such as a Windows service). You can only change 
# access permissions via the AppID. This is used by RPCSS under the hood when the dll is called.
New-ItemProperty -Path "HKLM:\Software\CLASSES\Appid\{$appId}" -Name "AccessPermission" -PropertyType Binary -Value ([byte[]]$hexSddl)
New-ItemProperty -Path "HKLM:\Software\CLASSES\Appid\{$appId}" -Name "LaunchPermission" -PropertyType Binary -Value ([byte[]]$hexSddl)
New-ItemProperty -Path "HKLM:\Software\CLASSES\Appid\{$appId}" -Name "DllSurrogate" -Value ""

# 5. Here, we register a CLSID for the dll. We specify the AppId created previously as a property of the CLSID, as well as point to the location where the dll is stored. 
# More information on CLISD can be found here: https://learn.microsoft.com/en-us/windows/win32/com/clsid-key-hklm
New-item -path "HKLM:\SOFTWARE\CLASSES\CLSID\{$CLSID}"
New-ItemProperty -path "HKLM:\SOFTWARE\CLASSES\CLSID\{$CLSID}" -Name "AppID" -Value "{$appId}"
New-item -path "HKLM:\SOFTWARE\CLASSES\CLSID\{$CLSID}\InprocServer32" -Value $dllFileLocationEscaped

# 6. Here we define the thread access rules of the COM object. In this case we want the dll to run in the same 'apartment' as the client process. 
# Apartments are a way to controll access to the process from specific threads, read more here: https://learn.microsoft.com/en-us/windows/win32/com/processes--threads--and-apartments#the-apartment-and-the-com-threading-architecture
New-ItemProperty -path "HKLM:\SOFTWARE\CLASSES\CLSID\{$CLSID}\InprocServer32" -Name "ThreadingModel" -Value "Both"

# 7. set owner of key to current user
if(enable-privilege SeTakeOwnershipPrivilege){
	Write-Host "Enabled SeTakeOwnershipPrivilege privilege"
}
else{
	Write-Host "Enabling SeTakeOwnershipPrivilege privilege failed"
}

# The current user owns the COM class.
# here we use the .NET registry class to open a COM class subkey 
# This key won't exist if the containers feature isn't enabled 
$key = [Microsoft.Win32.Registry]::LocalMachine.OpenSubKey("SYSTEM\CurrentControlSet\Control\CCG\COMClasses",[Microsoft.Win32.RegistryKeyPermissionCheck]::ReadWriteSubTree,[System.Security.AccessControl.RegistryRights]::takeownership)
$acl = $key.GetAccessControl()
$originalOwner = $acl.owner
$user = whoami 
$me = [System.Security.Principal.NTAccount]$user
$acl.SetOwner($me)
$key.SetAccessControl($acl)

#Add new access rule that gives full control for current user.
$acl = $key.GetAccessControl()
$idRef = [System.Security.Principal.NTAccount]($user)
$regRights = [System.Security.AccessControl.RegistryRights]::FullControl
$inhFlags = [System.Security.AccessControl.InheritanceFlags]::ContainerInherit
$prFlags = [System.Security.AccessControl.PropagationFlags]::None
$acType = [System.Security.AccessControl.AccessControlType]::Allow
$rule = New-Object System.Security.AccessControl.RegistryAccessRule($idRef, $regRights, $inhFlags, $prFlags, $acType)
$acl.AddAccessRule($rule)
$key.SetAccessControl($acl)

# 8. Register the COM Class under the CLSID defined earlier 
New-item -path  "HKLM:\SYSTEM\CurrentControlSet\Control\CCG\COMClasses\{$CLSID}" -Value ""

# 9. Set owner back to original owner and remove access rule for current user. 
$acl = $key.GetAccessControl()
$acl.RemoveAccessRule($rule)
$acl.SetOwner([System.Security.Principal.NTAccount]$originalowner)
if (enable-privilege SeRestorePrivilege){
	Write-Host "Enabled SeRestorePrivilege privilege"
}
else{
	Write-Host "Enabling SeRestorePrivilege privilege failed"
}
$key.SetAccessControl($acl)
$key.close()

#Disable privileges now that the reigsteration is done. 
if (enable-privilege SeRestorePrivilege -disable){
	Write-Host "Disabled SeRestorePrivilege privilege"
}
else{
	Write-Host "Disabling SeRestorePrivilege privilege failed"
}

if (enable-privilege SeTakeOwnershipPrivilege -disable){
	Write-Host "Disabled SeTakeOwnershipPrivilege privilege"	
}
else{
	Write-Host "Disabling SeTakeOwnershipPrivilege privilege failed"
}