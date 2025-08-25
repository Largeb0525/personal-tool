package database

import "time"

type Chat struct {
	ID    int64
	Title string
}

type EventHistory struct {
	TransactionHash  string
	USDT             string
	FromAddress      string
	ToAddress        string
	Platform         string
	WalletUSDT       string
	OrderID          string
	AskEnergySuccess bool
}

type DelegateRecord struct {
	ID              int
	ReceiverAddress string
	TxID            string
	Undelegated     bool
	CreatedAt       time.Time
}

type PendingOrder struct {
	MerchantOrderID    string
	CustomerUsername   string
	AdvertiserUsername string
	OrderStatus        string
	DisplayFiatAmount  float64
	Retries            int
	OriginalChatID     int64
	ReplyToMessageID   int64
	CreatedAt          time.Time
	UpdatedAt          time.Time
}
