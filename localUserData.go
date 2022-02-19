package appolloxapi

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"

	"github.com/gorilla/websocket"
)

type UserDataBranch struct {
	account            AccountBranch
	cancel             *context.CancelFunc
	httpUpdateInterval int
	errs               chan error
	trades             chan TradeData
}

type AccountBranch struct {
	sync.RWMutex
	Data *AccountResponse
}

type TradeData struct {
	Symbol    string
	Side      string
	Oid       string
	IsMaker   bool
	Price     decimal.Decimal
	Qty       decimal.Decimal
	Fee       decimal.Decimal
	TimeStamp time.Time
}

func (u *UserDataBranch) Close() {
	(*u.cancel)()
}

// default is 60 sec
func (u *UserDataBranch) SetHttpUpdateInterval(input int) {
	u.httpUpdateInterval = input
}

func (u *UserDataBranch) AccountData() (*AccountResponse, error) {
	u.account.RLock()
	defer u.account.RUnlock()
	return u.account.Data, u.readerrs()
}

func (u *UserDataBranch) ReadTrade() (TradeData, error) {
	if data, ok := <-u.trades; ok {
		return data, nil
	}
	return TradeData{}, errors.New("trade channel already closed.")
}

// default errs cap 5, trades cap 100
func (c *Client) LocalUserData(logger *log.Logger) *UserDataBranch {
	var u UserDataBranch
	ctx, cancel := context.WithCancel(context.Background())
	u.cancel = &cancel
	u.httpUpdateInterval = 60
	u.initialChannels()
	userData := make(chan map[string]interface{}, 100)
	// stream user data
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				res, err := c.GetListenKey() // delete listen key
				if err != nil {
					log.Println("retry listen key for user data stream in 5 sec..")
					time.Sleep(time.Second * 5)
					continue
				}
				if err := c.userData(ctx, res.ListenKey, logger, &userData); err == nil {
					return
				}
				time.Sleep(time.Second)
			}
		}
	}()
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				if err := u.maintainUserData(ctx, c, &userData); err == nil {
					return
				} else {
					logger.Warningf("Refreshing apx local user data with err: %s.\n", err.Error())
				}
			}
		}
	}()
	// wait for connecting
	time.Sleep(time.Second * 5)
	return &u
}

// internal funcs ------------------------------------------------

func (u *UserDataBranch) getAccountSnapShot(client *Client) error {
	u.account.Lock()
	defer u.account.Unlock()
	res, err := client.Account()
	if err != nil {
		return err
	}
	u.account.Data = res
	return nil
}

func (u *UserDataBranch) maintainUserData(
	ctx context.Context,
	client *Client,
	userData *chan map[string]interface{},
) error {
	innerErr := make(chan error, 1)
	// get the first snapshot to initial data struct
	if err := u.getAccountSnapShot(client); err != nil {
		return err
	}

	// update snapshot with steady interval
	go func() {
		snap := time.NewTicker(time.Second * time.Duration(u.httpUpdateInterval))
		defer snap.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-innerErr:
				return
			case <-snap.C:
				if err := u.getAccountSnapShot(client); err != nil {
					u.insertErr(err)
				}
			default:
				time.Sleep(time.Second)
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			close(u.errs)
			close(u.trades)
			return nil
		default:
			message := <-(*userData)
			event, ok := message["e"].(string)
			if !ok {
				continue
			}
			switch event {
			case "ACCOUNT_UPDATE":
				if data, ok := message["a"].(map[string]interface{}); !ok {
					continue
				} else {
					u.handleAccountUpdate(&data)
				}
			case "ORDER_TRADE_UPDATE":
				if data, ok := message["o"].(map[string]interface{}); !ok {
					continue
				} else {
					if event, ok := data["x"].(string); ok {
						switch event {
						case "TRADE":
							u.handleTrade(&data)
						default:
							// order update in the future
						}
					}
				}
			default:
				// pass
			}
		}
	}
}

// default fee asset is USDT
func (u *UserDataBranch) handleTrade(res *map[string]interface{}) {
	data := TradeData{}
	if symbol, ok := (*res)["s"].(string); ok {
		data.Symbol = symbol
	} else {
		return
	}
	if side, ok := (*res)["S"].(string); ok {
		data.Side = strings.ToLower(side)
	} else {
		return
	}
	if qty, ok := (*res)["l"].(string); ok {
		data.Qty, _ = decimal.NewFromString(qty)
	} else {
		return
	}
	if price, ok := (*res)["L"].(string); ok {
		data.Price, _ = decimal.NewFromString(price)
	} else {
		return
	}
	if oid, ok := (*res)["i"].(float64); ok {
		data.Oid = decimal.NewFromFloat(oid).String()
	} else {
		return
	}
	if execType, ok := (*res)["m"].(bool); ok {
		data.IsMaker = execType
	}
	if fee, ok := (*res)["n"].(string); ok {
		data.Fee, _ = decimal.NewFromString(fee)
	}
	if st, ok := (*res)["T"].(float64); ok {
		stamp := FormatingTimeStamp(st)
		data.TimeStamp = stamp
	}
	u.insertTrade(&data)
}

func (u *UserDataBranch) handleAccountUpdate(res *map[string]interface{}) {
	if balances, ok := (*res)["B"].([]interface{}); ok {
		for _, item := range balances {
			data := item.(map[string]interface{})
			var asset, walletBalance, crossWalletBalance string
			if a, ok := data["a"].(string); ok {
				asset = a
			}
			if wb, ok := data["wb"].(string); ok {
				walletBalance = wb
			}
			if cw, ok := data["cw"].(string); ok {
				crossWalletBalance = cw
			}
			u.updateBalanceData(asset, walletBalance, crossWalletBalance)
		}
	}
	if positions, ok := (*res)["P"].([]interface{}); ok {
		for _, item := range positions {
			data := item.(map[string]interface{})
			var symbol, amount, entryPrice, unPnl, marginType, positionSide string
			if s, ok := data["s"].(string); ok {
				symbol = s
			}
			if pa, ok := data["pa"].(string); ok {
				amount = pa
			}
			if ep, ok := data["ep"].(string); ok {
				entryPrice = ep
			}
			if up, ok := data["up"].(string); ok {
				unPnl = up
			}
			if mt, ok := data["mt"].(string); ok {
				marginType = mt
			}
			if ps, ok := data["ps"].(string); ok {
				positionSide = ps
			}
			u.updatePositionData(symbol, amount, entryPrice, unPnl, marginType, positionSide)
		}
	}

}

func (u *UserDataBranch) updateBalanceData(asset, walletBalance, crossWalletBalance string) {
	u.account.Lock()
	defer u.account.Unlock()
	for idx, item := range u.account.Data.Assets {
		if item.Asset == asset {
			u.account.Data.Assets[idx].WalletBalance = walletBalance
			u.account.Data.Assets[idx].CrossWalletBalance = crossWalletBalance
			break
		}
	}
}

func (u *UserDataBranch) updatePositionData(symbol, amount, entryPrice, unPnl, marginType, positionSide string) {
	u.account.Lock()
	defer u.account.Unlock()
	for idx, item := range u.account.Data.Positions {
		if item.Symbol == symbol {
			u.account.Data.Positions[idx].PositionAmt = amount
			u.account.Data.Positions[idx].EntryPrice = entryPrice
			u.account.Data.Positions[idx].UnrealizedProfit = unPnl
			if marginType == "isolated" {
				if !u.account.Data.Positions[idx].Isolated {
					u.account.Data.Positions[idx].Isolated = true
				}
			} else {
				if u.account.Data.Positions[idx].Isolated {
					u.account.Data.Positions[idx].Isolated = false
				}
			}
			u.account.Data.Positions[idx].PositionSide = positionSide
			break
		}
	}
}

func (c *Client) userData(ctx context.Context, listenKey string, logger *log.Logger, mainCh *chan map[string]interface{}) error {
	var w wS
	var duration time.Duration = 1810
	w.Logger = logger
	w.OnErr = false
	var buffer bytes.Buffer
	innerErr := make(chan error, 1)
	buffer.WriteString("wss://fstream.apollox.finance/ws/")
	buffer.WriteString(listenKey)
	url := buffer.String()
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return err
	}
	log.Println("Connected:", url)
	w.Conn = conn
	defer conn.Close()
	if err := w.Conn.SetReadDeadline(time.Now().Add(time.Second * duration)); err != nil {
		return err
	}
	w.Conn.SetPingHandler(nil)
	go func() {
		putKey := time.NewTicker(time.Minute * 30)
		defer putKey.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-innerErr:
				return
			case <-putKey.C:
				if err := c.PutListenKey(listenKey); err != nil {
					// time out in 1 sec
					w.Conn.SetReadDeadline(time.Now().Add(time.Second))
				}
				w.Conn.SetReadDeadline(time.Now().Add(time.Second * duration))
			default:
				time.Sleep(time.Second)
			}
		}
	}()
	for {
		select {
		case <-ctx.Done():
			w.OutApxErr()
			message := "Apx User Data closed..."
			log.Println(message)
			return errors.New(message)
		default:
			if w.Conn == nil {
				w.OutApxErr()
				message := "Apx User Data reconnect..."
				log.Println(message)
				innerErr <- errors.New("restart")
				return errors.New(message)
			}
			_, buf, err := w.Conn.ReadMessage()
			if err != nil {
				w.OutApxErr()
				message := "Apx User Data reconnect..."
				log.Println(message)
				innerErr <- errors.New("restart")
				return errors.New(message)
			}
			res, err1 := DecodingMap(buf, logger)
			if err1 != nil {
				w.OutApxErr()
				message := "Apx User Data reconnect..."
				log.Println(message, err1)
				innerErr <- errors.New("restart")
				return err1
			}
			// check event time first
			handleUserData(&res, mainCh)
			if err := w.Conn.SetReadDeadline(time.Now().Add(time.Second * duration)); err != nil {
				innerErr <- errors.New("restart")
				return err
			}
		}
	}
}

func handleUserData(res *map[string]interface{}, mainCh *chan map[string]interface{}) {
	if eventTimeUnix, ok := (*res)["E"].(float64); ok {
		eventTime := FormatingTimeStamp(eventTimeUnix)
		if time.Now().After(eventTime.Add(time.Minute * 60)) {
			return
		}
		// insert to chan
		*mainCh <- *res
	}
}

func (u *UserDataBranch) initialChannels() {
	// 5 err is allowed
	u.errs = make(chan error, 5)
	u.trades = make(chan TradeData, 100)
}

func (u *UserDataBranch) insertErr(input error) {
	if len(u.errs) == cap(u.errs) {
		<-u.errs
	}
	u.errs <- input
}

func (u *UserDataBranch) insertTrade(input *TradeData) {
	if len(u.trades) == cap(u.trades) {
		<-u.trades
	}
	u.trades <- *input
}

func (u *UserDataBranch) readerrs() error {
	var buffer bytes.Buffer
	for {
		select {
		case err, ok := <-u.errs:
			if ok {
				buffer.WriteString(err.Error())
				buffer.WriteString(", ")
			} else {
				buffer.WriteString("errs chan already closed, ")
			}
		default:
			if buffer.Cap() == 0 {
				return nil
			}
			return errors.New(buffer.String())
		}
	}
}
