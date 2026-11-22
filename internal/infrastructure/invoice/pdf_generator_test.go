package invoice

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"ecommerce/internal/domain/models"

	"github.com/jung-kurt/gofpdf"
)

func TestPDFGeneratorGeneratesDetailedInvoice(t *testing.T) {
	addressLine2 := "Apartment 4B"
	orderNumber := strings.Repeat("DRP-2026-001-", 5)
	order := &models.Order{
		OrderNumber:    orderNumber,
		Subtotal:       125000,
		DiscountAmount: 5000,
		ShippingAmount: 150,
		TotalAmount:    120150,
		PaymentStatus:  "paid",
		PlacedAt:       time.Date(2026, time.November, 22, 12, 0, 0, 0, time.UTC),
		Address: &models.OrderAddress{
			Name:         "Niyas Mohammed",
			Phone:        "9876543210",
			AddressLine1: "12 Market Road",
			AddressLine2: &addressLine2,
			City:         "Kochi",
			State:        "Kerala",
			PostalCode:   "682001",
			Country:      "India",
		},
		Payments: []models.Payment{{PaymentMethod: "razorpay"}},
		Items: []models.OrderItem{{
			ProductName: "Premium Cotton Shirt",
			VariantInfo: json.RawMessage(`{"sku":"SHIRT-BLU-M","size":"M","color":"Blue"}`),
			UnitPrice:   2500,
			Quantity:    2,
			Subtotal:    5000,
		}},
	}

	pdfBytes, err := NewPDFGenerator("Drape").Generate(order)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if !bytes.HasPrefix(pdfBytes, []byte("%PDF-")) {
		t.Fatalf("Generate() did not return a PDF document")
	}
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	description := itemDescription(
		pdf,
		order.Items[0],
		pdf.UnicodeTranslatorFromDescriptor(""),
	)
	for _, expected := range []string{"Premium Cotton Shirt", "SHIRT-BLU-M", "Size M", "Color Blue"} {
		if !strings.Contains(description, expected) {
			t.Errorf("item description %q does not contain %q", description, expected)
		}
	}

	pdf.SetFont("Arial", "B", 8.5)
	wrappedOrderNumber := wrapTextByWidth(pdf, orderNumber, 77)
	if !strings.Contains(wrappedOrderNumber, "\n") {
		t.Fatal("expected a long order number to wrap onto multiple lines")
	}
	for _, line := range strings.Split(wrappedOrderNumber, "\n") {
		if pdf.GetStringWidth(line) > 77 {
			t.Errorf("wrapped order-number line exceeds its panel width: %q", line)
		}
	}
}

func TestPDFGeneratorPaginatesLongOrders(t *testing.T) {
	items := make([]models.OrderItem, 40)
	for i := range items {
		items[i] = models.OrderItem{
			ProductName: "Premium Cotton Shirt",
			UnitPrice:   100,
			Quantity:    1,
			Subtotal:    100,
		}
	}
	pdfBytes, err := NewPDFGenerator("").Generate(&models.Order{
		OrderNumber: "DRP-MULTI-PAGE",
		Items:       items,
	})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if bytes.Count(pdfBytes, []byte("/Type /Page\n")) < 2 {
		t.Error("expected a long invoice to be split across multiple pages")
	}
}

func TestMoneyUsesIndianDigitGrouping(t *testing.T) {
	if got := money(1234567.89); got != "Rs. 12,34,567.89" {
		t.Errorf("money() = %q, want Rs. 12,34,567.89", got)
	}
	if got := money(-1250); got != "Rs. -1,250.00" {
		t.Errorf("money() = %q, want Rs. -1,250.00", got)
	}
}

func TestPDFGeneratorRejectsNilOrder(t *testing.T) {
	if _, err := NewPDFGenerator("Drape").Generate(nil); err == nil {
		t.Fatal("Generate(nil) expected an error")
	}
}
