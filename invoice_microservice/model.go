package invoice_microservice

type InvoiceState struct {
	ID   string `json:"state_id,omitempty" db:"state_id"`
	Name string `json:"state_name,omitempty" db:"state_name"`
}

type Invoice struct {
	ID                string  `json:"invoice_id,omitempty" db:"invoice_id"`
	Amount            float64 `json:"invoice_amount,omitempty" db:"invoice_amount"`
	State             int     `json:"invoice_state,omitempty" db:"invoice_state"`
	ExpirationDate    string  `json:"invoice_expiration_date,omitempty" db:"invoice_expiration_date"`
	AccountPayerId    string  `json:"invoice_payer_id,omitempty" db:"account_invoice_payer_id"`
	AccountReceiverId string  `json:"invoice_receveiver_id,omitempty" db:"account_invoice_receiver_id"`
}
