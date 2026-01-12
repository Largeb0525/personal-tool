package andy

type RequestPayload struct {
	Amount   string `json:"amount" binding:"required"`
	Platform string `json:"platform" binding:"required"`
	Method   string `json:"method" binding:"required"`
	Name     string `json:"name" binding:"required"`
}

type Payload struct {
	Merchant_id       string `json:"merchant_id"`
	Merchant_order_id string `json:"merchant_order_id"`
	Amount            string `json:"amount"`
	Notify_url        string `json:"notify_url"`
	Payer             string `json:"payer"`
	Payment_method    string `json:"payment_method"`
	Apply_timestamp   int64  `json:"apply_timestamp"`
	Md5_sign          string `json:"md5_sign"`
}

type Log struct {
	Address          string   `json:"address"`
	BlockHash        string   `json:"blockHash"`
	BlockNumber      string   `json:"blockNumber"`
	Data             string   `json:"data"`
	LogIndex         string   `json:"logIndex"`
	Removed          bool     `json:"removed"`
	Topics           []string `json:"topics"`
	TransactionHash  string   `json:"transactionHash"`
	TransactionIndex string   `json:"transactionIndex"`
}

type TransactionReceipt struct {
	BlockHash         string `json:"blockHash"`
	BlockNumber       string `json:"blockNumber"`
	ContractAddress   string `json:"contractAddress"`
	CumulativeGasUsed string `json:"cumulativeGasUsed"`
	EffectiveGasPrice string `json:"effectiveGasPrice"`
	From              string `json:"from"`
	GasUsed           string `json:"gasUsed"`
	Logs              []Log  `json:"logs"`
	LogsBloom         string `json:"logsBloom"`
	Status            string `json:"status"`
	To                string `json:"to"`
	TransactionHash   string `json:"transactionHash"`
	TransactionIndex  string `json:"transactionIndex"`
	Type              string `json:"type"`
}

type MatchingReceipts struct {
	MatchingReceipts []TransactionReceipt `json:"matchingReceipts"`
}

type ParsedTransaction struct {
	TransactionHash string `json:"transaction_hash"`
	URL             string `json:"url"`
	USDT            string `json:"usdt"`
	FromAddress     string `json:"from_address"`
	ToAddress       string `json:"to_address"`
}

type TokenInfo struct {
	TokenPriceInUSD         string      `json:"token_price_in_usd"`
	FrozenTokenValueInUSD   string      `json:"frozen_token_value_in_usd,omitempty"`
	Level                   interface{} `json:"level"` // 有的時候是 string，有的時候是 int，所以用 interface{}
	Frozen                  interface{} `json:"frozen,omitempty"`
	TokenValue              string      `json:"token_value"`
	TokenType               int         `json:"token_type"`
	TokenPrice              string      `json:"token_price"`
	TokenDecimal            int         `json:"token_decimal"`
	TokenValueInUSD         string      `json:"token_value_in_usd"`
	FrozenV2                interface{} `json:"frozenV2,omitempty"`
	TokenID                 string      `json:"token_id"`
	TokenAbbr               string      `json:"token_abbr"`
	Balance                 string      `json:"balance"`
	FrozenV2TokenValueInUSD string      `json:"frozenV2_token_value_in_usd,omitempty"`
	TokenName               string      `json:"token_name"`
	PairID                  int         `json:"pair_id,omitempty"`
	VIP                     bool        `json:"vip"`
	TokenURL                string      `json:"token_url"`
	TransferCount           int64       `json:"transferCount,omitempty"`
	NrOfTokenHolders        int64       `json:"nrOfTokenHolders,omitempty"`
}

type TronWalletFullResponse struct {
	Data  []TokenInfo `json:"data"`
	Count int         `json:"count"`
}

type AskEnergyData struct {
	Id       int    `json:"Id"`
	OptionId string `json:"optionId"`
	Address  string `json:"address"`
}

type AskEnergyResponse struct {
	Code int               `json:"code"`
	Msg  string            `json:"msg"`
	Data AskEnergyData     `json:"data"`
}

type thresholdRequest struct {
	BThreshold float64 `json:"bThreshold"`
	IThreshold float64 `json:"iThreshold"`
}

type IndiaOrderResponse struct {
	Count    int         `json:"count"`
	Next     *string     `json:"next"`
	Previous *string     `json:"previous"`
	Results  []IndiaOrder `json:"results"`
}

type IndiaOrder struct {
	ID                     string  `json:"id"`
	DisplayCryptoAmount    float64 `json:"displayCryptoAmount"`
	DisplayFiatAmount      float64 `json:"displayFiatAmount"`
	AdvertiserExchangeRate float64 `json:"advertiserExchangeRate"`
	CustomerExchangeRate   float64 `json:"customerExchangeRate"`
	CustomerHandlingFee    float64 `json:"customerHandlingFee"`
	CustomerUsername       string  `json:"customerUsername"`
	AdvertiserUsername     string  `json:"advertiserUsername"`
	MarketName             string  `json:"marketName"`
	UpiAccountInformation  *UpiAccountInformation `json:"upiAccountInformation"`
	ImpsAccountInformation *interface{} `json:"impsAccountInformation"` // Use interface{} for null or unknown structure
	AlipayAccountInformation *interface{} `json:"alipayAccountInformation"`
	ECnyAccountInformation   *interface{} `json:"eCnyAccountInformation"`
	MerchantOrderId        string  `json:"merchantOrderId"`
	PayerName              string  `json:"payerName"`
	UtrID                  *string `json:"utrId"`
	TransactionReference   string  `json:"transactionReference"`
	TransactionType        *string `json:"transactionType"`
	OrderDatetime          string  `json:"orderDatetime"`
	PaymentStatus          *string `json:"paymentStatus"`
	OrderStatus            string  `json:"orderStatus"`
	OrderStatusRecords     []OrderStatusRecord `json:"orderStatusRecords"`
	OrderRiskControlStatus string  `json:"orderRiskControlStatus"`
	Source                 string  `json:"source"`
	CreatedAt              string  `json:"createdAt"`
	UpdatedAt              string  `json:"updatedAt"`
	CompletedAt            *string `json:"completedAt"`
	CanceledAt             *string `json:"canceledAt"`
}

type UpiAccountInformation struct {
	UpiID        string `json:"upiId"`
	Name         string `json:"name"`
	UploadFileUrl string `json:"uploadFileUrl"`
	AccountUrl   string `json:"accountUrl"`
}

type OrderStatusRecord struct {
	Status    string `json:"status"`
	Operator  string `json:"operator"`
	CreatedAt string `json:"createdAt"`
}
