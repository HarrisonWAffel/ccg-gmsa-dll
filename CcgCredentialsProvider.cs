using System;
using System.EnterpriseServices;
using System.Runtime.InteropServices;
using System.Net.Http;
using System.IO;
using System.Net;
using System.Diagnostics;
using System.Collections.Generic;
using System.Web.Script.Serialization;

namespace rancher.gmsa
{

    // TODO; env vars. How can we deploy this DLL in 'dev mode' so that we can more easily debug and 
    // assess issues? We could probably get away with using the registry for simple flags. 
    // We would want: Disable SSL/mTLS certs, log to server (?), 

    [Guid("6ECDA518-2010-4437-8BC3-46E752B7B172")]
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

    [Guid("e4781092-f116-4b79-b55e-28eb6a224e26")]
    [ProgId("CcgCredProvider")]
    public class CcgCredProvider : ServicedComponent, ICcgDomainAuthCredentials
    {
        // logger is our Event Logger. We log to the Application source, not a custom source
        // this allows us to circumvent privileged operations required to setup a new source
        private EventLog logger;
        public CcgCredProvider()
        {
            logger = new EventLog("Application");
            logger.Source = "Application";
        }
        
        private void LogInfo(string log)
        {
            logger.WriteEntry(log, EventLogEntryType.Information, 101, 1);
        }

        private void LogWarn(string log)
        {
            logger.WriteEntry(log, EventLogEntryType.Warning, 201, 1);
        }

        private void LogError(string log)
        {
            logger.WriteEntry(log, EventLogEntryType.Error, 301, 1);
        }

        public void GetPasswordCredentials(
            [MarshalAs(UnmanagedType.LPWStr), In] string pluginInput,
            [MarshalAs(UnmanagedType.LPWStr)] out string domainName,
            [MarshalAs(UnmanagedType.LPWStr)] out string username,
            [MarshalAs(UnmanagedType.LPWStr)] out string password)
        {
            try
            {
               GetCredential(DecodeInput(pluginInput));
            }
            catch (Exception e)
            {
                // log the exception our self 
                // so we know we can find it
                LogError(e.ToString());
                // throw it again so ccg can catch it
                // and print its own error logs
                throw e;
            }

            domainName = "test.com";
            username = "user1";
            password = "pass1";

            LogInfo("we exited from the dll");
        }

        public void GetCredential(PluginInput pluginInput)
        {
            // disable SSL checks for development 
            ServicePointManager.ServerCertificateValidationCallback += (sender, cert, chain, sslPolicyErrors) => true;

            var secretUri = "https://haffel-webhook.suse.ngrok.io";
            var httpClient = new HttpClient();
            LogInfo("we created an http client");
            try
            {              
                var content = new StringContent(pluginInput.ActiveDirectory + " and " + pluginInput.SecretName + " with port " + pluginInput.Port);
                LogInfo("making request, content is " + content.ToString());
                var response = httpClient.PostAsync(secretUri, content);
            }
            catch (Exception ex)
            {
                LogError("Http Client Hit An Exception: \n " + ex.ToString());
            }
        }

        public PluginInput DecodeInput(string pluginInput)
        {
            return new PluginInput(pluginInput);
        }

        public class PluginInput
        {
            // test
            public PluginInput(bool isJson, string json) {
                input = new JavaScriptSerializer().Deserialize<Dictionary<string, string>>(json);
            }
            public Dictionary<string, string> input;

            public PluginInput(string pluginInput)
            {
                var parts = pluginInput.Split(':');
                if (parts.Length != 2) {
                    throw new Exception("Invalid Plugin Input Format");
                }
                this.ActiveDirectory = parts[0];
                this.SecretName = parts[1];
                this.Port = GetPort(pluginInput);
            }

            public string ActiveDirectory { get; set; }
            public string SecretName { get; set; }
            public string Port { get; set; }

            public string GetPort(string pluginInput)
            {
               string subDirFile = "/var/lib/rancher/gmsa/" + this.ActiveDirectory + "/port.txt";
               string text = "";
               try { 
                    text = File.ReadAllText(subDirFile);
               } catch (Exception e) {
                    throw new Exception("Failed to open port file");
               }
               return text;
            }
        }
    }
}

