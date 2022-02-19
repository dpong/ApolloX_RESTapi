package appolloxapi

import (
	"net/http"
)

type AccountResponse struct {
	FeeTier                     int                  `json:"feeTier"`
	CanTrade                    bool                 `json:"canTrade"`
	CanDeposit                  bool                 `json:"canDeposit"`
	CanWithdraw                 bool                 `json:"canWithdraw"`
	UpdateTime                  int64                `json:"updateTime"`
	TotalInitialMargin          string               `json:"totalInitialMargin"`
	TotalMaintMargin            string               `json:"totalMaintMargin"`
	TotalWalletBalance          string               `json:"totalWalletBalance"`
	TotalUnrealizedProfit       string               `json:"totalUnrealizedProfit"`
	TotalMarginBalance          string               `json:"totalMarginBalance"`
	TotalPositionInitialMargin  string               `json:"totalPositionInitialMargin"`
	TotalOpenOrderInitialMargin string               `json:"totalOpenOrderInitialMargin"`
	TotalCrossWalletBalance     string               `json:"totalCrossWalletBalance"`
	TotalCrossUnPnl             string               `json:"totalCrossUnPnl"`
	AvailableBalance            string               `json:"availableBalance"`
	MaxWithdrawAmount           string               `json:"maxWithdrawAmount"`
	Assets                      []AssetsInAccount    `json:"assets"`
	Positions                   []PositionsInAccount `json:"positions"`
}

type AssetsInAccount struct {
	Asset                  string `json:"asset"`
	WalletBalance          string `json:"walletBalance"`
	UnrealizedProfit       string `json:"unrealizedProfit"`
	MarginBalance          string `json:"marginBalance"`
	MaintMargin            string `json:"maintMargin"`
	InitialMargin          string `json:"initialMargin"`
	PositionInitialMargin  string `json:"positionInitialMargin"`
	OpenOrderInitialMargin string `json:"openOrderInitialMargin"`
	CrossWalletBalance     string `json:"crossWalletBalance"`
	CrossUnPnl             string `json:"crossUnPnl"`
	AvailableBalance       string `json:"availableBalance"`
	MaxWithdrawAmount      string `json:"maxWithdrawAmount"`
}

type PositionsInAccount struct {
	Symbol                 string `json:"symbol"`
	InitialMargin          string `json:"initialMargin"`
	MaintMargin            string `json:"maintMargin"`
	UnrealizedProfit       string `json:"unrealizedProfit"`
	PositionInitialMargin  string `json:"positionInitialMargin"`
	OpenOrderInitialMargin string `json:"openOrderInitialMargin"`
	Leverage               string `json:"leverage"`
	Isolated               bool   `json:"isolated"`
	EntryPrice             string `json:"entryPrice"`
	MaxNotional            string `json:"maxNotional"`
	PositionSide           string `json:"positionSide"`
	PositionAmt            string `json:"positionAmt"`
}

func (b *Client) Account() (*AccountResponse, error) {
	res, err := b.do(http.MethodGet, "fapi/v2/account", nil, true, false)
	if err != nil {
		return nil, err
	}
	resp := &AccountResponse{}
	err = json.Unmarshal(res, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

type BalanceResponse struct {
	AccountAlias       string `json:"accountAlias"`
	Asset              string `json:"asset`
	Balace             string `json:"balance"`
	CrossWalletBalance string `json:"crossWalletBalance"`
	AvailableBalance   string `json:"availableBalance"`
	MaxWithdrawAmount  string `json:"maxWithdrawAmount"`
}

func (b *Client) Balance() ([]*BalanceResponse, error) {
	res, err := b.do(http.MethodGet, "fapi/v2/balance", nil, true, false)
	if err != nil {
		return nil, err
	}
	resp := []*BalanceResponse{}
	err = json.Unmarshal(res, &resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (b *Client) GetIncomeHistory(method, symbol string, limit int, start, end int64) ([]*IncomeResponse, error) {
	opts := IncomeHisOpts{
		Symbol:     symbol,
		Limit:      limit,
		IncomeType: method,
	}
	if start != 0 && end != 0 {
		opts.StartTime = start
		opts.EndTime = end
	}
	if opts.Limit == 0 || opts.Limit > 1000 {
		opts.Limit = 1000
	}
	res, err := b.do(http.MethodGet, "fapi/v1/income", opts, true, false)
	if err != nil {
		return nil, err
	}
	income := []*IncomeResponse{}
	err = json.Unmarshal(res, &income)
	if err != nil {
		return nil, err
	}
	return income, nil
}

type IncomeHisOpts struct {
	Symbol     string `url:"symbol"`
	IncomeType string `url:"incomeType"`
	StartTime  int64  `url:"startTime,omitempty"`
	EndTime    int64  `url:"endTime,omitempty"`
	Limit      int    `url:"limit"`
}

type IncomeResponse struct {
	Symbol     string `json:"symbol"`
	IncomeType string `json:"incomeType"`
	Income     string `json:"income"`
	Asset      string `json:"asset"`
	Info       string `json:"info"`
	Time       int64  `json:"time"`
	TranID     int    `json:"tranId"`
	TradeID    string `json:"tradeId"`
}

type CommissionRateResponse struct {
	Symbol              string `json:"symbol"`
	Makercommissionrate string `json:"makerCommissionRate"`
	Takercommissionrate string `json:"takerCommissionRate"`
}

type onlySymbolOpts struct {
	Symbol string `json:"symbol"`
}

func (b *Client) CommissionRate(symbol string) (*CommissionRateResponse, error) {
	input := onlySymbolOpts{
		Symbol: symbol,
	}
	res, err := b.do(http.MethodGet, "fapi/v1/commissionRate", input, true, false)
	if err != nil {
		return nil, err
	}
	resp := CommissionRateResponse{}
	err = json.Unmarshal(res, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type PositionResponse struct {
	EntryPrice       string `json:"entryPrice"`
	MarginType       string `json:"marginType"`
	IsAutoAddMargin  string `json:"isAutoAddMargin"`
	IsolatedMargin   string `json:"isolatedMargin"`
	Leverage         string `json:"leverage"`
	LiquidationPrice string `json:"liquidationPrice"`
	MarkPrice        string `json:"markPrice"`
	MaxNotionalValue string `json:"maxNotionalValue"`
	PositionAmt      string `json:"positionAmt"`
	Symbol           string `json:"symbol"`
	UnRealizedProfit string `json:"unRealizedProfit"`
	PositionSide     string `json:"positionSide"`
}

func (b *Client) Positions() ([]*PositionResponse, error) {
	res, err := b.do(http.MethodGet, "fapi/v2/positionRisk", nil, true, false)
	if err != nil {
		return nil, err
	}
	resp := []*PositionResponse{}
	err = json.Unmarshal(res, &resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
