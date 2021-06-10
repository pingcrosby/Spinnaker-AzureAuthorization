From the spinnaker side.. this is high level .. overview

+ update your gate.local.yml to use "${azureTenantId}/oauth2" endpoints (dont use the older ones)
+ change the "scope:" to use a "custom azure scope" ( i will come back to this ) to ensure that the access_token returns the correct [aud] "audience"
+ Spinnaker internally adds other fields in the url which is why we cannot add scopes to the "Authorize" endpoint
+ The ORDER of the scopes is important if you ask for email for example before api://spin you will get a token with an AUD [00000003-0000-0000-c000-000000000000] which is a Resource identifier for https://graph.microsoft.com and therefore NOT you app so you wont get a users roles or groups (if you added them) (see https://www.shawntabrizi.com/aad/common-microsoft-resources-azure-active-directory/
+ Once we have a valid access_token... which has all roles and groups inside it for a user acct.. Spinnaker will complain because its hard wired to always call the [userInfoUri] ( which by default is the microsoft graphApi ) and the token we just received  is not for the graph api... its for our appliation ( audience is api://???).. This is fustrating as that extra call is actually not needed as we have all we need inside the token - be cool if the devs could add a field "getinfofromToken" or equiv ?
+ On the positive side everything we need is inside the token anyways .. so all we need to do now is NOT call the graphapi via userInfoUri and instead just call our own endpoint; which in my case simply validates/extracts the access_token and returns the field mappings ...
+ The rest is AD based and I can supply screenshots to show you how its set it :slightly_smiling_face:
+ See below for gate-local.yml .. and a super quick easy service "http://getuserinfo-svc:8008/getuserinfo" that I knocked up quickly in GO to demo it..  deploy the svc locally - it just extracts/validates and returns json for mapper
