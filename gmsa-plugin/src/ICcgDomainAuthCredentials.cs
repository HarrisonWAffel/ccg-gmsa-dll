using System;
using System.Runtime.InteropServices;

namespace rancher_gmsa
{
    /*
       This is the interface queried by CCG. It _must_ be named
       'ICcgDomainAuthCredentials'. The Guid attribute indicates the
        AppId of the DLL. All COM classes _must_ implement
        ComInterfaceType.InterfaceIsIUnknown.

        Changing the interface name or type _will_ result in
        CCG errors.
     */


    [Guid("24DC734A-E2D4-4F12-A387-F6209CBAF4FC")]
    [InterfaceType(ComInterfaceType.InterfaceIsIUnknown)]
    [ComImport]
    public interface ICcgDomainAuthCredentials
    {
        void GetPasswordCredentials(
            [MarshalAs(UnmanagedType.LPWStr), In] string pluginInput,
            [MarshalAs(UnmanagedType.LPWStr)] out string domainName,
            [MarshalAs(UnmanagedType.LPWStr)] out string username,
            [MarshalAs(UnmanagedType.LPWStr)] out string password);
    }
}
