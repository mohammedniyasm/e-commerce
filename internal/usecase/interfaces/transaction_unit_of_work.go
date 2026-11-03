package interfaces

type TransactionRepositories struct {
	User         UserRepository
	Notification NotificationRepository

	Order              OrderRepository
	OrderItem          OrderItemRepository
	OrderAddress       OrderAddressRepository
	OrderStatusHistory OrderStatusHistoryRepository
	Shipment           ShipmentRepository
	ReturnRequest      ReturnRequestRepository

	Referral         ReferralRepository
	ReferralSettings ReferralSettingsRepository
	Wallet           WalletRepository
	Payment          PaymentRepository
	Coupon           CouponRepository
	CouponUsage      CouponUsageRepository

	Address        AddressRepository
	ProductVariant ProductVariantRepository
	Cart           CartRepository
}
