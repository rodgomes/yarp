package main

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"fmt"
	"log"
	"strings"
	"regexp"
	"os"
	"encoding/json"
	"io/ioutil"
)

//struct to hold what can de configurable in this proxy
type ProxyConf struct {
	target *url.URL
	rule RouterRule
}

//just define a simple method to choose a target
type RouterRule interface {
	matches(r *http.Request) bool
}

//default router. Always sends incoming traffic to target
type DefaultRouterRule struct {

}

func (r DefaultRouterRule) matches(req *http.Request) bool {
	return true
}

//route to a given target based on a regex
type RegexRouterRule struct {
	regex regexp.Regexp
}

func (r RegexRouterRule) matches(req *http.Request) bool {
	return r.regex.MatchString(req.URL.Path)
}

type settings struct {
	Port string `json:port`
	Routers []RouterConf `json:routers`
}

type RouterConf struct {
	TargetUrl string `json:targetUrl`
	Scheme string `json:scheme`
	PathPattern string `json:pathPattern`
}

//just copied from reverseproxy.go
func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}

func NewMultiHostReverseProxy(targets []*ProxyConf) *httputil.ReverseProxy {

	director := func(req *http.Request) {

		//here I need a target that matches given income request
		var target *url.URL = nil
		for i := 0; i < len(targets); i++ {
			if targets[i].rule.matches(req) {
				target = targets[i].target
				break
			}
		}
		fmt.Println("Proxing to target ", target, " path ", req.URL.Path)

		targetQuery := target.RawQuery
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.URL.Path = singleJoiningSlash(target.Path, req.URL.Path)
		if targetQuery == "" || req.URL.RawQuery == "" {
			req.URL.RawQuery = targetQuery + req.URL.RawQuery
		} else {
			req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
		}
	}

	return &httputil.ReverseProxy{Director: director}
}

var (
	port *string
	targets []*ProxyConf
)

func loadSettings(){
	file, e := ioutil.ReadFile("./config.json")
	if e != nil {
		fmt.Printf("could not read configuraton file. Error: %v\n", e)
		os.Exit(1)
	}

	var rawSettings settings
	err := json.Unmarshal(file, &rawSettings)
	if err != nil {
		fmt.Println("Could not parse configuration file. ", err)
		os.Exit(1)
	}
	fmt.Println(rawSettings)
	fmt.Println("Starting app using port ", rawSettings.Port)
	port = &rawSettings.Port

	for i:=0; i<len(rawSettings.Routers);i++ {

		if rawSettings.Routers[i].PathPattern != "" {
			targets = append(targets, &ProxyConf{
				rule:RegexRouterRule{regex: *regexp.MustCompile(rawSettings.Routers[i].PathPattern)},
				target: &url.URL{Scheme:rawSettings.Routers[i].Scheme,
					Host:rawSettings.Routers[i].TargetUrl},
			})
		} else {
			targets = append(targets, &ProxyConf{
				rule:DefaultRouterRule{},
				target: &url.URL{Scheme:rawSettings.Routers[i].Scheme, Host:rawSettings.Routers[i].TargetUrl},
			},)
		}
	}

}

func main() {

	//TODO read conf from external file
//	proxy := NewMultiHostReverseProxy([]*ProxyConf{
//		&ProxyConf{
//			rule:RegexRouterRule{regex: *regexp.MustCompile("/accounts/tl*")},
//			target: &url.URL{Scheme:"http", Host:"localhost:8001"},
//		},
//		&ProxyConf{
//			rule:DefaultRouterRule{},
//			target: &url.URL{Scheme:"http", Host:"localhost:8000"},
//		},
//	})
	loadSettings()
	proxy := NewMultiHostReverseProxy(targets)
	//fmt.Println("Starting server on port 9090")
	log.Fatal(http.ListenAndServe(":" + *port, proxy))
}