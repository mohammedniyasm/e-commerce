package reportexport

import (
	"bytes"
	"fmt"

	"ecommerce/internal/domain/models"

	"github.com/jung-kurt/gofpdf"
	"github.com/xuri/excelize/v2"
)

func GenerateSalesPDF(report *models.SalesReport) ([]byte, error) {
	if report == nil {
		return nil, fmt.Errorf("sales report is nil")
	}
	pdf := newReportPDF("SALES REPORT")
	pdf.SetFont("Arial", "", 11)
	pdf.CellFormat(180, 8, fmt.Sprintf(
		"Total Sales: Rs. %.2f", report.Summary.TotalSales,
	), "", 1, "L", false, 0, "")
	pdf.CellFormat(180, 8, fmt.Sprintf(
		"Total Orders: %d", report.Summary.TotalOrders,
	), "", 1, "L", false, 0, "")
	pdf.CellFormat(180, 8, fmt.Sprintf(
		"Units Sold: %d", report.Summary.TotalUnitsSold,
	), "", 1, "L", false, 0, "")
	pdf.CellFormat(180, 8, fmt.Sprintf(
		"Total Discount: Rs. %.2f", report.Summary.TotalDiscount,
	), "", 1, "L", false, 0, "")
	pdf.Ln(8)
	addTableHeader(pdf, []string{
		"Date", "Sales", "Orders", "Units Sold",
	}, []float64{50, 50, 35, 45})
	pdf.SetFont("Arial", "", 9)
	for _, row := range report.Breakdown {
		pdf.CellFormat(50, 8, row.Date.Format("02 Jan 2006"), "1", 0, "L", false, 0, "")
		pdf.CellFormat(50, 8, fmt.Sprintf("Rs. %.2f", row.Sales), "1", 0, "R", false, 0, "")
		pdf.CellFormat(35, 8, fmt.Sprintf("%d", row.Orders), "1", 0, "C", false, 0, "")
		pdf.CellFormat(45, 8, fmt.Sprintf("%d", row.UnitsSold), "1", 1, "C", false, 0, "")
	}
	return pdfBytes(pdf)
}
func GenerateOrdersPDF(report *models.OrdersReport) ([]byte, error) {
	if report == nil {
		return nil, fmt.Errorf("orders report is nil")
	}
	pdf := newReportPDF("ORDERS REPORT")
	summary := report.Summary
	pdf.SetFont("Arial", "", 11)
	summaryRows := []struct {
		label string
		value int64
	}{
		{"Total Orders", summary.TotalOrders},
		{"Pending Orders", summary.PendingOrders},
		{"Confirmed Orders", summary.ConfirmedOrders},
		{"Processing Orders", summary.ProcessingOrders},
		{"Shipped Orders", summary.ShippedOrders},
		{"Delivered Orders", summary.DeliveredOrders},
		{"Cancelled Orders", summary.CancelledOrders},
		{"Returned Orders", summary.ReturnedOrders},
	}
	for _, row := range summaryRows {
		pdf.CellFormat(
			180, 8,
			fmt.Sprintf("%s: %d", row.label, row.value),
			"", 1, "L", false, 0, "",
		)
	}
	pdf.Ln(8)
	addTableHeader(pdf, []string{
		"Date", "Total Orders",
	}, []float64{90, 90})
	pdf.SetFont("Arial", "", 10)
	for _, row := range report.Breakdown {
		pdf.CellFormat(90, 8, row.Date.Format("02 Jan 2006"), "1", 0, "L", false, 0, "")
		pdf.CellFormat(90, 8, fmt.Sprintf("%d", row.TotalOrders), "1", 1, "C", false, 0, "")
	}
	return pdfBytes(pdf)
}
func GenerateSalesExcel(report *models.SalesReport) ([]byte, error) {
	if report == nil {
		return nil, fmt.Errorf("sales report is nil")
	}
	summary := [][]any{
		{"Total Sales", report.Summary.TotalSales},
		{"Total Orders", report.Summary.TotalOrders},
		{"Total Units Sold", report.Summary.TotalUnitsSold},
		{"Total Discount", report.Summary.TotalDiscount},
	}
	headers := []string{
		"Date", "Sales", "Orders", "Units Sold",
	}
	var rows [][]any
	for _, row := range report.Breakdown {
		rows = append(rows, []any{
			row.Date,
			row.Sales,
			row.Orders,
			row.UnitsSold,
		})
	}
	return generateExcel("Sales Report", summary, headers, rows)
}
func GenerateOrdersExcel(report *models.OrdersReport) ([]byte, error) {
	if report == nil {
		return nil, fmt.Errorf("orders report is nil")
	}
	s := report.Summary
	summary := [][]any{
		{"Total Orders", s.TotalOrders},
		{"Pending Orders", s.PendingOrders},
		{"Confirmed Orders", s.ConfirmedOrders},
		{"Processing Orders", s.ProcessingOrders},
		{"Shipped Orders", s.ShippedOrders},
		{"Delivered Orders", s.DeliveredOrders},
		{"Cancelled Orders", s.CancelledOrders},
		{"Returned Orders", s.ReturnedOrders},
	}
	headers := []string{"Date", "Total Orders"}
	var rows [][]any
	for _, row := range report.Breakdown {
		rows = append(rows, []any{
			row.Date,
			row.TotalOrders,
		})
	}
	return generateExcel("Orders Report", summary, headers, rows)
}
func GenerateProductsExcel(report *models.ProductsReport) ([]byte, error) {
	if report == nil {
		return nil, fmt.Errorf("products report is nil")
	}
	s := report.Summary
	summary := [][]any{
		{"Total Products Sold", s.TotalProductsSold},
		{"Total Units Sold", s.TotalUnitsSold},
		{"Total Revenue", s.TotalRevenue},
	}
	headers := []string{
		"Product ID",
		"Product Name",
		"Units Sold",
		"Total Orders",
		"Revenue",
	}
	var rows [][]any
	for _, row := range report.Breakdown {
		rows = append(rows, []any{
			row.ProductID,
			row.ProductName,
			row.UnitsSold,
			row.TotalOrders,
			row.Revenue,
		})
	}
	return generateExcel("Products Report", summary, headers, rows)
}
func newReportPDF(title string) *gofpdf.Fpdf {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetTitle(title, false)
	pdf.SetMargins(15, 15, 15)
	pdf.SetAutoPageBreak(true, 15)
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 18)
	pdf.CellFormat(180, 12, title, "", 1, "C", false, 0, "")
	pdf.Ln(8)
	return pdf
}
func addTableHeader(pdf *gofpdf.Fpdf, headers []string, widths []float64) {
	pdf.SetFont("Arial", "B", 10)
	pdf.SetFillColor(230, 230, 230)
	for i, header := range headers {
		pdf.CellFormat(
			widths[i], 9, header,
			"1", 0, "C", true, 0, "",
		)
	}
	pdf.Ln(-1)
}
func pdfBytes(pdf *gofpdf.Fpdf) ([]byte, error) {
	var output bytes.Buffer
	if err := pdf.Output(&output); err != nil {
		return nil, fmt.Errorf("generate PDF: %w", err)
	}
	return output.Bytes(), nil
}
func generateExcel(sheetName string, summary [][]any, headers []string, rows [][]any) ([]byte, error) {
	file := excelize.NewFile()
	defer func() {
		_ = file.Close()
	}()
	const sheet = "Report"
	if err := file.SetSheetName("Sheet1", sheet); err != nil {
		return nil, err
	}
	file.SetCellValue(sheet, "A1", sheetName)
	file.SetCellValue(sheet, "A3", "Summary")
	for i, row := range summary {
		file.SetCellValue(sheet, fmt.Sprintf("A%d", i+4), row[0])
		file.SetCellValue(sheet, fmt.Sprintf("B%d", i+4), row[1])
	}
	headerRow := len(summary) + 6
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, headerRow)
		file.SetCellValue(sheet, cell, header)
	}
	for rowIndex, row := range rows {
		for colIndex, value := range row {
			cell, _ := excelize.CoordinatesToCellName(
				colIndex+1,
				headerRow+rowIndex+1,
			)
			file.SetCellValue(sheet, cell, value)
		}
	}
	file.SetColWidth(sheet, "A", "E", 20)
	output, err := file.WriteToBuffer()
	if err != nil {
		return nil, fmt.Errorf("generate Excel: %w", err)
	}
	return output.Bytes(), nil
}
