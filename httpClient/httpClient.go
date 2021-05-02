package httpClient

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"log"
	"net/http"

	"github.com/google/go-querystring/query"
	"github.com/tw7613781/arbitrageBot/util"
)

type client struct {
	c           http.Client
	baseURL     string
	apiKey      string
	apiSecret   string
	queryString string
}

func InitClient(apiKey string, apiSecret string, baseURL string) *client {
	c := client{
		c:         http.Client{},
		apiKey:    apiKey,
		apiSecret: apiSecret,
		baseURL:   baseURL,
	}
	return &c
}

/*
* pair should be trading symbol pair. like "eth-krw"
 */
func (c *client) GetOrderBookBoth(pair string) (OrderResult, error) {
	method := "/public/getorderbook"

	params := &struct {
		Market string `url:"market"`
		Type   string `url:"type"`
	}{
		pair,
		"both",
	}

	resp, err := c.getPublic(method, params)
	if err != nil {
		log.Fatalf("Get balance error: %s", err)
	}

	var output OrderBookBoth
	HttpRespToStruct(resp, &output)
	if !output.Success {
		log.Printf("Error with data message: %s", output.Message)
		return output.Result, errors.New(output.Message)
	}
	return output.Result, nil
}

/*
* pair should be trading symbol pair. like "eth-krw"
* t should be the order book type: "buy", "sell"
 */
func (c *client) GetOrderBookBuyOrSell(pair string, t string) ([]Order, error) {
	method := "/public/getorderbook"

	params := &struct {
		Market string `url:"market"`
		Type   string `url:"type"`
	}{
		pair,
		t,
	}

	resp, err := c.getPublic(method, params)
	if err != nil {
		log.Fatalf("Get balance error: %s", err)
	}

	var output OrderBookBuyOrSell
	HttpRespToStruct(resp, &output)
	if !output.Success {
		log.Printf("Error with data message: %s", output.Message)
		return output.Result, errors.New(output.Message)
	}
	return output.Result, nil
}

func (c *client) GetMarkets() ([]MarketResult, error) {
	method := "/public/getmarkets"

	resp, err := c.getPublic(method, nil)
	if err != nil {
		log.Fatalf("Get balance error: %s", err)
	}

	var output Market
	HttpRespToStruct(resp, &output)
	if !output.Success {
		log.Printf("Error with data message: %s", output.Message)
		return output.Result, errors.New(output.Message)
	}
	return output.Result, nil
}

/*
* pair should be trading symbol pair. like "krw-eth"
 */
func (c *client) GetTicker(pair string) (TickerResult, error) {
	method := "/public/getticker"

	params := &struct {
		Market string `url:"market"`
	}{
		pair,
	}

	resp, err := c.getPublic(method, params)
	if err != nil {
		log.Fatalf("Get balance error: %s", err)
	}

	var output Ticker
	HttpRespToStruct(resp, &output)
	if !output.Success {
		log.Printf("Error with data message: %s", output.Message)
		return output.Result, errors.New(output.Message)
	}
	return output.Result, nil
}

/*
* currency should be cryptocurrency symbol string. like "BTC", "ETH"
 */
func (c *client) GetBalance(currency string) (BalanceResult, error) {
	method := "/account/getbalance"

	params := &struct {
		Apikey   string `url:"apikey"`
		Currency string `url:"currency"`
		Nonce    string `url:"nonce"`
	}{
		c.apiKey,
		currency,
		util.GetTimestampMili(),
	}

	resp, err := c.get(method, params)
	if err != nil {
		log.Fatalf("Get balance error: %s", err)
	}

	var output Balance
	HttpRespToStruct(resp, &output)
	if !output.Success {
		log.Printf("Error with data message: %s", output.Message)
		return output.Result, errors.New(output.Message)
	}
	return output.Result, nil
}

// /*
// * name can be any string,
// * t should be 0 - general wallet, 1 - trade wallet
//  */
// func (c *client) AddWallet(name string, t uint8) (string, error) {
// 	method := "/account/addwallet"

// 	params := &struct {
// 		Apikey string `url:"apikey"`
// 		Name   string `url:"name"`
// 		Nonce  string `url:"nonce"`
// 		Type   uint8  `url:"type"`
// 	}{
// 		c.apiKey,
// 		name,
// 		util.GetTimestampMili(),
// 		t,
// 	}

// 	resp, err := c.get(method, params)
// 	if err != nil {
// 		log.Fatalf("Add wallet error: %s", err)
// 	}

// 	return util.HttpRespToStruct(resp)
// }

/*
* getPublic function queries the endpoints that no need authentication
* method -> "/public/getticker"
* params => {market: 'btc-eth'}
 */
func (c *client) getPublic(method string, params interface{}) (resp *http.Response, err error) {
	if params != nil {
		v, err := query.Values(params)
		if err != nil {
			log.Fatalf("Fail to parse params: %s", err)
		}

		c.queryString = method + "?" + v.Encode()
	} else {
		c.queryString = method
	}

	req, err := http.NewRequest("GET", c.baseURL+c.queryString, nil)
	if err != nil {
		return nil, err
	}

	log.Println(req.URL)
	return c.c.Do(req)
}

/*
* get funcstion queyies the endpoints that need authentication
* method -> "/markert/buylimit"
* params => "{market: 'dash-btc', quantity: '1', rate: '1'}" params except apikey and nonce and the params should strictly listed by alpha orders
 */
func (c *client) get(method string, params interface{}) (*http.Response, error) {
	v, err := query.Values(params)
	if err != nil {
		log.Fatalf("Fail to parse params: %s", err)
	}

	c.queryString = method + "?" + v.Encode()

	req, err := http.NewRequest("GET", c.baseURL+c.queryString, nil)
	if err != nil {
		return nil, err
	}

	return c.do(req)
}

func (c *client) do(req *http.Request) (*http.Response, error) {
	h := hmac.New(sha512.New, []byte(c.apiSecret))
	h.Write([]byte(c.baseURL + c.queryString))
	sha := hex.EncodeToString(h.Sum(nil))

	req.Header.Add("apisign", sha)
	log.Println(req.URL)
	return c.c.Do(req)
}
