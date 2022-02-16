package apxapi

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

func (p *Client) Ping() {
	res, err := p.sendRequest(http.MethodGet, "/fapi/v1/ping", nil, nil)
	if err != nil {
		log.Println("func1 Ping:", err)
	}
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Println("func2 Ping:", err)
	}
	fmt.Println(string(b))
}

type CheckServerTimeResponse struct {
	ServerTime int64 `json:"serverTime"`
}

func (p *Client) CheckServerTime() (result CheckServerTimeResponse, err error) {
	res, err := p.sendRequest(http.MethodGet, "/fapi/v1/time", nil, nil)
	if err != nil {
		log.Println("func1 CheckServerTime:", err)
		return
	}

	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)
	err = json.Unmarshal([]byte(body), &result)
	if err != nil {
		log.Println("func2 CheckServerTime:", err)
		return
	}
	return result, nil
}

type GetExchangeInfoResponse struct {
	Timezone    string `json:"timezone"`
	ServerTime  int64  `json:"serverTime"`
	FuturesType string `json:"futuresType"`
	RateLimits  []struct {
		RateLimitType string `json:"rateLimitType"`
		Interval      string `json:"interval"`
		IntervalNum   int    `json:"intervalNum"`
		Limit         int    `json:"limit"`
	} `json:"rateLimits"`
	ExchangeFilters []interface{} `json:"exchangeFilters"`
	Assets          []struct {
		Asset             string `json:"asset"`
		MarginAvailable   bool   `json:"marginAvailable"`
		AutoAssetExchange string `json:"autoAssetExchange"`
	} `json:"assets"`
	Symbols []struct {
		Symbol                string   `json:"symbol"`
		Pair                  string   `json:"pair"`
		ContractType          string   `json:"contractType"`
		DeliveryDate          int64    `json:"deliveryDate"`
		OnboardDate           int64    `json:"onboardDate"`
		Status                string   `json:"status"`
		MaintMarginPercent    string   `json:"maintMarginPercent"`
		RequiredMarginPercent string   `json:"requiredMarginPercent"`
		BaseAsset             string   `json:"baseAsset"`
		QuoteAsset            string   `json:"quoteAsset"`
		MarginAsset           string   `json:"marginAsset"`
		PricePrecision        int      `json:"pricePrecision"`
		QuantityPrecision     int      `json:"quantityPrecision"`
		BaseAssetPrecision    int      `json:"baseAssetPrecision"`
		QuotePrecision        int      `json:"quotePrecision"`
		UnderlyingType        string   `json:"underlyingType"`
		UnderlyingSubType     []string `json:"underlyingSubType"`
		SettlePlan            int      `json:"settlePlan"`
		TriggerProtect        string   `json:"triggerProtect"`
		LiquidationFee        string   `json:"liquidationFee"`
		MarketTakeBound       string   `json:"marketTakeBound"`
		Filters               []struct {
			MinPrice          string `json:"minPrice,omitempty"`
			MaxPrice          string `json:"maxPrice,omitempty"`
			FilterType        string `json:"filterType"`
			TickSize          string `json:"tickSize,omitempty"`
			StepSize          string `json:"stepSize,omitempty"`
			MaxQty            string `json:"maxQty,omitempty"`
			MinQty            string `json:"minQty,omitempty"`
			Limit             int    `json:"limit,omitempty"`
			Notional          string `json:"notional,omitempty"`
			MultiplierDown    string `json:"multiplierDown,omitempty"`
			MultiplierUp      string `json:"multiplierUp,omitempty"`
			MultiplierDecimal string `json:"multiplierDecimal,omitempty"`
		} `json:"filters"`
		OrderTypes  []string `json:"orderTypes"`
		TimeInForce []string `json:"timeInForce"`
	} `json:"symbols"`
}

func (p *Client) GetExchangeInfo() (result GetExchangeInfoResponse, err error) {
	res, err := p.sendRequest(http.MethodGet, "/fapi/v1/exchangeInfo", nil, nil)
	if err != nil {
		log.Println("func1 GetExchangeInfo:", err)
		return
	}

	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)

	err = json.Unmarshal([]byte(body), &result)
	if err != nil {
		log.Println("func2 GetExchangeInfo:", err)
		return
	}
	return result, nil
}
