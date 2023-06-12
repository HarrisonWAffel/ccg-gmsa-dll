
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
// on Windows.
using System.EnterpriseServices;

namespace gmsaPlugin
{
    // a COM component must explicitly define an interface
    // as well as an interface implementation. The interface
    // is known by callers of the component, only the component
    // is aware of the implementation. All COM components inherit
    // the IUnknown Interafece (Specified in the Interface Type),
    // as well as register a Guid which is used to identify the
    // component within Windows. 
    [Guid("24DC734A-E2D4-4F12-A387-F6209CBAF4FC")]
    [InterfaceType(ComInterfaceType.InterfaceIsIUnknown)]
    [ComImport]
    public interface gMSACredentialGetter
    {
        // call details should match  
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

        public async Task<GmsaOperatorResponse?> RequestOperator(HttpClient client, String input)
        {
            var jsonStream = await client.GetStringAsync("https://localhost:7586");

            var response = System.Text.Json.JsonSerializer.Deserialize<GmsaOperatorResponse>(jsonStream);

            return response;
        }
    }
}

