package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"strconv"
	"strings"
	"regexp"
)

var configFile string
var config Config

func init() {
	flag.StringVar(&configFile, "config", "", "required file that has base config values")
}

func main() {
	flag.Parse()
	if configFile == "" {
		log.Fatal("user must supply 'config'")
	}

	// Read config file.
	bytes, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Fatal(err.Error())
	}

	// Unmarshall config to struct
	err = json.Unmarshal(bytes, &config)
	if err != nil {
		log.Fatal("error unmarshalling config:", err.Error())
	}

	// Check that config is valid.
	if !config.OK() {
		log.Fatal("config is not valid")
	}

	config.RegexFindReplace = convertToRegex(config.RegexStrFindReplace)
	log.Println("SimpleReverseProxy serving request on", config.Port)
	http.ListenAndServe(config.Port, NewSimpleReverseProxy(config))
}

// re.ReplaceAllString("base input", "replace pattern")
func convertToRegex(regexStrValue map[string]string) map[*regexp.Regexp]string {
	m := make(map[*regexp.Regexp]string)
	for k, v := range regexStrValue {
		re := regexp.MustCompile(k)
		m[re] = v
	}
	return m
}

// A basic interface to say if contents of struct are valid.
type OKER interface {
	OK() bool
}

// Create a new SimpleReverseProxy given a well formed config.
func NewSimpleReverseProxy(config Config) *SimpleReverseProxy {
	proxyServer := new(SimpleReverseProxy)
	proxyServer.ReverseProxy = NewReverseProxy(config)
	proxyServer.FileServer = http.StripPrefix(config.StaticDirUrlRoot, http.FileServer(http.Dir(config.StaticDirRoot)))
	return proxyServer
}

// Proxy server that will overwrite the response body of the request, also serves static files.
type SimpleReverseProxy struct {
	// Path to static files
	StaticPath string
	// Prefix to static file url i.e '/public'
	FileServerPrefix string
	// Host you would like to proxy request to
	ProxyHost    string
	ReverseProxy *httputil.ReverseProxy
	FileServer   http.Handler
}

// Proxy request or server static content depending on url.
func (p *SimpleReverseProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// `&& config.StaticDirUrlRoot` only is to help testing
	if strings.HasPrefix(r.URL.Path, config.StaticDirUrlRoot) && config.StaticDirUrlRoot != "" {
		p.FileServer.ServeHTTP(w, r)
	} else {
		p.ReverseProxy.ServeHTTP(w, r)
	}
}

// Config obj for creating a new SimpleReverseProxy.
type Config struct {
	// Host that request are proxied to.
	ProxyHost string `json:"proxy-host"`
	// Regex as string read from config file.
	RegexStrFindReplace map[string]string `json:"regex-find-replace"`
	// Compiled version of RegexStrFindReplace.
	RegexFindReplace map[*regexp.Regexp]string
	// Port on which ProxyServer runs.
	Port string `json:"port"`
	// Url root of static content i.e /public
	StaticDirUrlRoot string `json:"static-dir-url-root"`
	// Local dir of static content.
	StaticDirRoot string `json:"static-dir-root"`
}

// Validates that necessary values of a Config are OK. FindReplace map is not a necessary condition.
func (c Config) OK() bool {
	if c.ProxyHost == "" {
		return false
	}
	if c.Port == "" {
		return false
	}
	if c.StaticDirRoot == "" {
		return false
	}
	if c.StaticDirUrlRoot == "" {
		return false
	}
	return true
}

type Transport struct {
	http.RoundTripper
	regexFindReplace map[*regexp.Regexp]string
}

// Read response of proxied request and do a find-and-replace on the body.
func (t *Transport) RoundTrip(req *http.Request) (res *http.Response, err error) {
	res, err = t.RoundTripper.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	err = res.Body.Close()
	if err != nil {
		return nil, err
	}

	s := string(b)
	for re, v := range t.regexFindReplace {
		s = re.ReplaceAllString(s, v)
	}
	b = []byte(s)

	body := ioutil.NopCloser(bytes.NewReader(b))
	res.Body = body
	res.ContentLength = int64(len(b))
	res.Header.Set("Content-Length", strconv.Itoa(len(b)))
	return res, nil
}

// Create ReverseProxy that adds Director and Transport which allow the manipulation of the
// request to the proxy server and the response leaving the proxy server.
func NewReverseProxy(config Config) *httputil.ReverseProxy {
	director := func(req *http.Request) {
		// Manipulate request here.
		req.URL.Scheme = "http"
		req.URL.Host = config.ProxyHost
	}
	reverseProxy := &httputil.ReverseProxy{
		Director: director,
	}
	reverseProxy.Transport = &Transport{
		http.DefaultTransport,
		config.RegexFindReplace,
	}
	return reverseProxy
}
