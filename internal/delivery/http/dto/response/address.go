package response

import "time"

type AddressResponse struct {
	ID           uint       `json:"id"`
	Name         string     `json:"name"`
	Phone        string     `json:"phone"`
	AddressLine1 string     `json:"address_line1"`
	AddressLine2 string     `json:"address_line_2"`
	City         string     `json:"city"`
	PostalCode   string     `json:"postal_code"`
	State        string     `json:"state"`
	Country      string     `json:"country"`
	IsDefault    bool       `json:"is_default"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    *time.Time `json:"updated_at"`
}