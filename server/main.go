package main

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"fmt"
	"log"
	"strings"
	"regexp"
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

func main() {

	//TODO read conf from external file
	proxy := NewMultiHostReverseProxy([]*ProxyConf{
		&ProxyConf{
			rule:RegexRouterRule{regex: *regexp.MustCompile("/accounts/tl*")},
			target: &url.URL{Scheme:"http", Host:"localhost:8001"},
		},
		&ProxyConf{
			rule:DefaultRouterRule{},
			target: &url.URL{Scheme:"http", Host:"localhost:8000"},
		},
	})

	fmt.Println("Starting server on port 9090")
	log.Fatal(http.ListenAndServe(":9090", proxy))
}