package apxapi

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type CancelAllOpenOrdersResponse struct {
}

func (p *Client) CancelAllOpenOrders() (result CancelAllOpenOrdersResponse, err error) {
	nonce := time.Now().UnixMilli()
	nonceStr := strconv.FormatInt(nonce, 10)
	fmt.Println(nonceStr)
	params := make(map[string]string)
	params["symbol"] = "BTCUSDT"
	//params["recvWindow"] = "5000"
	params["timestamp"] = nonceStr

	data, err := json.Marshal(params)
	signature := p.Sign(string(data))
	params["signature"] = signature

	u, _ := url.ParseRequestURI(ENDPOINT)
	u.Path = u.Path + "/fapi/v1/batchOrders"
	if params != nil {
		q := u.Query()
		for k, v := range params {
			q.Set(k, v)
		}
		u.RawQuery = q.Encode()
	}
	req, err := http.NewRequest(http.MethodDelete, u.String(), nil)
	fmt.Println(u.String())
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json; charset=UTF-8")
	req.Header.Add("X-MBX-APIKEY", p.apiKey)

	res, err := p.HTTPC.Do(req)
	if err != nil {
		log.Println(err)
	}

	if res.StatusCode != 200 {
		buf := new(bytes.Buffer)
		buf.ReadFrom(res.Body)
		return result, fmt.Errorf("faild to get data. status: %s", res.Status)
	}

	fmt.Println(u.String())

	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)

	fmt.Println(string(body))
	err = json.Unmarshal([]byte(body), &result)
	if err != nil {
		log.Println("func2 CancelAllOpenOrders:", err)
		return
	}
	return result, nil
}

func (p *Client) PlaceOrder() (result CancelAllOpenOrdersResponse, err error) {
	nonce := time.Now().UnixMilli()
	nonceStr := strconv.FormatInt(nonce, 10)
	fmt.Println(nonceStr)
	params := make(map[string]string)
	params["symbol"] = "BTCUSDT"
	params["side"] = "BUY"
	params["type"] = "LIMIT"
	params["timelnForce"] = "GTC"
	params["quantity"] = "1"
	params["price"] = "9000"
	params["recvWindow"] = "5000"
	params["timestamp"] = nonceStr

	data, err := json.Marshal(params)

	signature := p.Sign(string(data))
	params["signature"] = " " + signature

	u, _ := url.ParseRequestURI(ENDPOINT)
	u.Path = u.Path + "/fapi/v1/order"
	if params != nil {
		q := u.Query()
		for k, v := range params {
			q.Set(k, v)
		}
		u.RawQuery = q.Encode()
	}
	req, err := http.NewRequest(http.MethodPost, u.String(), nil)
	fmt.Println(u.String())
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json; charset=UTF-8")
	req.Header.Add("X-MBX-APIKEY", p.apiKey)

	res, err := p.HTTPC.Do(req)
	if err != nil {
		log.Println(err)
	}

	if res.StatusCode != 200 {
		buf := new(bytes.Buffer)
		buf.ReadFrom(res.Body)
		return result, fmt.Errorf("faild to get data. status: %s", res.Status)
	}

	fmt.Println(u.String())

	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)

	fmt.Println(string(body))
	err = json.Unmarshal([]byte(body), &result)
	if err != nil {
		log.Println("func2 CancelAllOpenOrders:", err)
		return
	}
	return result, nil

}
