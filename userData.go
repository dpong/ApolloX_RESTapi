package appolloxapi

import "net/http"

type PutListenKeyOpts struct {
	ListenKey string `url:"listenKey"`
}

type ListenKeyResponse struct {
	ListenKey string `json:"listenKey"`
}

func (b *Client) GetListenKey() (*ListenKeyResponse, error) {
	res, err := b.do(http.MethodPost, "fapi/v1/listenKey", nil, false, true)
	if err != nil {
		return nil, err
	}
	resp := &ListenKeyResponse{}
	err = json.Unmarshal(res, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (b *Client) PutListenKey(listenKey string) error {
	opts := PutListenKeyOpts{
		ListenKey: listenKey,
	}
	_, err := b.do(http.MethodPut, "fapi/v1/listenKey", opts, false, true)
	if err != nil {
		return err
	}
	return nil
}
