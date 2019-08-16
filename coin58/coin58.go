package coin58

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"time"

	. "github.com/nntaoli-project/GoEx"
)

type Coin58 struct {
	httpClient *http.Client
	baseUrl    string
	accessKey  string
	secretKey  string
}

func NewCoin58(client *http.Client, apikey, secretkey string) *Coin58 {
	coin58 := new(Coin58)
	coin58.baseUrl = "https://openapi.58ex.com/v1/"
	coin58.httpClient = client
	coin58.accessKey = apikey
	coin58.secretKey = secretkey
	return coin58
}

func New(client *http.Client, api_key, secret_key string) *Coin58 {
	return &Coin58{accessKey: api_key, secretKey: secret_key, httpClient: client}
}

func (coin58 *Coin58) parseDepthData(tick map[string]interface{}) *Depth {
	bids, _ := tick["bids"].([]interface{})
	asks, _ := tick["asks"].([]interface{})

	depth := new(Depth)
	for _, r := range asks {
		var dr DepthRecord
		rr := r.([]interface{})
		dr.Price = ToFloat64(rr[0])
		dr.Amount = ToFloat64(rr[1])
		depth.AskList = append(depth.AskList, dr)
	}

	for _, r := range bids {
		var dr DepthRecord
		rr := r.([]interface{})
		dr.Price = ToFloat64(rr[0])
		dr.Amount = ToFloat64(rr[1])
		depth.BidList = append(depth.BidList, dr)
	}

	sort.Sort(sort.Reverse(depth.AskList))

	return depth
}

func (coin58 *Coin58) parseOrder(ordmap orderData) Order {
	ord := Order{
		OrderID:    ToInt(ordmap.ID),
		OrderID2:   fmt.Sprint(ToInt(ordmap.ID)),
		Amount:     ToFloat64(ordmap.Amount) / ToFloat64(ordmap.Price),
		Price:      ToFloat64(ordmap.Price),
		DealAmount: ToFloat64(ordmap.BaseFilled),
		Fee:        0,
		OrderTime:  ToInt(ordmap.CreatedTime),
	}

	switch ordmap.Status {
	case "Received", "Active":
		ord.Status = ORDER_UNFINISH
	case "Finished":
		ord.Status = ORDER_FINISH
	case "Cancelling", "Cancelled":
		ord.Status = ORDER_CANCEL
	default:
		ord.Status = ORDER_UNFINISH
	}

	if ord.DealAmount > 0.0 {
		ord.AvgPrice = ToFloat64(ordmap.Price)
	}

	typeS := fmt.Sprintf("%s-%s", ordmap.Side, ordmap.Type)
	switch typeS {
	case "buy-limit":
		ord.Side = BUY
	case "buy-market":
		ord.Side = BUY_MARKET
	case "sell-limit":
		ord.Side = SELL
	case "sell-market":
		ord.Side = SELL_MARKET
	}
	return ord
}

func (coin58 *Coin58) toJson(params url.Values) string {
	parammap := make(map[string]string)
	for k, v := range params {
		parammap[k] = v[0]
	}
	jsonData, _ := json.Marshal(parammap)
	return string(jsonData)
}

func (coin58 *Coin58) signature(postForm *url.Values) map[string]string {
	data := url.Values{}
	data.Set("AccessKeyId", coin58.accessKey)
	data.Set("SignatureMethod", "HmacSHA256")
	data.Set("SignatureVersion", "2")
	data.Set("Timestamp", strconv.Itoa(int(time.Now().UnixNano()/1e6)))
	payload := fmt.Sprintf("%s&%s", data.Encode(), postForm.Encode())
	sign, _ := GetParamHmacSHA256Sign(coin58.secretKey, payload)
	data.Set("Signature", sign)

	return map[string]string{
		"X-58COIN-APIKEY": coin58.accessKey,
		"Timestamp":       data.Get("Timestamp"),
		"Signature":       data.Get("Signature"),
		"Content-Type":    "application/json",
		"Accept-Language": "zh-cn",
	}
}

func (coin58 *Coin58) placeOrder(amount, price string, pair CurrencyPair, orderType, side string) (string, error) {
	path := "spot/my/order/place"

	params := url.Values{}

	params.Set("side", side)
	params.Set("symbol", pair.AdaptUsdToUsdt().ToLower().ToSymbol("_"))
	params.Set("type", orderType)

	switch orderType {
	case "limit":
		params.Set("price", price)
	}
	params.Set("amount", amount)

	headers := coin58.signature(&params)

	resp, err := HttpPostForm2(coin58.httpClient, coin58.baseUrl+path, params, headers)
	if err != nil {
		return "", err
	}

	log.Println("resp:", string(resp))
	respmap := orderResp{}
	err = json.Unmarshal(resp, &respmap)
	if err != nil {
		return "", err
	}

	_, ok := respmap.Code.(string)

	if ok {
		return "", errors.New(respmap.Message)
	}

	return respmap.Data.ID, nil

}

func (coin58 *Coin58) LimitBuy(amount, price string, currency CurrencyPair) (*Order, error) {
	orderId, err := coin58.placeOrder(amount, price, currency, "limit", "buy")
	if err != nil {
		return nil, err
	}
	return &Order{
		Currency: currency,
		OrderID:  ToInt(orderId),
		OrderID2: orderId,
		Amount:   ToFloat64(amount),
		Price:    ToFloat64(price),
		Side:     BUY}, nil
}

func (coin58 *Coin58) LimitSell(amount, price string, currency CurrencyPair) (*Order, error) {
	orderId, err := coin58.placeOrder(amount, price, currency, "limit", "sell")
	if err != nil {
		return nil, err
	}
	return &Order{
		Currency: currency,
		OrderID:  ToInt(orderId),
		OrderID2: orderId,
		Amount:   ToFloat64(amount),
		Price:    ToFloat64(price),
		Side:     BUY}, nil
}

func (coin58 *Coin58) MarketBuy(amount, price string, currency CurrencyPair) (*Order, error) {
	orderId, err := coin58.placeOrder(amount, price, currency, "market", "buy")
	if err != nil {
		return nil, err
	}
	return &Order{
		Currency: currency,
		OrderID:  ToInt(orderId),
		OrderID2: orderId,
		Amount:   ToFloat64(amount),
		Price:    ToFloat64(price),
		Side:     BUY}, nil
}

func (coin58 *Coin58) MarketSell(amount, price string, currency CurrencyPair) (*Order, error) {
	orderId, err := coin58.placeOrder(amount, price, currency, "market", "sell")
	if err != nil {
		return nil, err
	}
	return &Order{
		Currency: currency,
		OrderID:  ToInt(orderId),
		OrderID2: orderId,
		Amount:   ToFloat64(amount),
		Price:    ToFloat64(price),
		Side:     BUY}, nil
}

func (coin58 *Coin58) GetAccount() (*Account, error) {
	path := "account/assets/sites"
	params := url.Values{}
	params.Set("siteId", "1")

	headers := coin58.signature(&params)

	resp, err := HttpGet2(coin58.httpClient, coin58.baseUrl+path+"?"+params.Encode(), headers)
	if err != nil {
		return nil, err
	}

	respJson, _ := json.Marshal(resp)
	respmap := accountResp{}

	err = json.Unmarshal(respJson, &respmap)
	if err != nil {
		log.Println("GetAccount.Unmarshal error:", err)
		return nil, err
	}

	_, ok := respmap.Code.(string)

	if ok {
		return nil, errors.New(respmap.Message)
	}

	acc := new(Account)
	acc.SubAccounts = make(map[Currency]SubAccount, 6)
	acc.Exchange = coin58.GetExchangeName()

	subAccMap := make(map[Currency]*SubAccount)

	for _, item := range respmap.Data {
		currencySymbol := item.CurrencyName
		currency := NewCurrency(currencySymbol, "")
		if subAccMap[currency] == nil {
			subAccMap[currency] = new(SubAccount)
		}
		subAccMap[currency].Currency = currency
		subAccMap[currency].Amount = item.Available
		subAccMap[currency].ForzenAmount = item.Hold
	}

	for k, v := range subAccMap {
		acc.SubAccounts[k] = *v
	}

	return acc, nil
}

func (coin58 *Coin58) CancelOrder(orderId string, currency CurrencyPair) (bool, error) {
	path := "spot/my/order/cancel"

	params := url.Values{}

	params.Set("order_id", orderId)
	params.Set("symbol", currency.AdaptUsdToUsdt().ToLower().ToSymbol("_"))

	headers := coin58.signature(&params)

	resp, err := HttpPostForm2(coin58.httpClient, coin58.baseUrl+path, params, headers)
	if err != nil {
		return false, err
	}

	// log.Println(string(resp))

	respmap := orderResp{}
	err = json.Unmarshal(resp, &respmap)
	if err != nil {
		return false, err
	}

	_, ok := respmap.Code.(string)

	if ok {
		return false, errors.New(respmap.Message)
	}

	return true, nil
}

func (coin58 *Coin58) GetOneOrder(orderId string, currency CurrencyPair) (*Order, error) {
	path := "spot/my/order"

	params := url.Values{}

	params.Set("order_id", orderId)
	params.Set("symbol", currency.AdaptUsdToUsdt().ToLower().ToSymbol("_"))

	headers := coin58.signature(&params)

	resp, err := HttpGet2(coin58.httpClient, coin58.baseUrl+path+"?"+params.Encode(), headers)
	if err != nil {
		return &Order{}, err
	}

	respJson, _ := json.Marshal(resp)
	respmap := orderResp{}
	err = json.Unmarshal(respJson, &respmap)
	if err != nil {
		return &Order{}, err
	}

	_, ok := respmap.Code.(string)

	if ok {
		return &Order{}, errors.New(respmap.Message)
	}

	data := respmap.Data

	order := coin58.parseOrder(data)
	order.Currency = currency
	return &order, nil

}

func (coin58 *Coin58) GetUnfinishOrders(currency CurrencyPair) ([]Order, error) {
	path := "spot/my/orders"

	params := url.Values{}

	params.Set("symbol", currency.AdaptUsdToUsdt().ToLower().ToSymbol("_"))

	headers := coin58.signature(&params)

	resp, err := HttpGet2(coin58.httpClient, coin58.baseUrl+path+"?"+params.Encode(), headers)
	if err != nil {
		return []Order{}, err
	}

	respJson, _ := json.Marshal(resp)

	respmap := ordersResp{}
	err = json.Unmarshal(respJson, &respmap)
	if err != nil {
		return []Order{}, err
	}

	_, ok := respmap.Code.(string)

	if ok {
		return []Order{}, errors.New(respmap.Message)
	}

	datas := respmap.Data

	orders := []Order{}
	for _, data := range datas {
		order := coin58.parseOrder(data)
		order.Currency = currency
		orders = append(orders, order)
	}

	return orders, nil
}

func (coin58 *Coin58) GetOrderHistorys(currency CurrencyPair, currentPage, pageSize int) ([]Order, error) {
	panic("unimplements")
}

func (coin58 *Coin58) GetDepth(size int, currency CurrencyPair) (*Depth, error) {
	path := "spot/order_book"

	params := url.Values{}
	params.Set("symbol", currency.AdaptUsdToUsdt().ToLower().ToSymbol("_"))
	params.Set("limit", strconv.Itoa(size))

	respmap, err := HttpGet(coin58.httpClient, coin58.baseUrl+path+"?"+params.Encode())
	if err != nil {
		return nil, err
	}
	// log.Println(respmap)

	if 0 != respmap["code"].(float64) {
		return nil, errors.New(respmap["message"].(string))
	}

	tick, _ := respmap["data"].(map[string]interface{})

	dep := coin58.parseDepthData(tick)
	dep.Pair = currency

	return dep, nil
}

func (coin58 *Coin58) GetKlineRecords(currency CurrencyPair, period, size, since int) ([]Kline, error) {
	panic("unimplements")
}

func (coin58 *Coin58) GetTrades(currencyPair CurrencyPair, since int64) ([]Trade, error) {
	panic("unimplements")
}

func (coin58 *Coin58) GetTicker(currency CurrencyPair) (*Ticker, error) {
	panic("unimplements")
}

func (coin58 *Coin58) GetExchangeName() string {
	return "58ex.com"
}
