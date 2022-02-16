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

type GetIncomeHistoryResponse []struct {
	Symbol     string `json:"symbol"`
	IncomeType string `json:"incomeType"`
	Income     string `json:"income"`
	Asset      string `json:"asset"`
	Info       string `json:"info"`
	Time       int64  `json:"time"`
	TranID     string `json:"tranId"`
	TradeID    string `json:"tradeId"`
}

func (p *Client) GetIncomeHistory() (result GetIncomeHistoryResponse, err error) {
	nonce := time.Now().Add(time.Second).Unix()
	nonceStr := strconv.FormatInt(nonce, 10)

	params := make(map[string]string)
	/* params["symbol"] = ""
	params["incomeType"] = ""
	params["startTime"] = ""
	params["endTime"] = ""
	params["limit"] = ""
	params["recvWindow"] = "" */
	params["timestamp"] = nonceStr
	data, err := json.Marshal(params)
	signature := p.Sign(string(data))
	params["signature"] = signature

	u, _ := url.ParseRequestURI(ENDPOINT)
	u.Path = u.Path + "/fapi/v1/income"
	if params != nil {
		q := u.Query()
		for k, v := range params {
			q.Set(k, v)
		}
		u.RawQuery = q.Encode()
	}
	req, err := http.NewRequest(http.MethodGet, u.String(), nil)

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
		return nil, fmt.Errorf("faild to get data. status: %s", res.Status)
	}
	//return res, nil
	fmt.Println(u.String())

	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)

	fmt.Println(string(body))
	err = json.Unmarshal([]byte(body), &result)
	if err != nil {
		log.Println("func2 GetIncomeHistory:", err)
		return
	}
	return result, nil
}
