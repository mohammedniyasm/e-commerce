package request

type AddAddressRequest struct {
	Name         string `json:"name" binding:"required"`
	Phone        string `json:"phone" binding:"required"`
	AddressLine1 string `json:"address_line1" binding:"required"`
	AddressLine2 string `json:"address_line_2"`
	City         string `json:"city" binding:"required"`
	PostalCode   string `json:"postal_code" binding:"required"`
	State        string `json:"state" binding:"required"`
	Country      string `json:"country"`
	IsDefault    bool   `json:"is_default"`
}