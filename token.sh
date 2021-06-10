secret=from azure
ApplicationId=from azure
tenant=from azure

authorization_code=$(cat ./code | jq -r ".args.code" )
resp=$(curl -s -w -X  \
    --header "application/x-www-form-urlencoded" \
    --data-urlencode "response_type=token" \
    --data-urlencode "client_id=$ApplicationId" \
    --data-urlencode "client_secret=$secret" \
    --data-urlencode "grant_type=authorization_code" \
    --data-urlencode "code=$authorization_code" \
    --data-urlencode "scope=openid email api://spin/GetRoles" \
    --data-urlencode "redirect_uri=http://localhost:8085/anything" \
    "https://login.microsoftonline.com/$tenant/oauth2/v2.0/token")
echo $resp
access_token=$(echo $resp | jq -r ".access_token")
# use jq to decode the access token rather than pasting into jwt.io :)
echo $access_token| jq -R 'split(".") | .[0],.[1] | @base64d | fromjson'
# test my get userinfo api with a token as returned by spin
curl -s --header "Authorization: Bearer $access_token"  http://localhost:8008/getuserinfo | jq
