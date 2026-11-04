package models

type RazorpayWebhook struct {
	Event   string `json:"event"`
	Payload struct {
		Payment struct {
			Entity struct {
				ID       string `json:"id"`
				OrderID  string `json:"order_id"`
				Amount   int64  `json:"amount"`
				Currency string `json:"currency"`
				Status   string `json:"status"`
			} `json:"entity"`
		} `json:"payment"`
	} `json:"payload"`
}
