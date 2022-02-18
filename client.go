package apxapi

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	jsoniter "github.com/json-iterator/go"
)

/*
The base endpoint is: https://fapi.apollox.finance
All endpoints return either a JSON object or array.
Data is returned in ascending order. Oldest first, newest last.
All time and timestamp related fields are in milliseconds.
All data types adopt definition in JAVA.
*/
var ENDPOINT = "https://fapi.apollox.finance"
var json = jsoniter.ConfigCompatibleWithStandardLibrary

type Client struct {
	privateKey string
	subaccount string
	HTTPC      *http.Client

	apiKey    string
	apiSecret string
}

func New(privateKey, subaccount, apiKey, apiSecret string) *Client {
	hc := &http.Client{
		Timeout: 10 * time.Second,
	}
	return &Client{
		privateKey: privateKey,
		subaccount: subaccount,
		apiKey:     apiKey,
		apiSecret:  apiSecret,
		HTTPC:      hc,
	}
}

func (p *Client) newRequest(method, spath string, body []byte, params *map[string]string) (*http.Request, error) {
	u, _ := url.ParseRequestURI(ENDPOINT)
	u.Path = u.Path + spath
	if params != nil {
		q := u.Query()
		for k, v := range *params {
			q.Set(k, v)
		}
		u.RawQuery = q.Encode()
	}
	req, err := http.NewRequest(method, u.String(), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	p.Headers(req)

	return req, nil
}

func (c *Client) sendRequest(method, spath string, body []byte, params *map[string]string) (*http.Response, error) {
	req, err := c.newRequest(method, spath, body, params)
	if err != nil {
		return nil, err
	}
	res, err := c.HTTPC.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != 200 {
		buf := new(bytes.Buffer)
		buf.ReadFrom(res.Body)
		return nil, fmt.Errorf("faild to get data. status: %s", res.Status)
	}
	return res, nil
}

func decode(res *http.Response, out interface{}) error {
	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)
	err := json.Unmarshal([]byte(body), &out)
	if err == nil {
		return nil
	}
	return err
}

func responseLog(res *http.Response) string {
	b, _ := httputil.DumpResponse(res, true)
	return string(b)
}
func requestLog(req *http.Request) string {
	b, _ := httputil.DumpRequest(req, true)
	return string(b)
}

func (p *Client) Headers(request *http.Request) {
	request.Header.Add("Accept", "application/json")
	request.Header.Add("Content-Type", "application/json; charset=UTF-8")
}

func (p *Client) Sign(data string) (signature string) {
	apiSecret := p.apiSecret
	h := hmac.New(sha256.New, []byte(apiSecret))
	h.Write([]byte(data))
	signature = hex.EncodeToString(h.Sum(nil))
	return
}
