package quicknode

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
