package handler

import (
	"encoding/json"
	"testing"

	"ecommerce/internal/domain/models"
)

func TestMapOrderToResponseIncludesCODPaymentMethod(t *testing.T) {
	addressLine2 := "Apartment 4B"
	order := &models.Order{
		User: &models.User{
			ID:    3,
			Name:  "Alex Customer",
			Email: "alex@example.test",
			Phone: "9876543210",
		},
		Items: []models.OrderItem{
			{
				ID:          2,
				VariantID:   5,
				ProductName: "Everyday Shirt",
				ImageURL:    "https://example.test/shirt.jpg",
				VariantInfo: json.RawMessage(`{"sku":"SHIRT-M","size":"M","color":"Blue"}`),
				Variant:     models.ProductVariant{ProductID: 12},
			},
		},
		Address: &models.OrderAddress{
			Name:         "Alex Customer",
			Phone:        "9876543210",
			AddressLine1: "12 Example Road",
			AddressLine2: &addressLine2,
			City:         "Kochi",
			State:        "Kerala",
			PostalCode:   "682001",
			Country:      "India",
		},
		Payments: []models.Payment{
			{PaymentMethod: "razorpay"},
			{PaymentMethod: "cash_on_delivery"},
		},
	}

	got := mapOrderToResponse(order)
	if got.PaymentMethod != "cash_on_delivery" {
		t.Fatalf("expected COD payment method in response, got %q", got.PaymentMethod)
	}
	if len(got.Items) != 1 || got.Items[0].ProductID != 12 {
		t.Fatalf("expected order item product ID 12 in response, got %+v", got.Items)
	}
	item := got.Items[0]
	if item.ProductName != "Everyday Shirt" || item.ImageURL != "https://example.test/shirt.jpg" ||
		string(item.VariantInfo) != `{"sku":"SHIRT-M","size":"M","color":"Blue"}` {
		t.Fatalf("expected product image and variant snapshot in response, got %+v", item)
	}
	if got.Address == nil ||
		got.Address.AddressLine1 != "12 Example Road" ||
		got.Address.AddressLine2 != addressLine2 ||
		got.Address.Pincode != "682001" ||
		got.Address.Country != "India" {
		t.Fatalf("expected complete shipping address in response, got %+v", got.Address)
	}
	if got.User == nil || got.User.Email != "alex@example.test" {
		t.Fatalf("expected customer email in order response, got %+v", got.User)
	}
}
