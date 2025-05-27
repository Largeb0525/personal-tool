package database

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
