package invoice

import (
	"bytes"
	"ecommerce/internal/domain/models"
	"fmt"
	"time"

	"github.com/jung-kurt/gofpdf"
)

type PDFGenerator struct {
	storeName string
}

func NewPDFGenerator(storeName string) *PDFGenerator {
	return &PDFGenerator{
		storeName: storeName,
	}
}
func (g *PDFGenerator) Generate(order *models.Order) ([]byte, error) {
	if order == nil {
		return nil, fmt.Errorf("cannot generate invoice for nil order")
	}
	pdf := gofpdf.New("P", "mm", "A4", "")
	//p-portrait
	//mm-millimeter
	//a4-size
	//""-custom font-directory
	pdf.SetTitle("Order Invoice", false)
	pdf.SetMargins(15, 15, 15)
	pdf.SetAutoPageBreak(true, 15)
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 20)
	pdf.CellFormat(
		180,         // Width in mm
		12,          // Height in mm
		g.storeName, // Text to print
		"",          // Border
		1,           // Move to next line afterwards
		"C",         // Horizontal alignment: center
		false,       // No background fill
		0,           // No hyperlink
		"",          // No hyperlink target
	)
	pdf.SetFont("Arial", "B", 16)
	pdf.CellFormat(180, 10, "INVOICE", "", 1, "C", false, 0, "")
	pdf.Ln(5)
	pdf.SetFont("Arial", "", 11)
	pdf.CellFormat(
		90, 7,
		"Order Number: "+order.OrderNumber,
		"", 0, "L", false, 0, "",
	)
	pdf.CellFormat(
		90, 7,
		"Invoice Date: "+invoiceDate(order),
		"", 1, "R", false, 0, "",
	)
	pdf.Ln(5)
	pdf.SetFont("Arial", "B", 12)
	pdf.CellFormat(180, 8, "Billing / Delivery Address", "", 1, "L", false, 0, "")
	pdf.SetFont("Arial", "", 10)
	if order.Address != nil {
		address := order.Address
		pdf.CellFormat(180, 6, address.Name, "", 1, "L", false, 0, "")
		pdf.CellFormat(180, 6, address.Phone, "", 1, "L", false, 0, "")
		pdf.CellFormat(180, 6, address.AddressLine1, "", 1, "L", false, 0, "")
		if address.AddressLine2 != nil && *address.AddressLine2 != "" {
			pdf.CellFormat(180, 6, *address.AddressLine2, "", 1, "L", false, 0, "")
		}
		pdf.CellFormat(
			180, 6,
			fmt.Sprintf(
				"%s, %s - %s",
				address.City,
				address.State,
				address.PostalCode,
			),
			"", 1, "L", false, 0, "",
		)
	}
	pdf.Ln(6)
	pdf.SetFont("Arial", "B", 10)
	pdf.SetFillColor(230, 230, 230)
	pdf.CellFormat(65, 9, "Product", "1", 0, "L", true, 0, "")
	pdf.CellFormat(25, 9, "Unit Price", "1", 0, "R", true, 0, "")
	pdf.CellFormat(20, 9, "Quantity", "1", 0, "C", true, 0, "")
	pdf.CellFormat(30, 9, "Subtotal", "1", 1, "R", true, 0, "")
	pdf.SetFont("Arial", "", 9)
	for _, item := range order.Items {
		productName := item.ProductName
		if productName == "" {
			productName = "Product"
		}
		pdf.CellFormat(65, 8, productName, "1", 0, "L", false, 0, "")
		pdf.CellFormat(25, 8, money(item.UnitPrice), "1", 0, "R", false, 0, "")
		pdf.CellFormat(20, 8, fmt.Sprintf("%d", item.Quantity), "1", 0, "C", false, 0, "")
		pdf.CellFormat(30, 8, money(item.Subtotal), "1", 1, "R", false, 0, "")
	}
	pdf.Ln(6)
	pdf.SetFont("Arial", "", 11)
	pdf.CellFormat(140, 7, "Subtotal", "", 0, "R", false, 0, "")
	pdf.CellFormat(40, 7, money(order.Subtotal), "", 1, "R", false, 0, "")
	pdf.CellFormat(140, 7, "Discount", "", 0, "R", false, 0, "")
	pdf.CellFormat(40, 7, "-"+money(order.DiscountAmount), "", 1, "R", false, 0, "")
	pdf.CellFormat(140, 7, "Shipping", "", 0, "R", false, 0, "")
	pdf.CellFormat(40, 7, money(order.ShippingAmount), "", 1, "R", false, 0, "")
	pdf.SetFont("Arial", "B", 12)
	pdf.CellFormat(140, 9, "Total", "", 0, "R", false, 0, "")
	pdf.CellFormat(40, 9, money(order.TotalAmount), "", 1, "R", false, 0, "")
	pdf.Ln(12)
	pdf.SetFont("Arial", "I", 10)
	pdf.CellFormat(
		180, 7,
		"Thank you for shopping with us!",
		"", 1, "C", false, 0, "",
	)
	var output bytes.Buffer
	if err := pdf.Output(&output); err != nil {
		return nil, fmt.Errorf("generate invoice PDF: %w", err)
	}
	return output.Bytes(), nil
}
func invoiceDate(order *models.Order) string {
	if !order.PlacedAt.IsZero() {
		return order.PlacedAt.Format("02 Jan 2006")
	}
	if !order.CreatedAt.IsZero() {
		return order.CreatedAt.Format("02 Jan 2006")
	}
	return time.Time{}.Format("02 Jan 2006")
}
func money(amount float64) string {
	return fmt.Sprintf("Rs. %.2f", amount)
}
