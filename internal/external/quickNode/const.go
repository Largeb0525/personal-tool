package quickNode

const (
	quickAlertsURL        = "https://api.quicknode.com/quickalerts/rest/v1/notifications/"
	freezeBalanceURL      = "https://methodical-smart-yard.tron-mainnet.quiknode.pro/%s/wallet/freezebalancev2"
	unfreezeBalanceURL    = "https://methodical-smart-yard.tron-mainnet.quiknode.pro/%s/wallet/unfreezebalancev2"
	delegateResourceURL   = "https://methodical-smart-yard.tron-mainnet.quiknode.pro/%s/wallet/delegateresource"
	undelegateResourceURL = "https://methodical-smart-yard.tron-mainnet.quiknode.pro/%s/wallet/undelegateresource"
	broadcastURL          = "https://methodical-smart-yard.tron-mainnet.quiknode.pro/%s/wallet/broadcasttransaction"
	smartContractURL      = "https://methodical-smart-yard.tron-mainnet.quiknode.pro/%s/wallet/triggerconstantcontract"
)

var (
	ApiKey       = ""
	QuickAlertID = ""
	AppID        = ""
)
