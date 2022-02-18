package appolloxapi

import (
	"net/http"
)

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

func (b *Client) GetExchangeInfo() (*GetExchangeInfoResponse, error) {
	res, err := b.do(http.MethodGet, "fapi/v1/exchangeInfo", nil, false, false)
	if err != nil {
		return nil, err
	}
	exchange := &GetExchangeInfoResponse{}
	err = json.Unmarshal(res, &exchange)
	if err != nil {
		return nil, err
	}
	return exchange, nil
}
