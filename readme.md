## Configure Spinnaker

+ Update your `gate.local.yml` to use "${azureTenantId}/oauth2" endpoints (dont use the older ones)
+ Change the `scope:` to use a "custom azure scope" ( i will come back to this ) to ensure that the access_token returns the correct [aud] "audience"
+ Spinnaker internally adds other fields in the url which is why we cannot add scopes to the "Authorize" endpoint
+ The ORDER of the scopes is important if you ask for email for example before api://spin you will get a token with an AUD [00000003-0000-0000-c000-000000000000] which is a resource identifier for https://graph.microsoft.com and therefore NOT you app so you wont get a users roles or groups (if you added them) (see https://www.shawntabrizi.com/aad/common-microsoft-resources-azure-active-directory/
+ Once we have a valid access_token... which has all roles and groups inside it for a user acct.. Spinnaker will complain because its hard wired to always call the [userInfoUri] ( which by default if you use the out of thebox settings is the microsoft graphApi ) and the token we just received  is not for the graph api... its for our appliation ( audience is api://???).. This is fustrating as the graph api call to get user info is actually not needed as we have all we need inside the token - be cool if the devs could add a field "getinfofromToken" or equiv ?
+ On the positive side everything we need is inside the token anyways .. so all we need to do now is NOT call the graphapi via userInfoUri and instead just call our own endpoint; which in my case simply validates/extracts the access_token and returns the field mappings ...
+ The rest is AD based and I can supply screenshots to show you how its set it :slightly_smiling_face:
+ See repos for gate-local.yml .. and a super quick easy service "http://getuserinfo-svc:8008/getuserinfo" that I knocked up quickly in GO to demo it..  deploy the svc locally - it just extracts/validates and returns json for mapper

## Important that you include the `username` mapping
* DONT forget that [`username`: YourField...] HAS to be in the mapping.. It must be mapped behind the scenes normally but when overriding with gate-local.yml if you dont specify it you with get errors around "user_id" missing... This was annoying and confusing as its called "username" in code but logged as "user_id" - so had to trawl source to find it

## GetUserInfo server..
+ See the authz-getuserinfo folder


## To generate a test flow...

+ Configure your AD callback to point to a server where u can grab a response eg.. http://httbin/2Fanything or localhost
+ I am using docker to run it locally ...docker run -d -p 8085:80 kennethreitz/httpbin
+ Chuck in a url like this. I have a few extra scopes in for fun but u wont need them :)

```
https://login.microsoftonline.com/{tenantid}/oauth2/v2.0/authorize?scope=profile%20openid%20email%20api://spin/GetRoles%20Files.Read.All&client_id={clientid}2&redirect_uri=http%3A%2F%2Flocalhost%3A8085%2Fanything&response_type=code&state=blob
This will give u a response like - You just need the code... its a one time use :)
{
  "args": {
    "code": "0.AUgAiuGw_fNh4kmJH84t75h7WUpJDwvDu1ZBjLvX_3D7plJIAAw.AQABAAIAAAD--DLA3VO7QrddgJg7WevrYzUX5uOyhU6SUEdP2CtY-IlW1zfWKPDk3z21q-Rj3PvBEAiaTo5bHleYHkgudCXIm97R_gR2KmRY86C57w_xdSbRR9ecXh_J-6cfp-rb9Uos8AVwalbrMC1QuZb9kMhUXypuOvm5cm-0mOH4RBQYlA8ANNJBXMOUnPNan3E2...one time only with this :)",
    "session_state": "f9535f09-edb9-4d06-9758-b86cd4058708",
    "state": "H6E1IW"
  },
  "data": "",
  "files": {},
  "form": {},
  "headers": {
    "Sec-Fetch-Mode": "navigate",
    "Sec-Fetch-Site": "none",
  },
  "json": null,
  "method": "GET",
  "origin": "172.17.0.1",
  "url": "abridged.."
}
```

+ Paste the above ^^ into a file called "code" and run this `token.sh` (update the settings for clientId first) script to consume the code and generate a nice access_token


## Azure Setup

+ Add some app roles
![image](https://user-images.githubusercontent.com/2591162/121576231-15e49c00-ca20-11eb-9671-dff9714c4ca4.png)
+ Allocate the roles to either a user or group ( see google ) but essentially if you click here you will end up in right place
![image](https://user-images.githubusercontent.com/2591162/121576672-825f9b00-ca20-11eb-9dcb-cc5fa9d5cf06.png)
+ Expose an Api
![image](https://user-images.githubusercontent.com/2591162/121576767-a1f6c380-ca20-11eb-8ce7-b91763ed9e96.png)




