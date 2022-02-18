package appolloxapi

import (
	"net/http"
)

func (b *Client) ChangeInitialLeverage(symbol string, leverage int) (*ChangeLeverageResponse, error) {
	opts := ChnageLeverageOpts{
		Symbol:   symbol,
		Leverage: leverage,
	}
	res, err := b.do(http.MethodPost, "fapi/v1/leverage", opts, true, false)
	if err != nil {
		return nil, err
	}
	resp := &ChangeLeverageResponse{}
	err = json.Unmarshal(res, resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

type ChnageLeverageOpts struct {
	Symbol   string `json:"json"`
	Leverage int    `json:"leverage"`
}

type ChangeLeverageResponse struct {
	Leverage         int    `json:"leverage"`
	MaxNotionalValue string `json:"maxNotionalValue"`
	Symbol           string `json:"symbol"`
}

func (b *Client) NotionalandLeverage() (*[]NotionalandLeverage, error) {
	res, err := b.do(http.MethodGet, "fapi/v1/leverageBracket", nil, true, false)
	if err != nil {
		return nil, err
	}
	resp := []NotionalandLeverage{}
	err = json.Unmarshal(res, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

type NotionalandLeverage struct {
	Symbol   string     `json:"symbol"`
	Brackets []Brackets `json:"brackets"`
}

type Brackets struct {
	Bracket          int     `json:"bracket"`
	InitialLeverage  int     `json:"initialLeverage"`
	NotionalCap      int     `json:"notionalCap"`
	NotionalFloor    int     `json:"notionalFloor"`
	MaintMarginRatio float64 `json:"maintMarginRatio"`
	Cum              float64 `json:"cum"`
}
