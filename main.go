package main

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"fmt"
	"log"
	"strings"
	"github.com/rodgomes/yarp/config"
	"os"
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

func main() {

	port, targets, e := config.LoadSettings()
	if e != nil {
		fmt.Println("Could not start server. Error ", e)
		os.Exit(1)
	}
	proxy := NewMultiHostReverseProxy(targets)
	fmt.Println("Starting app using port ", *port)
	log.Fatal(http.ListenAndServe(":" + *port, proxy))
}