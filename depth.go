package appolloxapi

import (
	"net/http"
)

type DepthOpts struct {
	Symbol string `url:"symbol"`
	Limit  int    `url:"limit"`
}

func (b *Client) Depth(symbol string, limit int) (*Depth, error) {
	opts := DepthOpts{
		Symbol: symbol,
		Limit:  limit,
	}
	if opts.Limit == 0 || opts.Limit > 1000 {
		opts.Limit = 100
	}
	res, err := b.do(http.MethodGet, "fapi/v1/depth", opts, false, false)
	if err != nil {
		return nil, err
	}
	depth := Depth{}
	err = json.Unmarshal(res, &depth)
	if err != nil {
		return nil, err
	}
	return &depth, nil
}

type Depth struct {
	LastUpdateID    int        `json:"lastUpdateId"`
	Bids            [][]string `json:"bids"`
	Asks            [][]string `json:"asks"`
	MessageOutTime  int64      `json:"E"`
	TransactionTime int64      `json:"T"`
}
