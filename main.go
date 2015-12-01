package main

import (

	"flag"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"log"
	"strings"
	"os"
	"github.com/rodgomes/yarp/config"
)

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

//this is basically a copy of built-in NewSingleHostReverseProxy, with extra logic
//to support multiple targets
//It simply goes through the list and uses the first target that matches
func NewMultiHostReverseProxy(targets []* config.ProxyConf) *httputil.ReverseProxy {

	director := func(req *http.Request) {

		//here I need a target that matches given income request
		var target *url.URL = nil
		for i := 0; i < len(targets); i++ {
			if targets[i].Rule.Matches(req) {
				target = targets[i].Target
				break
			}
		}

		if target != nil {

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
		} else {
			//TODO have to redirect to an error page
			fmt.Println("Could not find target for path ", req.URL.Path)
		}
	}

	return &httputil.ReverseProxy{Director: director}
}

func main() {

	cfgPath := flag.String("c", "./config.json", "config file path; if empty, ./config.json will be used")
	flag.Parse()

	port, targets, e := config.LoadSettings(cfgPath)
	if e != nil {
		fmt.Println("Could not start server. Error ", e)
		os.Exit(1)
	}
	proxy := NewMultiHostReverseProxy(targets)
	fmt.Println("Starting app using port ", *port)
	fmt.Println("Using configuration file ", *cfgPath)
	log.Fatal(http.ListenAndServe(":" + *port, proxy))
}