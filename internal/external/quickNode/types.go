package quickNode

type QuickAlert struct {
	ID           string        `json:"id"`
	CreatedAt    string        `json:"created_at"`
	UpdatedAt    string        `json:"updated_at"`
	Name         string        `json:"name"`
	Expression   string        `json:"expression"`
	Network      string        `json:"network"`
	Destinations []Destination `json:"destinations"`
	Enabled      bool          `json:"enabled"`
}

type Destination struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	To          string `json:"to"`
	WebhookType string `json:"webhook_type"`
	Service     string `json:"service"`
	PayloadType int    `json:"payload_type"`
}

type PatchQuickAlertRequest struct {
	Name           string   `json:"name,omitempty"`
	Expression     string   `json:"expression"`
	DestinationIDs []string `json:"destinationIds,omitempty"`
}

type FreezeRequest struct {
	OwnerAddress  string `json:"owner_address,required"`
	Resource      string `json:"resource,required"`
	FrozenBalance int64  `json:"frozen_balance,required"`
	Visible       bool   `json:"visible"`
}

type UnfreezeRequest struct {
	OwnerAddress    string `json:"owner_address,required"`
	Resource        string `json:"resource,required"`
	UnfreezeBalance int64  `json:"unfreeze_balance,required"`
	Visible         bool   `json:"visible"`
}

type Transaction struct {
	Visible    bool                   `json:"visible"`
	TxID       string                 `json:"txid"`
	RawData    map[string]interface{} `json:"raw_data"`
	RawDataHex string                 `json:"raw_data_hex"`
}

type BroadcastRequest struct {
	TxID       string                 `json:"txid"`
	RawData    map[string]interface{} `json:"raw_data"`
	RawDataHex string                 `json:"raw_data_hex"`
	Signature  string                 `json:"signature,required"`
	Visible    bool                   `json:"visible"`
}

type BroadcastResponse struct {
	Txid    string `json:"txid"`
	Code    string `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

type DelegateResourceRequest struct {
	OwnerAddress    string `json:"owner_address"`
	ReceiverAddress string `json:"receiver_address"`
	Balance         int64  `json:"balance"`
	Resource        string `json:"resource"`
	Lock            bool   `json:"lock"`
	Visible         bool   `json:"visible,omitempty"`
}

type UndelegateResourceRequest struct {
	OwnerAddress    string `json:"owner_address"`
	ReceiverAddress string `json:"receiver_address"`
	Balance         int64  `json:"balance"`
	Resource        string `json:"resource"`
	Visible         bool   `json:"visible,omitempty"`
}

type TriggerSmartContractRequest struct {
	OwnerAddress     string `json:"owner_address"`     // 查詢者地址（Base58）
	ContractAddress  string `json:"contract_address"`  // TRC20 合約地址（如 USDT）
	FunctionSelector string `json:"function_selector"` // e.g. "balanceOf(address)"
	Parameter        string `json:"parameter"`         // ABI 編碼參數（64 hex 字元）
	Visible          bool   `json:"visible,omitempty"` // Base58 格式時設為 true
}

type TriggerSmartContractResponse struct {
	ConstantResult []string `json:"constant_result"` // 呼叫 view 方法的 return 值（hex 字串）
	Result         struct {
		Result bool `json:"result"`
	} `json:"result"`
}
