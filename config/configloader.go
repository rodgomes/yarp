package config

import (
	"net/http"
	"net/url"
	"regexp"
	"fmt"
	"encoding/json"
	"io/ioutil"
)

//struct to hold what can de configurable in this proxy
type ProxyConf struct {
	Target *url.URL
	Rule RouterRule
}

func NewRegexProxyConf(conf *RouterConf) *ProxyConf{
	return &ProxyConf{
		Rule:RegexRouterRule{regex: *regexp.MustCompile(conf.PathPattern)},
							 Target: &url.URL{Scheme:conf.Scheme,
							 Host:conf.TargetUrl},
	}
}

func NewSimpleProxyConf(conf *RouterConf) *ProxyConf{
	return &ProxyConf{
		Rule:DefaultRouterRule{},
		Target: &url.URL{Scheme:conf.Scheme, Host:conf.TargetUrl},
	}
}

func NewProxyConf(conf *RouterConf) *ProxyConf {
	if conf.PathPattern != "" {
		return NewRegexProxyConf(conf)
	} else {
		return NewSimpleProxyConf(conf)
	}
}

//just define a simple method to choose a target
type RouterRule interface {
	Matches(r *http.Request) bool
}

//default router. Always sends incoming traffic to target
type DefaultRouterRule struct {

}

func (r DefaultRouterRule) Matches(req *http.Request) bool {
	return true
}

//route to a given target based on a regex
type RegexRouterRule struct {
	regex regexp.Regexp
}

func (r RegexRouterRule) Matches(req *http.Request) bool {
	return r.regex.MatchString(req.URL.Path)
}

//json<-> config mapping
//set of structs to hold config info
type settings struct {
	Port string `json:port`
	Routers []RouterConf `json:routers`
}

//this represents a single configuration, which maps this proxy to a target url
//this configuration can or cannot contain a pattern.
type RouterConf struct {
	TargetUrl string `json:targetUrl`
	Scheme string `json:scheme`
	PathPattern string `json:pathPattern`
}

func LoadSettings() (*string, [] *ProxyConf, error){
	var targets[] *ProxyConf
	var port *string
	file, err := ioutil.ReadFile("./config.json")
	if err != nil {
		fmt.Printf("could not read configuraton file. Error: %v\n", err)
		return port, targets, err
	}

	var rawSettings settings
	err = json.Unmarshal(file, &rawSettings)
	if err != nil {
		fmt.Println("Could not parse configuration file. ", err)
		return port, targets, err
	}

	port = &rawSettings.Port

	for i:=0; i<len(rawSettings.Routers);i++ {
		targets = append(targets, NewProxyConf(&rawSettings.Routers[i]))
	}

	return port, targets, nil

}