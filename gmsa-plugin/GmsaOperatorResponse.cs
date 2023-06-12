using System;

/* this file is used to deserialize responses from the
   gMSA operator.We expect the following structure to be
   returned from the operator

	{
		"username": "a username",
		"password": "some password",
		"domainName": "some domain name"
	}

	This format matches the expected return values in the COM component interface
*/
namespace gmsaPlugin
{
	public class GmsaOperatorResponse
	{
		public string username { get; set; }
		public string password { get; set; }
		public string domainName { get; set; }
	}
}