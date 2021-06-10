package main
import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"strings"
	"github.com/form3tech-oss/jwt-go"
)
type UserInfo struct {
	GivenName         string        `json:"givenName"`
	Surname           string        `json:"surname"`
	UserPrincipalName string        `json:"userPrincipalName"`
	Roles             []interface{} `json:"roles"`
}
var rsakeys map[string]*rsa.PublicKey
// jwks_issuer
var iss = "https://sts.windows.net/TENANT-HERE"
// the public discovery key url
var discovery_keys = "https://login.microsoftonline.com/TENANT-HERE/discovery/keys"
func GetPublicKeys() {
	// gets and caches the public key for our tenant
	rsakeys = make(map[string]*rsa.PublicKey)
	var body map[string]interface{}
	resp, _ := http.Get(discovery_keys)
	json.NewDecoder(resp.Body).Decode(&body)
	for _, bodykey := range body["keys"].([]interface{}) {
		key := bodykey.(map[string]interface{})
		kid := key["kid"].(string)
		rsakey := new(rsa.PublicKey)
		number, _ := base64.RawURLEncoding.DecodeString(key["n"].(string))
		rsakey.N = new(big.Int).SetBytes(number)
		rsakey.E = 65537
		rsakeys[kid] = rsakey
	}
	fmt.Printf("Loaded public (sig validation) keys from %s\n", discovery_keys)
}
func handler(w http.ResponseWriter, req *http.Request) {
	ui := UserInfo{}
	errorMessage := ""
	tokenString := req.Header.Get("Authorization")
	if strings.HasPrefix(tokenString, "Bearer ") {
		tokenString = strings.TrimPrefix(tokenString, "Bearer ")
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return rsakeys[token.Header["kid"].(string)], nil
		})
		if err != nil {
			errorMessage = err.Error()
		} else if !token.Valid {
			errorMessage = "Invalid token"
		} else if token.Header["alg"] == nil {
			errorMessage = "alg must be defined"
		} else if token.Claims.(jwt.MapClaims)["aud"] != "api://spin" {
			errorMessage = "Invalid aud"
		} else if !strings.Contains(token.Claims.(jwt.MapClaims)["iss"].(string), iss) {
			errorMessage = "Invalid iss"
		} else {
			// basically copy over the values from a default claim for us to return
			ui.GivenName = token.Claims.(jwt.MapClaims)["given_name"].(string)
			ui.UserPrincipalName = token.Claims.(jwt.MapClaims)["upn"].(string)
			ui.Surname = token.Claims.(jwt.MapClaims)["family_name"].(string)
			ui.Roles = token.Claims.(jwt.MapClaims)["roles"].([]interface{})
			// now return it
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(ui)
			return // we end here... when valid
		}
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(errorMessage))
	} else {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Unauthorized"))
	}
}
func main() {
	GetPublicKeys()
	http.HandleFunc("/getuserinfo", handler)
	address := ":8008"
	log.Println("Starting server on address", address)
	err := http.ListenAndServe(address, nil)
	if err != nil {
		panic(err)
	}
}
