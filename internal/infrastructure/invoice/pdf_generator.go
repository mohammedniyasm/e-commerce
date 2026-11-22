package invoice

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"ecommerce/internal/domain/models"

	"github.com/jung-kurt/gofpdf"
)

const (
	pageWidth       = 210.0
	pageMargin      = 15.0
	contentWidth    = pageWidth - 2*pageMargin
	pageBottomLimit = 270.0
)

var (
	inkColor       = [3]int{24, 31, 42}
	mutedColor     = [3]int{105, 114, 126}
	accentColor    = [3]int{31, 31, 31}
	borderColor    = [3]int{226, 229, 233}
	panelColor     = [3]int{247, 248, 250}
	tableHeader    = [3]int{30, 38, 50}
	totalPanelFill = [3]int{242, 243, 245}
)

type variantSnapshot struct {
	SKU   string  `json:"sku"`
	Size  *string `json:"size"`
	Color *string `json:"color"`
}

type PDFGenerator struct {
	storeName string
}

func NewPDFGenerator(storeName string) *PDFGenerator {
	return &PDFGenerator{storeName: storeName}
}

func (g *PDFGenerator) Generate(order *models.Order) ([]byte, error) {
	if order == nil {
		return nil, fmt.Errorf("cannot generate invoice for nil order")
	}

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetTitle("Invoice "+order.OrderNumber, false)
	pdf.SetAuthor(invoiceStoreName(g.storeName), false)
	pdf.SetMargins(pageMargin, pageMargin, pageMargin)
	pdf.SetAutoPageBreak(true, 18)
	pdf.AddPage()

	translate := pdf.UnicodeTranslatorFromDescriptor("")
	storeName := translate(invoiceStoreName(g.storeName))

	drawInvoiceHeader(pdf, storeName, order, translate, false)
	tableY := drawCustomerAndOrderPanels(pdf, order, translate)
	drawItemsHeader(pdf, tableY)
	y := tableY + 9

	for index, item := range order.Items {
		productDescription := itemDescription(pdf, item, translate)
		productWidth := 87.0
		lines := wrappedLineCount(pdf, productDescription, productWidth-8)
		rowHeight := maxFloat(10, float64(lines)*4.5+4)

		if y+rowHeight > pageBottomLimit {
			pdf.AddPage()
			drawInvoiceHeader(pdf, storeName, order, translate, true)
			drawItemsHeader(pdf, 47)
			y = 56
		}

		if index%2 == 1 {
			pdf.SetFillColor(250, 250, 251)
			pdf.Rect(pageMargin, y, contentWidth, rowHeight, "F")
		}
		pdf.SetDrawColor(borderColor[0], borderColor[1], borderColor[2])
		pdf.Line(pageMargin, y+rowHeight, pageMargin+contentWidth, y+rowHeight)

		pdf.SetFont("Arial", "", 9)
		pdf.SetTextColor(inkColor[0], inkColor[1], inkColor[2])
		pdf.SetXY(pageMargin+4, y+2)
		pdf.MultiCell(productWidth-8, 4.5, productDescription, "", "L", false)

		pdf.SetFont("Arial", "", 9)
		pdf.SetTextColor(mutedColor[0], mutedColor[1], mutedColor[2])
		pdf.SetXY(pageMargin+productWidth, y)
		pdf.CellFormat(28, rowHeight, money(item.UnitPrice), "", 0, "R", false, 0, "")
		pdf.CellFormat(20, rowHeight, fmt.Sprintf("%d", item.Quantity), "", 0, "C", false, 0, "")
		pdf.SetTextColor(inkColor[0], inkColor[1], inkColor[2])
		pdf.SetFont("Arial", "B", 9)
		pdf.CellFormat(45, rowHeight, money(item.Subtotal), "", 0, "R", false, 0, "")
		y += rowHeight
	}

	if y+64 > pageBottomLimit {
		pdf.AddPage()
		drawInvoiceHeader(pdf, storeName, order, translate, true)
		y = 42
	}
	drawTotals(pdf, y+8, order)
	drawInvoiceFooter(pdf, g.storeName)

	var output bytes.Buffer
	if err := pdf.Output(&output); err != nil {
		return nil, fmt.Errorf("generate invoice PDF: %w", err)
	}
	return output.Bytes(), nil
}

func drawInvoiceHeader(pdf *gofpdf.Fpdf, storeName string, order *models.Order, translate func(string) string, continued bool) {
	pdf.SetDrawColor(accentColor[0], accentColor[1], accentColor[2])
	pdf.SetLineWidth(1.1)
	pdf.Line(pageMargin, 14, pageMargin+contentWidth, 14)
	pdf.SetLineWidth(0.2)

	pdf.SetXY(pageMargin, 20)
	pdf.SetTextColor(inkColor[0], inkColor[1], inkColor[2])
	pdf.SetFont("Arial", "B", 21)
	pdf.CellFormat(105, 11, storeName, "", 0, "L", false, 0, "")

	pdf.SetTextColor(accentColor[0], accentColor[1], accentColor[2])
	pdf.SetFont("Arial", "B", 19)
	title := "INVOICE"
	if continued {
		title = "INVOICE / CONTINUED"
	}
	pdf.CellFormat(75, 11, title, "", 1, "R", false, 0, "")

	pdf.SetXY(pageMargin, 34)
	pdf.SetTextColor(mutedColor[0], mutedColor[1], mutedColor[2])
	pdf.SetFont("Arial", "", 9)
	pdf.CellFormat(90, 6, "ORDER  "+translate(nonEmpty(order.OrderNumber, "-")), "", 0, "L", false, 0, "")
	pdf.CellFormat(90, 6, "ISSUED  "+translate(invoiceDate(order)), "", 1, "R", false, 0, "")

	pdf.SetDrawColor(borderColor[0], borderColor[1], borderColor[2])
	pdf.Line(pageMargin, 43, pageMargin+contentWidth, 43)
}

func drawCustomerAndOrderPanels(pdf *gofpdf.Fpdf, order *models.Order, translate func(string) string) float64 {
	const panelY = 50.0
	const panelHeight = 52.0
	const panelGap = 6.0
	panelWidth := (contentWidth - panelGap) / 2
	rightX := pageMargin + panelWidth + panelGap

	drawPanel(pdf, pageMargin, panelY, panelWidth, panelHeight)
	drawPanel(pdf, rightX, panelY, panelWidth, panelHeight)

	pdf.SetTextColor(accentColor[0], accentColor[1], accentColor[2])
	pdf.SetFont("Arial", "B", 8)
	pdf.SetXY(pageMargin+5, panelY+4)
	pdf.CellFormat(panelWidth-10, 5, "DELIVER TO", "", 1, "L", false, 0, "")
	pdf.SetXY(rightX+5, panelY+4)
	pdf.CellFormat(panelWidth-10, 5, "ORDER DETAILS", "", 1, "L", false, 0, "")

	pdf.SetTextColor(inkColor[0], inkColor[1], inkColor[2])
	pdf.SetFont("Arial", "", 8.5)
	pdf.SetXY(pageMargin+5, panelY+11)
	if address := order.Address; address != nil {
		addressLines := []string{
			address.Name,
			address.AddressLine1,
		}
		if address.AddressLine2 != nil {
			addressLines = append(addressLines, *address.AddressLine2)
		}
		addressLines = append(addressLines, fmt.Sprintf("%s, %s %s", address.City, address.State, address.PostalCode))
		if address.Country != "" {
			addressLines = append(addressLines, address.Country)
		}
		if address.Phone != "" {
			addressLines = append(addressLines, "Phone: "+address.Phone)
		}
		pdf.MultiCell(panelWidth-10, 4.4, translate(strings.Join(nonEmptyLines(addressLines), "\n")), "", "L", false)
	} else {
		pdf.CellFormat(panelWidth-10, 5, "Delivery address unavailable", "", 1, "L", false, 0, "")
	}

	detailsX := rightX + 5
	detailsY := panelY + 11
	pdf.SetXY(detailsX, detailsY)
	pdf.SetFont("Arial", "B", 8.5)
	pdf.CellFormat(panelWidth-10, 4, "Order number", "", 1, "L", false, 0, "")
	orderNumber := translate(nonEmpty(order.OrderNumber, "-"))
	pdf.SetFont("Arial", "B", 8.5)
	wrappedOrderNumber := wrapTextByWidth(pdf, orderNumber, panelWidth-10)
	orderNumberLines := strings.Count(wrappedOrderNumber, "\n") + 1
	pdf.SetXY(detailsX, detailsY+4)
	pdf.MultiCell(panelWidth-10, 4.2, wrappedOrderNumber, "", "L", false)
	detailsY += 5 + float64(orderNumberLines)*4.2

	pdf.SetXY(detailsX, detailsY)
	pdf.SetFont("Arial", "B", 8)
	pdf.CellFormat(28, 4.5, "Payment", "", 0, "L", false, 0, "")
	pdf.SetFont("Arial", "", 8)
	pdf.CellFormat(panelWidth-38, 4.5, translate(paymentMethod(order)), "", 1, "R", false, 0, "")

	pdf.SetX(detailsX)
	pdf.SetFont("Arial", "B", 8)
	pdf.CellFormat(28, 4.5, "Status", "", 0, "L", false, 0, "")
	pdf.SetFont("Arial", "B", 8.5)
	pdf.SetTextColor(statusColor(order.PaymentStatus))
	pdf.CellFormat(panelWidth-38, 4.5, strings.ToUpper(translate(nonEmpty(order.PaymentStatus, "unknown"))), "", 1, "R", false, 0, "")
	pdf.SetTextColor(inkColor[0], inkColor[1], inkColor[2])

	pdf.SetX(detailsX)
	pdf.SetFont("Arial", "B", 8)
	pdf.CellFormat(28, 4.5, "Order date", "", 0, "L", false, 0, "")
	pdf.SetFont("Arial", "", 8)
	pdf.CellFormat(panelWidth-38, 4.5, translate(invoiceDate(order)), "", 1, "R", false, 0, "")

	return panelY + panelHeight + 8
}

func drawPanel(pdf *gofpdf.Fpdf, x, y, width, height float64) {
	pdf.SetFillColor(panelColor[0], panelColor[1], panelColor[2])
	pdf.SetDrawColor(borderColor[0], borderColor[1], borderColor[2])
	pdf.RoundedRect(x, y, width, height, 2, "1234", "FD")
}

func drawItemsHeader(pdf *gofpdf.Fpdf, y float64) {
	pdf.SetFillColor(tableHeader[0], tableHeader[1], tableHeader[2])
	pdf.RoundedRect(pageMargin, y, contentWidth, 9, 1.5, "1234", "F")
	pdf.SetXY(pageMargin+4, y)
	pdf.SetFont("Arial", "B", 8)
	pdf.SetTextColor(255, 255, 255)
	pdf.CellFormat(83, 9, "ITEM", "", 0, "L", false, 0, "")
	pdf.CellFormat(28, 9, "UNIT PRICE", "", 0, "R", false, 0, "")
	pdf.CellFormat(20, 9, "QTY", "", 0, "C", false, 0, "")
	pdf.CellFormat(45, 9, "AMOUNT", "", 0, "R", false, 0, "")
	pdf.SetTextColor(inkColor[0], inkColor[1], inkColor[2])
}

func drawTotals(pdf *gofpdf.Fpdf, y float64, order *models.Order) {
	x := pageMargin + 78
	width := contentWidth - 78
	pdf.SetDrawColor(borderColor[0], borderColor[1], borderColor[2])
	pdf.Line(x, y, pageMargin+contentWidth, y)

	pdf.SetXY(x, y+5)
	pdf.SetFont("Arial", "", 9)
	pdf.SetTextColor(mutedColor[0], mutedColor[1], mutedColor[2])
	pdf.CellFormat(width-30, 6, "Subtotal", "", 0, "R", false, 0, "")
	pdf.CellFormat(30, 6, money(order.Subtotal), "", 1, "R", false, 0, "")

	pdf.SetX(x)
	pdf.CellFormat(width-30, 6, "Discount", "", 0, "R", false, 0, "")
	pdf.CellFormat(30, 6, "-"+money(order.DiscountAmount), "", 1, "R", false, 0, "")

	pdf.SetX(x)
	pdf.CellFormat(width-30, 6, "Shipping", "", 0, "R", false, 0, "")
	pdf.CellFormat(30, 6, money(order.ShippingAmount), "", 1, "R", false, 0, "")

	pdf.SetFillColor(totalPanelFill[0], totalPanelFill[1], totalPanelFill[2])
	pdf.RoundedRect(x, y+26, width, 13, 1.5, "1234", "F")
	pdf.SetXY(x+4, y+28)
	pdf.SetTextColor(inkColor[0], inkColor[1], inkColor[2])
	pdf.SetFont("Arial", "B", 10)
	pdf.CellFormat(width-38, 8, "TOTAL", "", 0, "L", false, 0, "")
	pdf.SetTextColor(accentColor[0], accentColor[1], accentColor[2])
	pdf.SetFont("Arial", "B", 12)
	pdf.CellFormat(34, 8, money(order.TotalAmount), "", 0, "R", false, 0, "")
}

func drawInvoiceFooter(pdf *gofpdf.Fpdf, storeName string) {
	y := pdf.GetY()
	if y < 250 {
		y = 250
	}
	pdf.SetDrawColor(borderColor[0], borderColor[1], borderColor[2])
	pdf.Line(pageMargin, y, pageMargin+contentWidth, y)
	pdf.SetXY(pageMargin, y+4)
	pdf.SetFont("Arial", "B", 8)
	pdf.SetTextColor(inkColor[0], inkColor[1], inkColor[2])
	pdf.CellFormat(contentWidth, 5, "Thank you for choosing "+pdf.UnicodeTranslatorFromDescriptor("")(invoiceStoreName(storeName))+".", "", 1, "C", false, 0, "")
	pdf.SetFont("Arial", "", 7)
	pdf.SetTextColor(mutedColor[0], mutedColor[1], mutedColor[2])
	pdf.CellFormat(contentWidth, 4, "This document is your order invoice. Please retain it for your records.", "", 0, "C", false, 0, "")
}

func itemDescription(pdf *gofpdf.Fpdf, item models.OrderItem, translate func(string) string) string {
	productName := nonEmpty(item.ProductName, "Product")
	variant := variantSnapshot{SKU: item.Variant.SKU}
	if len(item.VariantInfo) > 0 {
		_ = json.Unmarshal(item.VariantInfo, &variant)
	}
	details := make([]string, 0, 3)
	if variant.SKU != "" {
		details = append(details, "SKU "+variant.SKU)
	}
	if variant.Size != nil && *variant.Size != "" {
		details = append(details, "Size "+*variant.Size)
	}
	if variant.Color != nil && *variant.Color != "" {
		details = append(details, "Color "+*variant.Color)
	}
	description := translate(productName)
	if len(details) > 0 {
		description += "\n" + translate(strings.Join(details, " / "))
	}
	return description
}

func wrappedLineCount(pdf *gofpdf.Fpdf, text string, width float64) int {
	pdf.SetFont("Arial", "", 9)
	count := 0
	for _, line := range strings.Split(text, "\n") {
		if line == "" {
			count++
			continue
		}
		count += len(pdf.SplitLines([]byte(line), width))
	}
	if count == 0 {
		return 1
	}
	return count
}

func wrapTextByWidth(pdf *gofpdf.Fpdf, text string, width float64) string {
	lines := make([]string, 0, 1)
	var line strings.Builder
	for _, character := range text {
		candidate := line.String() + string(character)
		if line.Len() > 0 && pdf.GetStringWidth(candidate) > width {
			lines = append(lines, line.String())
			line.Reset()
		}
		line.WriteRune(character)
	}
	if line.Len() > 0 {
		lines = append(lines, line.String())
	}
	return strings.Join(lines, "\n")
}

func paymentMethod(order *models.Order) string {
	if len(order.Payments) > 0 {
		return nonEmpty(order.Payments[0].PaymentMethod, "Not specified")
	}
	return "Not specified"
}

func statusColor(status string) (int, int, int) {
	switch strings.ToLower(status) {
	case "paid":
		return 31, 31, 31
	case "failed", "refunded", "partially_refunded":
		return 92, 92, 92
	default:
		return 64, 64, 64
	}
}

func invoiceStoreName(name string) string {
	return nonEmpty(name, "DRAPE")
}

func invoiceDate(order *models.Order) string {
	if !order.PlacedAt.IsZero() {
		return order.PlacedAt.Format("02 Jan 2006")
	}
	if !order.CreatedAt.IsZero() {
		return order.CreatedAt.Format("02 Jan 2006")
	}
	return "-"
}

func money(amount float64) string {
	formatted := fmt.Sprintf("%.2f", amount)
	sign := ""
	if strings.HasPrefix(formatted, "-") {
		sign = "-"
		formatted = strings.TrimPrefix(formatted, "-")
	}
	parts := strings.SplitN(formatted, ".", 2)
	integer := parts[0]
	if len(integer) > 3 {
		prefix := integer[:len(integer)-3]
		grouped := integer[len(integer)-3:]
		groups := make([]string, 0, len(prefix)/2+1)
		for len(prefix) > 2 {
			groups = append([]string{prefix[len(prefix)-2:]}, groups...)
			prefix = prefix[:len(prefix)-2]
		}
		if prefix != "" {
			groups = append([]string{prefix}, groups...)
		}
		integer = strings.Join(groups, ",") + "," + grouped
	}
	return "Rs. " + sign + integer + "." + parts[1]
}

func nonEmpty(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return strings.TrimSpace(value)
}

func nonEmptyLines(lines []string) []string {
	result := make([]string, 0, len(lines))
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			result = append(result, strings.TrimSpace(line))
		}
	}
	return result
}

func maxFloat(left, right float64) float64 {
	if left > right {
		return left
	}
	return right
}
