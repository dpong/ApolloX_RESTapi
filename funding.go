package appolloxapi

import "net/http"

func (b *Client) FundingRateHistory(symbol string, limit int, start, end int64) ([]*FundingData, error) {
	opts := FundingRateOpts{
		Symbol: symbol,
		Limit:  limit,
	}
	if start != 0 && end != 0 {
		opts.StartTime = start
		opts.EndTime = end
	}
	if opts.Limit == 0 || opts.Limit > 1000 {
		opts.Limit = 1000
	}
	res, err := b.do(http.MethodGet, "fapi/v1/fundingRate", opts, false, false)
	if err != nil {
		return nil, err
	}
	funding := []*FundingData{}
	err = json.Unmarshal(res, &funding)
	if err != nil {
		return nil, err
	}
	return funding, nil
}

type FundingRateOpts struct {
	Symbol    string `url:"symbol"`
	Limit     int    `url:"limit"`
	StartTime int64  `url:"startTime,omitempty"`
	EndTime   int64  `url:"endTime,omitempty"`
}

type FundingData struct {
	Symbol      string `json:"symbol"`
	FundingRate string `json:"fundingRate"`
	FundingTime int64  `json:"fundingTime"`
}
