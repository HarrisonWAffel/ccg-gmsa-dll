
using System.Runtime.InteropServices;
using System.IO;
using System.Runtime.CompilerServices;
using System;
using System.Text.Json;
using System.Text.Json.Serialization;

// System.EnterpriseServices is a package
// which can only be found on Windows OS's.
// Expect IDE errors when editing this file
// on Mac or Linux. Builds will only succeed
// on Windows. If this still can't be resolved
// on Windows, use the Object Explorer to add
// a reference to the library. 
using System.EnterpriseServices;


namespace gmsaPlugin
{
    // a COM component must explicitly define an interface
    // as well as an interface implementation. The interface
    // is known by callers of the component, only the component
    // is aware of the implementation. All COM components inherit
    // the IUnknown Interafece (Specified in the Interface Type),
    // as well as register a Guid which is used to identify the
    // component within the Windows registry.
    [Guid("24DC734A-E2D4-4F12-A387-F6209CBAF4FC")]
    [InterfaceType(ComInterfaceType.InterfaceIsIUnknown)]
    [ComImport]
    public interface gMSACredentialGetter
    {
        // call details should match expected api
        // https://learn.microsoft.com/en-us/windows/win32/api/ccgplugins/nf-ccgplugins-iccgdomainauthcredentials-getpasswordcredentials

        void GetPasswordCredentials(
        [MarshalAs(UnmanagedType.LPWStr), In] string pluginInput,
        [MarshalAs(UnmanagedType.LPWStr)] out string domainName,
        [MarshalAs(UnmanagedType.LPWStr)] out string username,
        [MarshalAs(UnmanagedType.LPWStr)] out string password);
    }

    [ProgId("RancherGmsaCredentialsProvider")]
    [Guid("24DC734A-E2D4-4F12-A387-F6209CBAF4FC")]
    public class Main : ServicedComponent, gMSACredentialGetter {

        public void GetPasswordCredentials(
            [MarshalAs(UnmanagedType.LPWStr), In] string pluginInput,
            [MarshalAs(UnmanagedType.LPWStr)] out string domainName,
            [MarshalAs(UnmanagedType.LPWStr)] out string username,
            [MarshalAs(UnmanagedType.LPWStr)] out string password)
        {
            domainName = "a";
            username = "b";
            password = "c";
            return;
        }
    }
}

