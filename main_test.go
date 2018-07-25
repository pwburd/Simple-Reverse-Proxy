package main

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func init() {
}

func newConfig() Config {
	testConfig := Config{}
	testConfig.StaticDirUrlRoot = "/test"
	testConfig.StaticDirRoot = "./test"
	testConfig.RegexStrFindReplace = make(map[string]string)
	testConfig.RegexStrFindReplace["A"] = "a"
	testConfig.RegexStrFindReplace["B"] = "b"
	testConfig.RegexStrFindReplace["C"] = "c"
	testConfig.RegexStrFindReplace["D"] = "d"
	testConfig.RegexStrFindReplace["E"] = "e"
	testConfig.RegexStrFindReplace["F"] = "f"
	testConfig.RegexStrFindReplace["p(x*)q"] = "T"
	testConfig.RegexFindReplace = convertToRegex(testConfig.RegexStrFindReplace)
	return testConfig
}

func TestConfigReader(t *testing.T) {
	// TODO: test that config reader is doing it's thing
}

func TestFileServer(t *testing.T) {
	// TODO: test that files are being returned
}

// Given a proxied response from a backend, the response body should have a 'find-and-replace' applied on it
// given the key/values of config.FindAndReplace
func TestReplaceSearch(t *testing.T) {
	testConfig := newConfig()
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("A B C D E F -pq-pxxq-"))
	}))
	defer backend.Close()

	frontend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		backURL, _ := url.Parse(backend.URL)
		testConfig.ProxyHost = backURL.Host
		wlp := NewSimpleReverseProxy(testConfig)
		wlp.ServeHTTP(w, r)
	}))
	defer frontend.Close()

	getReq, _ := http.NewRequest("GET", frontend.URL, nil)
	getReq.Host = frontend.URL
	getReq.Close = true

	res, err := http.DefaultClient.Do(getReq)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if g, e := res.StatusCode, http.StatusOK; g != e {
		t.Errorf("got res.StatusCode %d; expected %d", g, e)
	}
	bodyBytes, _ := ioutil.ReadAll(res.Body)
	if string(bodyBytes) != "a b c d e f -T-T-" {
		t.Errorf("get %s expected %s", string(bodyBytes), "a b c d e f -T-T-")
	}
}
