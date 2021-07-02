package main

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/form3tech-oss/jwt-go"
	"github.com/spf13/viper"
)

var rsakeys map[string]*rsa.PublicKey

// just static copies of the config elements - they dont change
var issuer string
var aud string
var appid string
var claims2spin map[string]interface{}

func LoadConfig() {
	configPath, found := os.LookupEnv("CONFIGPATH")
	if found != true || configPath == "" {
		log.Fatal("Env [CONFIGPATH] not found or empty")
	}

	log.Infof("Loading configuration from %s", configPath)

	viper.SetConfigFile(configPath)
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		log.Fatalf("Fatal error config file: %s (check your env CONFIGPATH)", err)
	}
	// mandatory fields that must be in config setting
	ValidateConfig("jwt.issuer")
	ValidateConfig("jwt.jwksUri")
	ValidateConfig("jwt.aud")
	ValidateConfig("jwt.appid")
	ValidateConfig("claim2spin")
	// sadly rest of claims are udf
	issuer = viper.GetString("jwt.issuer")
	aud = viper.GetString("jwt.aud")
	appid = viper.GetString("jwt.appid")
	claims2spin = viper.GetStringMap("claim2spin")
	// set up the log level
	SetLogLevel(viper.GetString("server.log"))
}

func SetLogLevel(loglevel string) {
	switch loglevel {
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "warn":
		log.SetLevel(log.WarnLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	case "trace":
		log.SetLevel(log.TraceLevel)
	default:
		log.SetLevel(log.InfoLevel)
		log.Warnf("No loglevel supplied (defaults to info) options are debug, info, warn, error, trace")
	}
	log.Infof("Logging set to %s", log.GetLevel())
}

func GetPublicKeys() {
	jwksUri := viper.GetString("jwt.jwksUri")

	// gets and caches the public key for our tenant
	rsakeys = make(map[string]*rsa.PublicKey)
	var body map[string]interface{}
	resp, err := http.Get(jwksUri)
	if err != nil {
		log.Fatalf("Error with [jwt.jwksUri] %s", err)
	}

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
	log.Infof("Successfully cached signature validation) keys from %s", jwksUri)
}

func ValidateConfig(field string) {
	if !viper.IsSet(field) {
		log.Fatalf("Mandatory field [%s] missing from configuraton", field)
	}
}

func hdlr_healthzprobe(w http.ResponseWriter, req *http.Request) {
	// really v little to do here as the code does so little
	w.WriteHeader(http.StatusOK)
}

func hdlr_userinfo(w http.ResponseWriter, req *http.Request) {

	log.Debug("Request received")

	errorMessage := ""

	tokenString := req.Header.Get("Authorization")
	if !strings.HasPrefix(tokenString, "Bearer ") {
		log.Debug("Unauthorized request missing header 'Authorization: Bearer {tok}'")
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Unauthorized"))
		return
	}

	tokenString = strings.TrimPrefix(tokenString, "Bearer ")

	log.Tracef("Token [access_token] %s", tokenString)

	// parse and validate against sig, exp etc
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if token.Header["kid"] == nil {
			return nil, fmt.Errorf("Invalid token supplied, missing [kid]")
		}
		return rsakeys[token.Header["kid"].(string)], nil // grab the appropriate key to check it against
	})

	if err != nil {
		// dont inform the calling app as too much info could be exposed but happy to console log it
		log.Debug(err.Error())
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// token is now valid regarding exp, and signature - check if it hae correct data
	if !token.Valid {
		errorMessage = "Invalid token"
	} else if token.Header["alg"] == nil {
		errorMessage = "Invlalid [alg] - it must be defined"
	} else if !token.Claims.(jwt.MapClaims).VerifyAudience(aud, true) {
		errorMessage = fmt.Sprintf("Invalid [aud] - expected [%s]", aud)
	} else if !strings.Contains(token.Claims.(jwt.MapClaims)["iss"].(string), issuer) {
		errorMessage = fmt.Sprintf("Invalid [iss] - expected [%s]", issuer)
	} else if token.Claims.(jwt.MapClaims)["appid"].(string) != appid {
		errorMessage = fmt.Sprintf("Invalid [appid] - expected [%s]", appid)
	} else {
		// copy over the values from a claim for us to return using mapping in config
		jsonout := make(map[string]interface{})
		for k, v := range claims2spin {
			jsonout[v.(string)] = token.Claims.(jwt.MapClaims)[k]
		}
		log.Debugf("Mapped claims to %s", jsonout)
		// now return it
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(jsonout)

		return // we end here... when valid
	}

	// something was wrong with the token
	log.Debug(errorMessage)
	w.WriteHeader(http.StatusForbidden)
	w.Write([]byte(errorMessage))
}

func hdlr_notfound(w http.ResponseWriter, req *http.Request) {
	log.Debug("NotFoundHandler invoked")

	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("Not found.. please try /userinfo (you will need an auth token)"))
}

func main() {
	log.SetOutput(os.Stdout) // Output to stdout instead of the default stderr
	log.SetLevel(log.DebugLevel)

	log.SetFormatter(&log.TextFormatter{
		DisableColors:    false,
		DisableTimestamp: true,
	})

	LoadConfig()
	GetPublicKeys()

	http.HandleFunc("/", hdlr_notfound)
	http.HandleFunc("/userinfo", hdlr_userinfo)
	http.HandleFunc("/healthz/liveness", hdlr_healthzprobe)
	http.HandleFunc("/healthz/readiness", hdlr_healthzprobe)
	http.HandleFunc("/healthz/startup", hdlr_healthzprobe)

	// use port 8080 unless set in config -- annoying we dont have tenary in Go :(
	port := 8080
	if viper.IsSet("server.port") {
		port = viper.GetInt("server.port")
	}
	address := fmt.Sprintf(":%d", port)

	log.Info("Starting server on address", address)
	log.Infof("route --> http://%s/userinfo", address)
	log.Infof("route --> http://%s/healthz/liveness", address)
	log.Infof("route --> http://%s/healthz/readiness", address)
	log.Infof("route --> http://%s/healthz/startup", address)

	err := http.ListenAndServe(address, nil)
	if err != nil {
		panic(err)
	}
}
