// Package github implements a simple client to consume gitlab API.
package gogitlab

import (
	"bytes"
	"crypto/tls"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

const (
	dashboardFeedPath = "/dashboard.atom"
)

type Gitlab struct {
	BaseUrl      string
	ApiPath      string
	RepoFeedPath string
	Token        string
	Client       *http.Client
}

const (
	dateLayout = "2006-01-02T15:04:05-07:00"
)

var (
	skipCertVerify = flag.Bool("gitlab.skip-cert-check", false,
		`If set to true, gitlab client will skip certificate checking for https, possibly exposing your system to MITM attack.`)
)

func NewGitlab(baseUrl, apiPath, token string) *Gitlab {
	config := &tls.Config{InsecureSkipVerify: *skipCertVerify}
	tr := &http.Transport{
		Proxy:           http.ProxyFromEnvironment,
		TLSClientConfig: config,
	}
	client := &http.Client{Transport: tr}

	return &Gitlab{
		BaseUrl: baseUrl,
		ApiPath: apiPath,
		Token:   token,
		Client:  client,
	}
}

func (g *Gitlab) ResourceUrl(url string, params map[string]string) string {

	if params != nil {
		for key, val := range params {
			url = strings.Replace(url, key, val, -1)
		}
	}

	return g.BaseUrl + g.ApiPath + url
}

func (g *Gitlab) ResourceUrlWithQueryValues(url2 string, params map[string]string, vals url.Values) string {
	u, _ := url.Parse(g.ResourceUrl(url2, params))
	q := u.Query()
	for k, vs := range vals {
		for _, v := range vs {
			q.Add(k, v)
		}
	}
	u.RawQuery = q.Encode()
	return u.String()
}

func (g *Gitlab) ResourceUrlWithQuery(url2 string, params, query map[string]string) string {
	u, _ := url.Parse(g.ResourceUrl(url2, params))
	q := u.Query()
	for k, v := range query {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()
	return u.String()
}

type respErr struct {
	status int
	msg string
}

func (e *respErr) Error() string {
	return fmt.Sprintf("Gitlab response error: (%d)%s", e.status, e.msg)
}

func IsNotFoundErr(err error) bool {
	re, ok := err.(*respErr)
	return ok && re.status == http.StatusNotFound
}

func (g *Gitlab) execRequest(method, url string, body []byte) (*http.Response, error) {
	var req *http.Request
	var err error

	if body != nil {
		reader := bytes.NewReader(body)
		req, err = http.NewRequest(method, url, reader)
	} else {
		req, err = http.NewRequest(method, url, nil)
	}

	req.Header.Add("PRIVATE-TOKEN", g.Token)
	if method == "POST" || method == "PUT" {
		req.Header.Add("Content-Type", "application/json")
	}

	if err != nil {
		panic("Error while building gitlab request")
	}

	resp, err := g.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Client.Do error: %q", err)
	}

	if resp.StatusCode >= http.StatusBadRequest {
		defer resp.Body.Close()
		msg, _ := ioutil.ReadAll(resp.Body)
		return nil, &respErr{resp.StatusCode, string(msg)}
	}

	return resp, err
}

func (g *Gitlab) buildAndExecRequest(method, url string, body []byte) ([]byte, error) {
	resp, err := g.execRequest(method, url, body)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("%s", err)
	}

	return contents, err
}

func (g *Gitlab) ResourceUrlRaw(u string, params map[string]string) (string, string) {

	if params != nil {
		for key, val := range params {
			u = strings.Replace(u, key, val, -1)
		}
	}

	path := u
	u = g.BaseUrl + g.ApiPath + path
	p, err := url.Parse(g.BaseUrl)
	if err != nil {
		return u, ""
	}
	opaque := "//" + p.Host + p.Path + g.ApiPath + path

	return u, opaque
}

func (g *Gitlab) buildAndExecRequestRaw(method, url, opaque string, body []byte) ([]byte, error) {

	var req *http.Request
	var err error

	if body != nil {
		reader := bytes.NewReader(body)
		req, err = http.NewRequest(method, url, reader)
	} else {
		req, err = http.NewRequest(method, url, nil)
	}

	req.Header.Add("PRIVATE-TOKEN", g.Token)
	if method == "POST" || method == "PUT" {
		req.Header.Add("Content-Type", "application/json")
	}

	if err != nil {
		panic("Error while building gitlab request")
	}

	if len(opaque) > 0 {
		req.URL.Opaque = opaque
	}

	resp, err := g.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Client.Do error: %q", err)
	}
	defer resp.Body.Close()
	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("%s", err)
	}

	if resp.StatusCode >= 400 {
		return nil, &respErr{resp.StatusCode, string(contents)}
	}

	return contents, err
}
