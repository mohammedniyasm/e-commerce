package request

type PlaceOrderRequest struct {
	CartItemIDs   []uint `json:"cart_item_ids" binding:"required,min=1"`
	AddressID     uint   `json:"address_id" binding:"required"`
	PaymentMethod string `json:"payment_method" binding:"required"`
	CouponCode    string `json:"coupon_code"`
}

type UpdateOrderStatusRequest struct {
	Status string  `json:"status" binding:"required"`
	Note   *string `json:"note"`
}

type UpdateOrderAddressRequest struct {
	Name         string `json:"name" binding:"required"`
	Phone        string `json:"phone" binding:"required"`
	AddressLine1 string `json:"address_line1" binding:"required"`
	AddressLine2 string `json:"address_line2"`
	City         string `json:"city" binding:"required"`
	State        string `json:"state" binding:"required"`
	Pincode      string `json:"pincode" binding:"required"`
}

type ListOrdersRequest struct {
	Search        string  `form:"search"`
	Limit         int     `form:"limit,default=10"`
	Page          int     `form:"page,default=1"`
	Sort          *string `form:"sort"`
	OrderStatus   *string `form:"order_status"`
	PaymentStatus *string `form:"payment_status"`
}

type ListCustomerOrdersRequest struct {
	Limit int     `form:"limit,default=10"`
	Page  int     `form:"page,default=1"`
	Sort  *string `form:"sort"`
}
type UpdateCODPaymentStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=paid"`
}
