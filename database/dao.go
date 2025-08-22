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
