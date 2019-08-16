package coin58

type baseResp struct {
	Code    interface{} `json:"code"`
	Message string      `json:"message"`
}

//账户信息
type accountResp struct {
	baseResp
	Data []accountData `json:"data"`
}

type accountData struct {
	Available    float64 `json:"available"`
	Hold         float64 `json:"hold"`
	CurrencyName string  `json:"currencyName"`
}

//订单
type orderResp struct {
	baseResp
	Data orderData `json:"data"`
}

type ordersResp struct {
	baseResp
	Data []orderData `json:"data"`
}

type orderData struct {
	ID          string `json:"order_id"`
	Amount      string `json:"amount"`
	Price       string `json:"price"`
	Side        string `json:"side	"`
	Type        string `json:"type"`
	Symbol      string `json:"symbol"`
	Status      string `json:"status"`
	BaseFilled  string `json:"base_filled"`
	QuoteFilled string `json:"quote_filled"`
	CreatedTime string `json:"created_time"`
}
