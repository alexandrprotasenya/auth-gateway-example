package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/spf13/viper"
)

type Config struct {
	Port           string
	ProxyTarget    string
	ProxyPath      string
	AuthTarget     string
	AuthHeaderName string
	UserDataHeader string
}

var appConfig Config

func init() {
	viper.SetEnvPrefix("GATEWAY")
	viper.BindEnv("PORT")
	viper.BindEnv("PROXY_TARGET")
	viper.BindEnv("PROXY_PATH")
	viper.BindEnv("AUTH_TARGET")
	viper.BindEnv("AUTH_HEADER_NAME")
	viper.BindEnv("USER_DATA_HEADER")

	viper.SetDefault("PORT", "80")
	viper.SetDefault("PROXY_TARGET", "")
	viper.SetDefault("PROXY_PATH", "/")
	viper.SetDefault("AUTH_TARGET", "")
	viper.SetDefault("AUTH_HEADER_NAME", "Authorization")
	viper.SetDefault("USER_DATA_HEADER", "X-User-Data")

	appConfig.Port = viper.GetString("PORT")
	appConfig.ProxyTarget = viper.GetString("PROXY_TARGET")
	appConfig.ProxyPath = viper.GetString("PROXY_PATH")
	appConfig.AuthTarget = viper.GetString("AUTH_TARGET")
	appConfig.AuthHeaderName = viper.GetString("AUTH_HEADER_NAME")
	appConfig.UserDataHeader = viper.GetString("USER_DATA_HEADER")
}

func serveReverseProxy(proxyResponse http.ResponseWriter, proxyRequest *http.Request) {
	log.Println(formatRequest(proxyRequest))

	txn := newrelic.FromContext(proxyRequest.Context())
	authClient := &http.Client{
		Timeout: time.Second * 10,
	}
	authClient.Transport = newrelic.NewRoundTripper(authClient.Transport)
	authRequest, err := http.NewRequest("GET", appConfig.AuthTarget, nil)
	if err != nil {
		return
	}
	authRequest.Header.Add("user-agent", "rps/auth-gateway")
	authRequest.Header.Add(appConfig.AuthHeaderName, proxyRequest.Header.Get(appConfig.AuthHeaderName))

	authRequest = newrelic.RequestWithTransactionContext(authRequest, txn)

	authResponse, err := authClient.Do(authRequest)
	if err != nil {
		return
	}
	defer authResponse.Body.Close()

	if authResponse.StatusCode != http.StatusOK {
		proxyResponse.WriteHeader(http.StatusForbidden)
		proxyResponse.Write([]byte("Forbidden"))
	} else {
		proxyRequest.Header.Add(appConfig.UserDataHeader, authResponse.Header.Get(appConfig.UserDataHeader))
		url, _ := url.Parse(appConfig.ProxyTarget)
		proxy := httputil.NewSingleHostReverseProxy(url)
		proxy.Transport = newrelic.NewRoundTripper(proxy.Transport)
		proxy.ServeHTTP(proxyResponse, proxyRequest)
	}
}

func health(res http.ResponseWriter, req *http.Request) {
	res.Write([]byte("{\"status\": \"ok\"}"))
}

func main() {
	app, err := newrelic.NewApplication(newrelic.ConfigFromEnvironment())
	if err != nil {
		log.Printf("newrelic error: %s\n", err)
	}

	http.HandleFunc(newrelic.WrapHandleFunc(app, appConfig.ProxyPath, serveReverseProxy))
	http.HandleFunc(newrelic.WrapHandleFunc(app, "/health", health))
	log.Printf("Server started on 0.0.0.0:%s", appConfig.Port)
	log.Fatal(http.ListenAndServe(":"+appConfig.Port, nil))
}

func formatRequest(r *http.Request) string {
	var request []string
	url := fmt.Sprintf("%v %v %v", r.Method, r.URL, r.Proto)
	request = append(request, url)
	return strings.Join(request, "\n")
}
