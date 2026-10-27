package interfaces

import "ecommerce/internal/domain/models"

type InvoiceGenerator interface {
	Generate(order *models.Order) ([]byte, error)
}
