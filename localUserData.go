package appolloxapi

import (
	"bytes"
	"context"
	"errors"
	"fmt"
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
	ExecType  string
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

func (u *UserDataBranch) ReadTrade() TradeData {
	data := <-u.trades
	return data
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
				u.maintainUserData(ctx, c, &userData)
				logger.Warningf("Refreshing apx local user data.\n")
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
) {
	innerErr := make(chan error, 1)
	// get the first snapshot to initial data struct
	if err := u.getAccountSnapShot(client); err != nil {
		return
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
			return
		default:
			message := <-(*userData)
			// test
			fmt.Println(message)
			event, ok := message["e"].(string)
			if !ok {
				continue
			}
			switch event {
			case "outboundAccountPosition":
				//u.updateSpotAccountData(&message)
			case "balanceUpdate":
				// next stage, no use for now
			default:
				// pass
			}
		}
	}
}

// func (u *UserDataBranch) updateSpotAccountData(message *map[string]interface{}) {
// 	array, ok := (*message)["B"].([]interface{})
// 	if !ok {
// 		return
// 	}
// 	u.spotAccount.Lock()
// 	defer u.spotAccount.Unlock()
// 	for _, item := range array {
// 		data, ok := item.(map[string]interface{})
// 		if !ok {
// 			continue
// 		}
// 		asset, oka := data["a"].(string)
// 		if !oka {
// 			continue
// 		}
// 		free, okf := data["f"].(string)
// 		if !okf {
// 			continue
// 		}
// 		lock, okl := data["l"].(string)
// 		if !okl {
// 			continue
// 		}
// 		for idx, bal := range u.spotAccount.Data.Balances {
// 			if bal.Asset == asset {
// 				u.Account.Data.Balances[idx].Free = free
// 				u.Account.Data.Balances[idx].Locked = lock
// 				u.Account.LastUpdated = time.Now()
// 				return
// 			}
// 		}
// 	}
// }

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
			// insert to chan
			*mainCh <- res
			if err := w.Conn.SetReadDeadline(time.Now().Add(time.Second * duration)); err != nil {
				innerErr <- errors.New("restart")
				return err
			}
		}
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

func (u *UserDataBranch) insertTrade(input TradeData) {
	if len(u.trades) == cap(u.trades) {
		<-u.trades
	}
	u.trades <- input
}

func (u *UserDataBranch) readerrs() error {
	var buffer bytes.Buffer
	for {
		select {
		case err := <-u.errs:
			buffer.WriteString(err.Error())
			buffer.WriteString(", ")
		default:
			if buffer.Cap() == 0 {
				return nil
			}
			return errors.New(buffer.String())
		}
	}
}
