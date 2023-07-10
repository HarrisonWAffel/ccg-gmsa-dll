using System;
using System.EnterpriseServices;
using System.Runtime.InteropServices;

namespace rancher_gmsa
{

    /*
        This file implements the ICcgDomainAuthCredentials interface.
        The class must use the ServicedComponent base class, provided
        by the System.EnterpriseServices library. 

        The Guid attribute added to the class must equal the CLSID used 
        during installation, and must be lower case.
        
        The ProgId attribute can by anything, and is just a plain text name.
     */


    [Guid("e4781092-f116-4b79-b55e-28eb6a224e26")]
    [ProgId("RanchergMSACredentialProvider")]
    public class RanchergMSACredentialProvider : ServicedComponent, ICcgDomainAuthCredentials
    {

     
        public RanchergMSACredentialProvider()
        {
        }

        // This function implements the ICcgDomainAuthCredentials interface.
        // All of the parameter attributes must match those in the supporting
        // interface. 
        public void GetPasswordCredentials(
            [MarshalAs(UnmanagedType.LPWStr), In] string pluginInput,
            [MarshalAs(UnmanagedType.LPWStr)] out string domainName,
            [MarshalAs(UnmanagedType.LPWStr)] out string username,
            [MarshalAs(UnmanagedType.LPWStr)] out string password)
        {
            password = "test";
            username = "test";
            domainName = "test";
        }
    }
}
  
